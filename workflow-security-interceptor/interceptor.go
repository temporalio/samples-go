package workflow_security_interceptor

import (
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"time"
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
	i := &workflowInboundInterceptor{}
	i.Next = next
	return i
}

type workflowInboundInterceptor struct {
	interceptor.WorkflowInboundInterceptorBase
}

func (w *workflowInboundInterceptor) Init(next interceptor.WorkflowOutboundInterceptor) error {
	i := &workflowOutboundInterceptor{}
	i.Next = next
	return w.Next.Init(i)
}

type workflowOutboundInterceptor struct {
	interceptor.WorkflowOutboundInterceptorBase
}

func ValidateChildWorkflowTypeActivity(childWorkflowType string) (bool, error) {
	return childWorkflowType == "ChildWorkflow", nil
}

type validatedChildWorkflowFuture struct {
	fAllowed   workflow.Future
	sExecution workflow.Settable
	fExecution workflow.Future
	sResult    workflow.Settable
	fResult    workflow.Future
	child      workflow.ChildWorkflowFuture
}

func NewValidatedChildWorkflowFuture(ctx workflow.Context, allowed workflow.Future) *validatedChildWorkflowFuture {
	r := &validatedChildWorkflowFuture{}
	r.fAllowed = allowed
	r.fExecution, r.sExecution = workflow.NewFuture(ctx)
	r.fResult, r.sResult = workflow.NewFuture(ctx)
	return r
}

func (v validatedChildWorkflowFuture) Get(ctx workflow.Context, valuePtr interface{}) error {
	return v.fResult.Get(ctx, valuePtr)
}

func (v validatedChildWorkflowFuture) IsReady() bool {
	return v.fResult.IsReady()
}

func (v validatedChildWorkflowFuture) GetChildWorkflowExecution() workflow.Future {
	return v.fExecution
}

func (v validatedChildWorkflowFuture) SignalChildWorkflow(ctx workflow.Context, signalName string, data interface{}) workflow.Future {
	f, s := workflow.NewFuture(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		// wait for validation
		err := v.fAllowed.Get(ctx, nil)
		if err != nil {
			s.SetError(err)
		}
		s.Chain(v.child.SignalChildWorkflow(ctx, signalName, data))
	})
	return f
}

func (w *workflowOutboundInterceptor) ExecuteChildWorkflow(
	ctx workflow.Context,
	childWorkflowType string,
	args ...interface{},
) workflow.ChildWorkflowFuture {
	aCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
	})
	fAllowed := workflow.ExecuteActivity(aCtx, ValidateChildWorkflowTypeActivity, childWorkflowType)
	result := NewValidatedChildWorkflowFuture(ctx, fAllowed)
	workflow.Go(ctx, func(ctx workflow.Context) {
		var allowed bool
		err := fAllowed.Get(ctx, &allowed)
		if err != nil {
			result.sResult.SetError(err)
			return
		}
		if !allowed {
			result.sResult.SetError(temporal.NewApplicationError("Child workflow type \""+childWorkflowType+"\" not allowed", "not-allowed"))
			return
		}
		childFuture := w.Next.ExecuteChildWorkflow(ctx, childWorkflowType, args...)
		result.child = childFuture
		result.sExecution.Chain(childFuture.GetChildWorkflowExecution())
		result.sResult.Chain(childFuture)
	})

	return result
}
