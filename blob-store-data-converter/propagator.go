package blobstore_data_converter

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

type (
	// contextKey is an unexported type used as key for items stored in the
	// Context object
	contextKey int

	// propagator implements the custom context propagator
	propagator struct{}
)

const (
	_                   contextKey = iota
	PropagatedValuesKey            // The key used to store PropagatedValues in the context
)

// propagationKey is the key used by the propagator to pass values through the
// Temporal Workflow Event History headers
const propagationKey = "context-propagation"

// PropagatedValues is the struct stored on the context under PropagatedValuesKey
//
// converter.GetDefaultDataConverter() converts this into a json string to be stored in the
// Temporal Workflow Event History headers under propagationKey
type PropagatedValues struct {
	TenantID       string   `json:"tenantID,omitempty"`
	BlobNamePrefix []string `json:"bsPathSegs,omitempty"`
}

// UnknownTenant returns a PropagatedValues struct with a default values
// This happens in edge cases where the tenantID is not set in the context
func UnknownTenant() PropagatedValues {
	return PropagatedValues{
		TenantID: "unknown-tenant",
	}
}

// NewContextPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewContextPropagator() workflow.ContextPropagator {
	return &propagator{}
}

// Inject injects values from context into headers for propagation
func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(PropagatedValuesKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *propagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	vals := ctx.Value(PropagatedValuesKey).(PropagatedValues)

	payload, err := converter.GetDefaultDataConverter().ToPayload(vals)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// errMissingHeaderContextPropagationKey is an edge case that can happen when the UI/CLI is used
// to start, signal, or query a workflow. It's up to the user to define this behavior.
//
// In this example, we just log the error and continue with a default value.
// This allows UI/CLIs to send json payloads. This also protects the workflow from failing to find the missing ctx key.
var errMissingHeaderContextPropagationKey = fmt.Errorf("context propagation key not found in header: %s", propagationKey)

// Extract extracts values from headers and puts them into context
func (s *propagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	value, ok := reader.Get(propagationKey)
	if !ok {
		fmt.Println(errMissingHeaderContextPropagationKey)
		return context.WithValue(ctx, PropagatedValuesKey, UnknownTenant()), nil
	}

	var data PropagatedValues
	if err := converter.GetDefaultDataConverter().FromPayload(value, &data); err != nil {
		return ctx, fmt.Errorf("failed to extract value from header: %w", err)
	}
	ctx = context.WithValue(ctx, PropagatedValuesKey, data)

	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *propagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	value, ok := reader.Get(propagationKey)
	if !ok {
		fmt.Println(errMissingHeaderContextPropagationKey)
		return workflow.WithValue(ctx, PropagatedValuesKey, UnknownTenant()), nil
	}

	var data PropagatedValues
	if err := converter.GetDefaultDataConverter().FromPayload(value, &data); err != nil {
		return ctx, fmt.Errorf("failed to extract value from header: %w", err)
	}
	ctx = workflow.WithValue(ctx, PropagatedValuesKey, data)

	return ctx, nil
}
