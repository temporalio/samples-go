package nexuscontextpropagation

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/temporalio/samples-go/ctxpropagation"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/encoding/protojson"
)

type WorkerInterceptor struct {
	interceptor.WorkerInterceptorBase
	DataConverter converter.DataConverter
}

func (w *WorkerInterceptor) InterceptWorkflow(ctx workflow.Context, next interceptor.WorkflowInboundInterceptor) interceptor.WorkflowInboundInterceptor {
	in := &workflowInboundInterceptor{parent: w}
	in.Next = next
	return in
}

func (w *WorkerInterceptor) InterceptNexusOperation(ctx context.Context, next interceptor.NexusOperationInboundInterceptor) interceptor.NexusOperationInboundInterceptor {
	i := &nexusOperationInboundInterceptor{parent: w}
	i.Next = next
	return i
}

type workflowInboundInterceptor struct {
	interceptor.WorkflowInboundInterceptorBase
	parent *WorkerInterceptor
}

func (in *workflowInboundInterceptor) Init(next interceptor.WorkflowOutboundInterceptor) error {
	out := &workflowOutboundInterceptor{parent: in.parent}
	out.Next = next
	return in.Next.Init(out)
}

type workflowOutboundInterceptor struct {
	interceptor.WorkflowOutboundInterceptorBase
	parent *WorkerInterceptor
}

type nexusErrorFuture struct {
	workflow.Future
}

func newNexusErrorFuture(ctx workflow.Context, err error) nexusErrorFuture {
	fut, settable := workflow.NewFuture(ctx)
	settable.SetError(err)
	return nexusErrorFuture{fut}
}

func (n nexusErrorFuture) GetNexusOperationExecution() workflow.Future {
	// Return the same future
	return n
}

// ExecuteNexusOperation implements interceptor.WorkflowOutboundInterceptor. It extracts values from workflow context
// and propagates them via a Nexus header.
func (out *workflowOutboundInterceptor) ExecuteNexusOperation(
	ctx workflow.Context,
	input interceptor.ExecuteNexusOperationInput,
) workflow.NexusOperationFuture {
	if values, ok := ctx.Value(ctxpropagation.PropagateKey).(ctxpropagation.Values); ok {
		payload, err := out.parent.DataConverter.ToPayload(values)
		if err != nil {
			return newNexusErrorFuture(ctx, fmt.Errorf("cannot encode context values: %w", err))
		}
		data, err := protojson.Marshal(payload)
		if err != nil {
			return newNexusErrorFuture(ctx, fmt.Errorf("cannot marshal context payload to JSON: %w", err))
		}

		h := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(data)
		input.NexusHeader[ctxpropagation.HeaderKey] = h
	}
	return out.Next.ExecuteNexusOperation(ctx, input)
}

// nexusOperationInboundInterceptor implements NexusOperationInboundInterceptor to intercept StartOperation.
// Implementation may also implement Init to inject a NexusOperationOutboundInterceptor that can customize logging,
// metrics, and the client, as well as CancelOperation to intercept operation cancelation.
type nexusOperationInboundInterceptor struct {
	interceptor.NexusOperationInboundInterceptorBase
	parent *WorkerInterceptor
}

// StartOperation implements internal.NexusOperationInboundInterceptor. It extracts context propagated via a Nexus
// header into a Go context value.
func (n *nexusOperationInboundInterceptor) StartOperation(ctx context.Context, input interceptor.NexusStartOperationInput) (nexus.HandlerStartOperationResult[any], error) {
	if h := input.Options.Header[ctxpropagation.HeaderKey]; h != "" {
		data, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(h)
		if err != nil {
			return nil, nexus.HandlerErrorf(nexus.HandlerErrorTypeBadRequest, "invalid %s header: %w", ctxpropagation.HeaderKey, err)
		}
		var payload common.Payload
		if err := protojson.Unmarshal(data, &payload); err != nil {
			return nil, nexus.HandlerErrorf(nexus.HandlerErrorTypeBadRequest, "invalid %s header: %w", ctxpropagation.HeaderKey, err)
		}
		var values ctxpropagation.Values
		if err := n.parent.DataConverter.FromPayload(&payload, &values); err != nil {
			return nil, nexus.HandlerErrorf(nexus.HandlerErrorTypeBadRequest, "invalid %s header: %w", ctxpropagation.HeaderKey, err)
		}
		ctx = context.WithValue(ctx, ctxpropagation.PropagateKey, values)
	}
	return n.Next.StartOperation(ctx, input)
}
