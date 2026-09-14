package update_cancel

import (
	"errors"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
)

const (
	UpdateCancelHandle = "update-cancel"
)

type workerInterceptor struct {
	interceptor.WorkerInterceptorBase
}

func NewWorkerInterceptor() interceptor.WorkerInterceptor {
	return &workerInterceptor{}
}

func (w *workerInterceptor) InterceptWorkflow(
	ctx workflow.Context,
	next interceptor.WorkflowInboundInterceptor,
) interceptor.WorkflowInboundInterceptor {
	i := &workflowInboundInterceptor{root: w}
	i.Next = next
	return i
}

type workflowInboundInterceptor struct {
	ctxMap map[string]workflow.CancelFunc
	interceptor.WorkflowInboundInterceptorBase
	root *workerInterceptor
}

func (w *workflowInboundInterceptor) Init(outbound interceptor.WorkflowOutboundInterceptor) error {
	w.ctxMap = make(map[string]workflow.CancelFunc)
	return w.Next.Init(outbound)
}

func (w *workflowInboundInterceptor) ExecuteWorkflow(ctx workflow.Context, in *interceptor.ExecuteWorkflowInput) (interface{}, error) {
	err := workflow.SetUpdateHandlerWithOptions(ctx, UpdateCancelHandle, func(ctx workflow.Context, updateID string) error {
		// Cancel the update
		w.ctxMap[updateID]()
		return nil
	}, workflow.UpdateHandlerOptions{
		Validator: func(ctx workflow.Context, updateID string) error {
			// Validate that the update ID is known
			if _, ok := w.ctxMap[updateID]; !ok {
				return errors.New("unknown update ID")
			}
			return nil
		},
	})
	if err != nil {
		return nil, err
	}
	return w.Next.ExecuteWorkflow(ctx, in)
}

func (w *workflowInboundInterceptor) ExecuteUpdate(ctx workflow.Context, in *interceptor.UpdateInput) (interface{}, error) {
	ctx, cancel := workflow.WithCancel(ctx)
	w.ctxMap[workflow.GetUpdateInfo(ctx).ID] = cancel
	return w.Next.ExecuteUpdate(ctx, in)
}
