package main

import (
	"context"
	"encoding/json"

	"go.uber.org/cadence/workflow"
)

type (
	// contextKey is an unexported type used as key for items stored in the
	// Context object
	contextKey struct{}

	// propagator implements the custom context propagator
	propagator struct{}

	// Values is a struct holding values
	Values struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
)

// propagateKey is the key used to store the value in the Context object
var propagateKey = contextKey{}

// propagationKey is the key used by the propagator to pass values through the
// cadence server headers
const propagationKey = "_prop"

// NewContextPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewContextPropagator() workflow.ContextPropagator {
	return &propagator{}
}

// Inject injects values from context into headers for propagation
func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(propagateKey)
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *propagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(propagateKey)
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *propagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if key == propagationKey {
			var values Values
			if err := json.Unmarshal(value, &values); err != nil {
				return err
			}
			ctx = context.WithValue(ctx, propagateKey, values)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *propagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if key == propagationKey {
			var values Values
			if err := json.Unmarshal(value, &values); err != nil {
				return err
			}
			ctx = workflow.WithValue(ctx, propagateKey, values)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}
