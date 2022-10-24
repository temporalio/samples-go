package cancel_in_progress

import (
	"errors"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

const ParentWorkflowSignalName = "parent-workflow-signal"

func SampleParentWorkflow(ctx workflow.Context) (result string, err error) {
	logger := workflow.GetLogger(ctx)

	var message string

	reBuildSignalChan := workflow.GetSignalChannel(ctx, ParentWorkflowSignalName)

	// This will not block because the workflow is started with a signal transactional.
	reBuildSignalChan.Receive(ctx, &message)

	cwo := workflow.ChildWorkflowOptions{
		WaitForCancellation: true,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var isProcessingDone = false

	for !isProcessingDone {
		ctx, cancelHandler := workflow.WithCancel(ctx)

		// it is important re-execute the child workflow after every signal
		// because we might have cancelled the previous execution
		childWorkflowFuture := workflow.ExecuteChildWorkflow(ctx, SampleChildWorkflow, message)

		selector := workflow.NewSelector(ctx)

		selector.AddFuture(childWorkflowFuture, func(f workflow.Future) {
			err = f.Get(ctx, &result)

			// we don't want to end the parent workflow when child workflow is cancelled
			if errors.Is(ctx.Err(), workflow.ErrCanceled) {
				logger.Info("Child execution cancelled.")
				return
			}

			if err != nil {
				logger.Error("Child execution failure.", "Error", err)
			}

			// Child workflow completed. Let's end the parent workflow.
			isProcessingDone = true
		})

		selector.AddReceive(reBuildSignalChan, func(c workflow.ReceiveChannel, more bool) {
			logger.Info("Received signal.", "Message", message)

			// cancel the child workflow as fast as possible after we received a new signal
			// if a child workflow execution is in progress we cancel everything and start a new child workflow (for loop)
			cancelHandler()

			// drain the channel to get the latest signal
			// Users might send multiple signals in a short period of time
			// and we are only interested in the latest signal
			message = GetLatestMessageFromChannel(logger, reBuildSignalChan)
		})

		// wait for the build workflow to finish or the signal to cancel the running execution through a signal
		selector.Select(ctx)

		// in case of an error we want to cancel the child workflow
		if err != nil {
			logger.Error("Child execution failure.", "Error", err)
			return "", err
		}
	}

	logger.Info("Parent execution completed.", "Result", result)

	return result, nil
}

func GetLatestMessageFromChannel(logger log.Logger, ch workflow.ReceiveChannel) string {
	var message string
	var messages []string

	for {
		var m string
		if ch.ReceiveAsync(&m) {
			messages = append(messages, m)
			logger.Info("Additional message received.", "message", message)
		} else {
			break
		}
	}

	for i, m := range messages {
		// continue with the latest signal
		if i == len(messages)-1 {
			// Update the workflow options to use the latest message for the next child workflow execution
			message = m
			logger.Info("Continue with latest message.", "message", message)
		} else {
			logger.Info("Cancel old workflow execution.", "message", message)
			// You might want to do some cleanup here
		}
	}

	return message
}
