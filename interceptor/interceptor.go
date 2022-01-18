package interceptor

import (
	"context"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

type workerInterceptor struct {
	interceptor.WorkerInterceptorBase
	customLogTags []interface{}
}

type InterceptorOptions struct {
	CustomLogTags map[string]interface{}
}

func NewWorkerInterceptor(options InterceptorOptions) interceptor.WorkerInterceptor {
	// Convert map to slice
	tags := make([]interface{}, 0, len(options.CustomLogTags)*2)
	for k, v := range options.CustomLogTags {
		tags = append(tags, k, v)
	}
	return &workerInterceptor{customLogTags: tags}
}

func (w *workerInterceptor) withCustomLogTags(logger log.Logger) log.Logger {
	if len(w.customLogTags) > 0 {
		return log.With(logger, w.customLogTags...)
	}
	return logger
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
	// Set our custom tags
	return a.root.withCustomLogTags(a.Next.GetLogger(ctx))
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
	// Set our custom tags
	return w.root.withCustomLogTags(w.Next.GetLogger(ctx))
}
