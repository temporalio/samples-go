package interceptor

import (
	"context"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type workerInterceptor struct {
	interceptor.WorkerInterceptorBase
	options InterceptorOptions
}

type InterceptorOptions struct {
	GetExtraLogTagsForWorkflow func(workflow.Context) []interface{}
	GetExtraLogTagsForActivity func(context.Context) []interface{}
}

func NewWorkerInterceptor(options InterceptorOptions) interceptor.WorkerInterceptor {
	return &workerInterceptor{options: options}
}

func (w *workerInterceptor) InterceptActivity(
	ctx context.Context,
	next interceptor.ActivityInboundInterceptor,
) interceptor.ActivityInboundInterceptor {
	i := &activityInboundInterceptor{root: w}
	i.Next = next
	return i
}

type activityInboundInterceptor struct {
	interceptor.ActivityInboundInterceptorBase
	root *workerInterceptor
}

func (a *activityInboundInterceptor) Init(outbound interceptor.ActivityOutboundInterceptor) error {
	i := &activityOutboundInterceptor{root: a.root}
	i.Next = outbound
	return a.Next.Init(i)
}

type activityOutboundInterceptor struct {
	interceptor.ActivityOutboundInterceptorBase
	root *workerInterceptor
}

func (a *activityOutboundInterceptor) GetLogger(ctx context.Context) log.Logger {
	logger := a.Next.GetLogger(ctx)
	// Add extra tags if any
	if a.root.options.GetExtraLogTagsForActivity != nil {
		if extraTags := a.root.options.GetExtraLogTagsForActivity(ctx); len(extraTags) > 0 {
			logger = log.With(logger, extraTags...)
		}
	}
	return logger
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
	interceptor.WorkflowInboundInterceptorBase
	root *workerInterceptor
}

func (w *workflowInboundInterceptor) Init(outbound interceptor.WorkflowOutboundInterceptor) error {
	i := &workflowOutboundInterceptor{root: w.root}
	i.Next = outbound
	return w.Next.Init(i)
}

type workflowOutboundInterceptor struct {
	interceptor.WorkflowOutboundInterceptorBase
	root *workerInterceptor
}

func (w *workflowOutboundInterceptor) GetLogger(ctx workflow.Context) log.Logger {
	logger := w.Next.GetLogger(ctx)
	// Add extra tags if any
	if w.root.options.GetExtraLogTagsForWorkflow != nil {
		if extraTags := w.root.options.GetExtraLogTagsForWorkflow(ctx); len(extraTags) > 0 {
			logger = log.With(logger, extraTags...)
		}
	}
	return logger
}
