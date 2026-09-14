package sagainterceptor

import (
	"time"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/multierr"
)

var (
	defaultActivityOpts = workflow.ActivityOptions{
		ScheduleToStartTimeout: 1 * time.Minute,
		StartToCloseTimeout:    5 * time.Minute,
	}
)

type (
	// SagaOptions is options for a saga transactional workflow
	SagaOptions struct{}

	// CompensationOptions is options for compensate.
	CompensationOptions struct {
		// ActivityType is the name of compensate activity.
		ActivityType string
		// ActivityOptions is the activity execute options, local activity is not supported.
		ActivityOptions *workflow.ActivityOptions
		// Converter optional. Convert req & response to request for compensate activity.
		// currently, activity func is not available for worker, so decode futures should be done by developer.
		Converter func(ctx workflow.Context, f workflow.Future, args ...interface{}) ([]interface{}, error)
	}

	//InterceptorOptions is options for saga interceptor.
	InterceptorOptions struct {
		// WorkflowRegistry names for workflow to be treated as Saga transaction.
		WorkflowRegistry map[string]SagaOptions
		// ActivityRegistry Action -> CompensateAction, key is activity type for action.
		ActivityRegistry map[string]CompensationOptions
	}

	sagaInterceptor struct {
		interceptor.WorkerInterceptorBase
		options InterceptorOptions
	}

	workflowInboundInterceptor struct {
		interceptor.WorkflowInboundInterceptorBase
		root *sagaInterceptor
		ctx  sagaContext
	}

	workflowOutboundInterceptor struct {
		interceptor.WorkflowOutboundInterceptorBase
		root *sagaInterceptor
		ctx  *sagaContext
	}

	compensation struct {
		Options      *CompensationOptions
		ActionFuture workflow.Future
		ActionArgs   []interface{}
	}

	sagaContext struct {
		Compensations []*compensation
	}
)

// NewInterceptor creates an interceptor for execute in Saga patterns.
// when workflow fails, registered compensate activities will be executed automatically.
func NewInterceptor(options InterceptorOptions) (interceptor.WorkerInterceptor, error) {
	return &sagaInterceptor{options: options}, nil
}

func (s *sagaInterceptor) InterceptWorkflow(
	ctx workflow.Context,
	next interceptor.WorkflowInboundInterceptor,
) interceptor.WorkflowInboundInterceptor {
	if _, ok := s.options.WorkflowRegistry[workflow.GetInfo(ctx).WorkflowType.Name]; !ok {
		return next
	}

	workflow.GetLogger(ctx).Debug("intercept saga workflow")
	i := &workflowInboundInterceptor{root: s}
	i.Next = next
	return i
}

func (w *workflowInboundInterceptor) Init(outbound interceptor.WorkflowOutboundInterceptor) error {
	i := &workflowOutboundInterceptor{root: w.root, ctx: &w.ctx}
	i.Next = outbound
	return w.Next.Init(i)
}

func (w *workflowInboundInterceptor) ExecuteWorkflow(
	ctx workflow.Context,
	in *interceptor.ExecuteWorkflowInput,
) (ret interface{}, err error) {
	workflow.GetLogger(ctx).Debug("intercept ExecuteWorkflow")
	ret, wferr := w.Next.ExecuteWorkflow(ctx, in)
	if wferr == nil || len(w.ctx.Compensations) == 0 {
		return ret, wferr
	}

	ctx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	for i := len(w.ctx.Compensations) - 1; i >= 0; i-- {
		c := w.ctx.Compensations[i]
		// only compensate action with success
		if err := c.ActionFuture.Get(ctx, nil); err != nil {
			continue
		}

		// add opts if not config
		activityOpts := c.Options.ActivityOptions
		if activityOpts == nil {
			activityOpts = &defaultActivityOpts
		}
		ctx = workflow.WithActivityOptions(ctx, *activityOpts)

		// use arg in action as default for compensate
		args := c.ActionArgs
		if c.Options.Converter != nil {
			args, err = c.Options.Converter(ctx, c.ActionFuture, c.ActionArgs...)
			if err != nil {
				workflow.GetLogger(ctx).Error("failed to convert to compensate req", "error", err)
				return ret, multierr.Append(wferr, err)
			}
		}

		if err := workflow.ExecuteActivity(ctx, c.Options.ActivityType, args...).Get(ctx, nil); err != nil {
			// best effort, save error and continue
			//wferr = multierr.Append(wferr, err)

			// fail fast, one compensation fails, it will not continue
			return ret, multierr.Append(wferr, err)
		}
	}
	return ret, wferr
}

func (w *workflowOutboundInterceptor) ExecuteActivity(
	ctx workflow.Context,
	activityType string,
	args ...interface{},
) workflow.Future {
	workflow.GetLogger(ctx).Debug("intercept ExecuteActivity", "activity_type", activityType)
	f := w.Next.ExecuteActivity(ctx, activityType, args...)
	if opts, ok := w.root.options.ActivityRegistry[activityType]; ok {
		workflow.GetLogger(ctx).Debug("save action future", "activity_type", activityType)
		w.ctx.Compensations = append(w.ctx.Compensations, &compensation{
			Options:      &opts,
			ActionArgs:   args,
			ActionFuture: f,
		})
	}

	return f
}

func (w *workflowOutboundInterceptor) ExecuteLocalActivity(
	ctx workflow.Context,
	activityType string,
	args ...interface{},
) workflow.Future {
	workflow.GetLogger(ctx).Debug("intercept ExecuteLocalActivity", "activity_type", activityType)
	f := w.Next.ExecuteLocalActivity(ctx, activityType, args...)
	if opts, ok := w.root.options.ActivityRegistry[activityType]; ok {
		workflow.GetLogger(ctx).Debug("save action future", "activity_type", activityType)
		w.ctx.Compensations = append(w.ctx.Compensations, &compensation{
			Options:      &opts,
			ActionArgs:   args,
			ActionFuture: f,
		})
	}

	return f
}
