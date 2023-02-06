package batch_sliding_window

import (
	"fmt"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
)

type ProcessBatchInput struct {
	PageSize          int
	SlidingWindowSize int
	Offset            int
	MaximumOffset     int
	Progress          int
	// The set of ids
	CurrentRecords map[int]bool // recordId -> ignored boolean
}

type SingleRecordOrError struct {
	Record SingleRecord
	Err    error
}

func SlidingWindowWorkflow(ctx workflow.Context, input ProcessBatchInput) (recordCount int, err error) {
	workflowId := workflow.GetInfo(ctx).WorkflowExecution.ID
	offset := input.Offset
	progress := input.Progress
	currentRecords := input.CurrentRecords
	var childrenStartedByThisRun []workflow.ChildWorkflowFuture

	// Start loading records asynchronously
	recordsChannel := recordsPump(ctx, input, err, offset)

	// Process child workflow completion signals asynchronously
	reportCompletionChannel := workflow.GetSignalChannel(ctx, "ReportCompletion")
	workflow.Go(ctx, func(ctx workflow.Context) {
		var recordId int
		_ = reportCompletionChannel.Receive(ctx, &recordId)
		// TODO: Add duplicate signal check
		// if currentRecords contains recordId then
		progress += 1
		delete(currentRecords, recordId)
	})

	// Process records
	for {
		// After starting slidingWindowSize children blocks until a completion signal is received.
		workflow.Await(ctx, func() bool {
			return len(currentRecords) < input.SlidingWindowSize
		})
		var record SingleRecordOrError
		more := recordsChannel.Receive(ctx, &record)
		// Completes workflow, if no more records to process.
		if !more {
			// Awaits for all children to complete
			workflow.Await(ctx, func() bool {
				return len(currentRecords) == 0
			})
			return offset + len(childrenStartedByThisRun), nil
		}
		// record pump failed
		if record.Err != nil {
			return progress, record.Err
		}
		options := workflow.ChildWorkflowOptions{
			ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
			WorkflowID:        fmt.Sprintf("%s/%d", workflowId, record.Record.Id),
		}
		childCtx := workflow.WithChildOptions(ctx, options)
		child := workflow.ExecuteChildWorkflow(childCtx, RecordProcessorWorkflow, record)
		childrenStartedByThisRun = append(childrenStartedByThisRun, child)
		currentRecords[record.Record.Id] = true // value is ignored

		// Continues-as-new after starting pageSize children
		if len(childrenStartedByThisRun) == input.PageSize {
			// Waits for all children to start. Without this wait, workflow completion through
			// continue-as-new might lead to a situation when they never start.
			for _, child := range childrenStartedByThisRun {
				err := child.GetChildWorkflowExecution().Get(ctx, nil)
				// Is not expected as children automatically generated
				// IDs are not expected to collide.
				if err != nil {
					return progress, err
				}
			}
			// Returns ContinueAsNewError with new workflow input
			newInput := ProcessBatchInput{
				PageSize:          input.PageSize,
				SlidingWindowSize: input.SlidingWindowSize,
				Offset:            offset + len(childrenStartedByThisRun),
				MaximumOffset:     input.MaximumOffset,
				Progress:          progress,
				CurrentRecords:    currentRecords,
			}
			return 0, workflow.NewContinueAsNewError(ctx, SlidingWindowWorkflow, newInput)
		}
	}
}

// recordsPump pumps into the returned channel the batches of records loaded through GetRecords activity.
// A failure is reported as a last record with Err field set. The Channel is closed at the end of the records
// or on error.
func recordsPump(ctx workflow.Context, input ProcessBatchInput, err error, offset int) workflow.Channel {
	recordsChannel := workflow.NewChannel(ctx)
	// Goroutine pumps records into the recordsChannel
	workflow.Go(ctx, func(ctx workflow.Context) {
		var loader *RecordLoader
		for {
			var records []SingleRecord
			err = workflow.ExecuteActivity(ctx, loader.GetRecords, input.PageSize, offset).Get(ctx, &records)
			if err != nil {
				recordsChannel.Send(ctx, SingleRecordOrError{Err: err})
				recordsChannel.Close()
				return
			}
			for _, record := range records {
				recordsChannel.Send(ctx, SingleRecordOrError{Record: record})
				offset += 1
				if offset > input.MaximumOffset {
					recordsChannel.Close()
					return
				}
			}
		}
	})
	return recordsChannel
}
