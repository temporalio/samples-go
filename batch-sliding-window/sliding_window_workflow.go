package batch_sliding_window

import (
	"fmt"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"time"
)

type (
	SlidingWindowWorkflowInput struct {
		PageSize          int
		SlidingWindowSize int
		Offset            int // inclusive
		MaximumOffset     int // exclusive
		Progress          int
		// The set of ids
		CurrentRecords map[int]bool // recordId -> ignored boolean
	}

	SlidingWindow struct {
		input                    SlidingWindowWorkflowInput
		currentRecords           map[int]bool
		childrenStartedByThisRun []workflow.ChildWorkflowFuture
		offset                   int
		progress                 int
	}

	SlidingWindowState struct {
		CurrentRecords           map[int]bool
		ChildrenStartedByThisRun int
		Offset                   int
		Progress                 int
	}
)

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
	err = workflow.SetQueryHandler(ctx, "state", func() (SlidingWindowState, error) {
		return impl.State()
	})
	if err != nil {
		return 0, err
	}
	return impl.Execute(ctx)
}

// State returns the current state of the batch.
// Used by the "state" workflow query.
func (s *SlidingWindow) State() (SlidingWindowState, error) {
	return SlidingWindowState{
		CurrentRecords:           s.currentRecords,
		ChildrenStartedByThisRun: len(s.childrenStartedByThisRun),
		Offset:                   s.offset,
		Progress:                 s.progress,
	}, nil
}

func (s *SlidingWindow) Execute(ctx workflow.Context) (recordCount int, err error) {
	fmt.Println("Execute Started !!!!!!!!!!! runId=" + workflow.GetInfo(ctx).WorkflowExecution.RunID)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	})

	// Starts processing child workflow completion signals asynchronously
	s.completionSignalPump(ctx)

	var getOutput GetRecordsOutput
	if s.offset < s.input.MaximumOffset {
		getInput := &GetRecordsInput{
			PageSize:  s.input.PageSize,
			Offset:    s.offset,
			MaxOffset: s.input.MaximumOffset,
		}
		var loader *RecordLoader
		err = workflow.ExecuteActivity(ctx, loader.GetRecords, getInput).Get(ctx, &getOutput)
		if err != nil {
			return 0, err
		}
	}
	workflowId := workflow.GetInfo(ctx).WorkflowExecution.ID
	// Process records
	for _, record := range getOutput.Records {
		// Blocks until the total number of children (including started by previous runs)
		// gets below the SlidingWindowSize.
		workflow.Await(ctx, func() bool {
			return len(s.currentRecords) < s.input.SlidingWindowSize
		})

		options := workflow.ChildWorkflowOptions{
			ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
			WorkflowID:        fmt.Sprintf("%s/%d", workflowId, record.Id),
		}
		childCtx := workflow.WithChildOptions(ctx, options)
		child := workflow.ExecuteChildWorkflow(childCtx, RecordProcessorWorkflow, record)

		s.childrenStartedByThisRun = append(s.childrenStartedByThisRun, child)
		s.currentRecords[record.Id] = true // value is ignored
	}
	return s.continueAsNewOrComplete(ctx)
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
		workflow.GetLogger(ctx).Info("NewContinueAsNewError")

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
	workflow.GetLogger(ctx).Info("drainCompletionSignalChannelAsync start")

	reportCompletionChannel := workflow.GetSignalChannel(ctx, "ReportCompletion")
	for {
		var recordId int
		ok := reportCompletionChannel.ReceiveAsync(&recordId)
		if !ok {
			workflow.GetLogger(ctx).Info("drainCompletionSignalChannelAsync drained")
			break
		}
		workflow.GetLogger(ctx).Info("drainCompletionSignalChannelAsync", "recordId", recordId)
		s.recordCompletion(ctx, recordId)
	}
}

// completionSignalPump asynchronously processes ReportCompletion signals.
// There is no need to clean up the pump goroutine in case a workflow completes due to an error.
// All goroutines are released back automatically upon workflow completion.
func (s *SlidingWindow) completionSignalPump(ctx workflow.Context) {
	reportCompletionChannel := workflow.GetSignalChannel(ctx, "ReportCompletion")
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var recordId int
			_ = reportCompletionChannel.Receive(ctx, &recordId)
			workflow.GetLogger(ctx).Info("completionSignalPump calls recordCompletion", "recordId", recordId)
			s.recordCompletion(ctx, recordId)
		}
	})
}

func (s *SlidingWindow) recordCompletion(ctx workflow.Context, recordId int) {
	workflow.GetLogger(ctx).Info("recordCompletion", "recordId", recordId)
	// duplicate signal check
	if _, ok := s.currentRecords[recordId]; ok {
		workflow.GetLogger(ctx).Info("recordCompletion recorded", "recordId", recordId)
		delete(s.currentRecords, recordId)
		s.progress += 1
	}
}
