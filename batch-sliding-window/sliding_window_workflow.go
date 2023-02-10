package batch_sliding_window

import (
	"fmt"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"time"
)

type SlidingWindowWorkflowInput struct {
	PageSize          int
	SlidingWindowSize int
	Offset            int // inclusive
	MaximumOffset     int // exclusive
	Progress          int
	// The set of ids
	CurrentRecords map[int]bool // recordId -> ignored boolean
}

type SingleRecordOrError struct {
	Record SingleRecord
	Err    error
}

type SlidingWindow struct {
	input                    SlidingWindowWorkflowInput
	currentRecords           map[int]bool
	childrenStartedByThisRun []workflow.ChildWorkflowFuture
	offset                   int
	progress                 int
}

func SlidingWindowWorkflow(ctx workflow.Context, input SlidingWindowWorkflowInput) (recordCount int, err error) {
	impl := &SlidingWindow{
		input:          input,
		currentRecords: input.CurrentRecords,
		offset:         input.Offset,
		progress:       input.Progress,
	}
	if impl.currentRecords == nil {
		impl.currentRecords = make(map[int]bool)
	}

	return impl.Execute(ctx)
}

func (s *SlidingWindow) Execute(ctx workflow.Context) (recordCount int, err error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	})

	workflowId := workflow.GetInfo(ctx).WorkflowExecution.ID

	// Starts processing child workflow completion signals asynchronously
	s.completionSignalPump(ctx)

	// Starts loading records asynchronously
	recordsChannel := s.recordsPump(ctx)

	childrenToStart := s.input.SlidingWindowSize
	if s.input.Offset+childrenToStart > s.input.MaximumOffset {
		childrenToStart = s.input.MaximumOffset - s.input.Offset
	}
	// Process records
	for {
		// After starting SlidingWindowSize children blocks until a completion signal is received.
		workflow.Await(ctx, func() bool {
			return len(s.currentRecords) < childrenToStart
		})
		var record SingleRecordOrError
		more := recordsChannel.Receive(ctx, &record)
		// records pump failed
		if record.Err != nil {
			return 0, record.Err
		}
		// Completes or continues workflow, if no more records to process.
		if !more {
			return s.continueAsNewOrComplete(ctx)
		}
		options := workflow.ChildWorkflowOptions{
			ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
			WorkflowID:        fmt.Sprintf("%s/%d", workflowId, record.Record.Id),
		}
		childCtx := workflow.WithChildOptions(ctx, options)
		child := workflow.ExecuteChildWorkflow(childCtx, RecordProcessorWorkflow, record.Record)
		s.childrenStartedByThisRun = append(s.childrenStartedByThisRun, child)
		s.currentRecords[record.Record.Id] = true // value is ignored
	}
}

func (s *SlidingWindow) continueAsNewOrComplete(ctx workflow.Context) (int, error) {
	// Continues-as-new after starting PageSize children
	if s.offset < s.input.MaximumOffset {
		// Waits for all children to start. Without this wait, workflow completion through
		// continue-as-new might lead to a situation when they never start.
		for _, child := range s.childrenStartedByThisRun {
			err := child.GetChildWorkflowExecution().Get(ctx, nil)
			// Is not expected as children automatically generated
			// IDs are not expected to collide.
			if err != nil {
				return 0, err
			}
		}
		// Must drain signal channel without blocking before calling continue-as-new.
		// Failure to do so can lead to signal loss.
		s.drainCompletionSignalChannelAsync(ctx)

		// Returns ContinueAsNewError with new workflow input
		newInput := SlidingWindowWorkflowInput{
			PageSize:          s.input.PageSize,
			SlidingWindowSize: s.input.SlidingWindowSize,
			Offset:            s.input.Offset + len(s.childrenStartedByThisRun),
			MaximumOffset:     s.input.MaximumOffset,
			Progress:          s.progress,
			CurrentRecords:    s.currentRecords,
		}
		return 0, workflow.NewContinueAsNewError(ctx, SlidingWindowWorkflow, newInput)
	}
	// The last run in the chain.
	// Awaits for all children to complete
	workflow.Await(ctx, func() bool {
		return len(s.currentRecords) == 0
	})
	return s.progress, nil
}

func (s *SlidingWindow) drainCompletionSignalChannelAsync(ctx workflow.Context) {
	reportCompletionChannel := workflow.GetSignalChannel(ctx, "ReportCompletion")
	ok := false
	for {
		var recordId int
		ok = reportCompletionChannel.ReceiveAsync(&recordId)
		if ok {
			s.recordCompletion(recordId)
		} else {
			break
		}
	}
}

// completionSignalPump asynchronously processes ReportCompletion signals.
// There is no need to clean up the pump goroutine in case a workflow completes due to an error.
// All goroutines are released back automatically upon workflow completion.
func (s *SlidingWindow) completionSignalPump(ctx workflow.Context) {
	reportCompletionChannel := workflow.GetSignalChannel(ctx, "ReportCompletion")
	workflow.Go(ctx, func(ctx workflow.Context) {
		var recordId int
		_ = reportCompletionChannel.Receive(ctx, &recordId)
		s.recordCompletion(recordId)
	})
}

func (s *SlidingWindow) recordCompletion(recordId int) {
	// duplicate signal check
	if _, ok := s.currentRecords[recordId]; ok {
		delete(s.currentRecords, recordId)
		s.progress += 1
	}
}

// recordsPump pumps into the returned channel the batches of records loaded through GetRecords activity.
// A failure is reported as a last record with Err field set. The Channel is closed at the end of the records
// or on error.
func (s *SlidingWindow) recordsPump(ctx workflow.Context) workflow.Channel {
	log := workflow.GetLogger(ctx)
	recordsChannel := workflow.NewChannel(ctx)
	// Goroutine pumps records into the recordsChannel
	workflow.Go(ctx, func(ctx workflow.Context) {
		var loader *RecordLoader
		for {
			if s.offset >= s.input.MaximumOffset || s.offset >= s.input.Offset+s.input.SlidingWindowSize {
				break
			}
			getInput := &GetRecordsInput{
				PageSize:  s.input.PageSize,
				Offset:    s.offset,
				MaxOffset: s.input.MaximumOffset,
			}
			var getOutput GetRecordsOutput
			err := workflow.ExecuteActivity(ctx, loader.GetRecords, getInput).Get(ctx, &getOutput)
			if err != nil {
				recordsChannel.Send(ctx, SingleRecordOrError{Err: err})
				break
			}
			for _, record := range getOutput.Records {
				log.Info("Sending record: ", record.Id)
				recordsChannel.Send(ctx, SingleRecordOrError{Record: record})
				log.Info("Sent record: ", record.Id)
				s.offset += 1
				if s.offset > s.input.MaximumOffset {
					panic(fmt.Sprintf("Unexpected record offset(%d)>maximumOffset(%d)", s.offset, s.input.MaximumOffset))
				}
			}
		}
		recordsChannel.Close()
		return
	})
	return recordsChannel
}
