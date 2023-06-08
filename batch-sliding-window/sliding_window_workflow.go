package batch_sliding_window

import (
	"fmt"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"sort"
	"time"
)

type (
	// SlidingWindowWorkflowInput contains SlidingWindowWorkflow arguments
	SlidingWindowWorkflowInput struct {
		PageSize          int
		SlidingWindowSize int
		Offset            int // inclusive
		MaximumOffset     int // exclusive
		Progress          int
		// The set of ids
		CurrentRecords map[int]bool // recordId -> ignored boolean
	}

	// SlidingWindow structure that implements the workflow logic
	SlidingWindow struct {
		input SlidingWindowWorkflowInput
		// currentRecords represents a set of records that are currently being processed by child workflows.
		// key is recordId. values are ignored.
		currentRecords map[int]bool
		// childrenStartedByThisRun is used to wait for children to start before calling continue as new.
		childrenStartedByThisRun []workflow.ChildWorkflowFuture
		// Offset into the next record to process.
		offset int
		// Count of completed records.
		progress int
		// completionSignalPumpCancellationHandler is used to request pump completion
		completionSignalPumpCancellationHandler workflow.CancelFunc
		// completionSignalPumpCompletion is used to wait for the pump completion.
		completionSignalPumpCompletion workflow.Future
	}

	// SlidingWindowState used as a "state" query result.
	SlidingWindowState struct {
		// currentRecords represents a set of record ids that are currently being processed by child workflows.
		CurrentRecords           []int
		ChildrenStartedByThisRun int
		Offset                   int
		Progress                 int
	}
)

// SlidingWindowWorkflow workflow processes a range of records using a requested number of child workflows.
// As soon as a child workflow completes a new one is started.
func SlidingWindowWorkflow(ctx workflow.Context, input SlidingWindowWorkflowInput) (recordCount int, err error) {
	workflow.GetLogger(ctx).Info("SlidingWindowWorkflow",
		"input", input.SlidingWindowSize,
		"PageSize", input.PageSize,
		"Offset", input.Offset,
		"MaximumOffset", input.MaximumOffset,
		"Progress", input.Progress)

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
	currentRecordIds := make([]int, len(s.currentRecords))
	i := 0
	// Range over map is a nondeterministic operation.
	// It is OK to have a non-deterministic operation in a query function.
	// Sorting of results makes the result deterministic anyway.
	for k := range s.currentRecords {
		currentRecordIds[i] = k
		i++
	}
	sort.Ints(currentRecordIds)
	return SlidingWindowState{
		CurrentRecords:           currentRecordIds,
		ChildrenStartedByThisRun: len(s.childrenStartedByThisRun),
		Offset:                   s.offset,
		Progress:                 s.progress,
	}, nil
}

func (s *SlidingWindow) Execute(ctx workflow.Context) (recordCount int, err error) {
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
		// Blocks until the total number of children (including started by the previous runs)
		// gets below the SlidingWindowSize.
		err := workflow.Await(ctx, func() bool {
			return len(s.currentRecords) < s.input.SlidingWindowSize
		})
		if err != nil {
			return 0, err
		}

		options := workflow.ChildWorkflowOptions{
			// Use ABANDON as child workflows have to survive the parent calling continue-as-new
			ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
			// Human readable child id.
			WorkflowID: fmt.Sprintf("%s/%d", workflowId, record.Id),
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
			// Is not expected as children's automatically generated
			// IDs are not expected to collide.
			if err != nil {
				return 0, err
			}
		}
		// Must drain the signal channel without blocking before calling continue-as-new.
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
	// The last run in the continue-as-new chain.
	// Awaits for all children to complete
	err := workflow.Await(ctx, func() bool {
		return len(s.currentRecords) == 0
	})
	if err != nil {
		return 0, err
	}
	return s.progress, nil
}

func (s *SlidingWindow) drainCompletionSignalChannelAsync(ctx workflow.Context) {
	// Request pump completion
	s.completionSignalPumpCancellationHandler()
	// Wait for the pump to complete to avoid signal loss.
	_ = s.completionSignalPumpCompletion.Get(ctx, nil)

	reportCompletionChannel := workflow.GetSignalChannel(ctx, "ReportCompletion")
	// Drains signals async
	for {
		var recordId int
		ok := reportCompletionChannel.ReceiveAsync(&recordId)
		if !ok {
			break
		}
		s.recordCompletion(ctx, recordId)
	}
}

// completionSignalPump asynchronously processes ReportCompletion signals.
// There is no need to clean up the pump goroutine in case a workflow completes due to an error.
// All goroutines are released back automatically upon workflow completion.
func (s *SlidingWindow) completionSignalPump(ctx workflow.Context) {
	// completionSignalPumpCancellationHandler is used to request pump completion
	ctx, cancellationHandler := workflow.WithCancel(ctx)
	s.completionSignalPumpCancellationHandler = cancellationHandler

	// completionSignalPumpCompletion is used to wait for the pump completion.
	completed, completedSettable := workflow.NewFuture(ctx)
	s.completionSignalPumpCompletion = completed

	reportCompletionChannel := workflow.GetSignalChannel(ctx, "ReportCompletion")

	workflow.Go(ctx, func(ctx workflow.Context) {
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(reportCompletionChannel, func(c workflow.ReceiveChannel, more bool) {
			var recordId int
			_ = reportCompletionChannel.Receive(ctx, &recordId)
			s.recordCompletion(ctx, recordId)
		})
		selector.AddReceive(ctx.Done(), func(c workflow.ReceiveChannel, more bool) {
			completedSettable.Set(nil, nil)
		})
		for !completed.IsReady() {
			selector.Select(ctx)
		}
	})
}

func (s *SlidingWindow) recordCompletion(ctx workflow.Context, recordId int) {
	// duplicate signal check
	if _, ok := s.currentRecords[recordId]; ok {
		delete(s.currentRecords, recordId)
		s.progress += 1
	}
}
