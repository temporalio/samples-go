// Package searchattributes is a sample that shows how to access search attributes from a workflow and query workflows.
package searchattributes

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var (
	CustomIntField         = temporal.NewSearchAttributeKeyInt64("CustomIntField")
	CustomKeywordField     = temporal.NewSearchAttributeKeyKeyword("CustomKeywordField")
	CustomBoolField        = temporal.NewSearchAttributeKeyBool("CustomBoolField")
	CustomDoubleField      = temporal.NewSearchAttributeKeyFloat64("CustomDoubleField")
	CustomDatetimeField    = temporal.NewSearchAttributeKeyTime("CustomDatetimeField")
	CustomKeywordListField = temporal.NewSearchAttributeKeyKeywordList("CustomKeywordListField")
)

// SearchAttributesWorkflow workflow definition
func SearchAttributesWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("SearchAttributes workflow started")
	sas := workflow.GetTypedSearchAttributes(ctx)
	currentIntValue, ok := sas.GetInt64(CustomIntField)
	if !ok {
		return errors.New("expected CustomIntField to be set")
	}
	logger.Info("Current search attribute value.", "CustomIntField", currentIntValue)

	// Upsert search attributes.
	// This won't persist search attributes on server because commands are sent to the server only when all workflow
	// goroutines are blocked, but the local view will be updated.
	err := workflow.UpsertTypedSearchAttributes(
		ctx,
		CustomIntField.ValueSet(2), // update CustomIntField from 1 to 2, then insert other fields
		CustomKeywordField.ValueSet("Keyword fields supports prefix search"),
		CustomBoolField.ValueSet(true),
		CustomDoubleField.ValueSet(3.14),
		CustomDatetimeField.ValueSet(workflow.Now(ctx).UTC()),
		CustomKeywordListField.ValueSet([]string{"value1", "value2"}),
	)
	if err != nil {
		return err
	}
	// Print current search attributes with the modifications above.
	printSearchAttributes(ctx)

	// Unset values with ValueUnset.
	err = workflow.UpsertTypedSearchAttributes(ctx, CustomDoubleField.ValueUnset())
	printSearchAttributes(ctx)

	// Yield to allow commands to be sent to the server and let the visibility store update.
	if err := workflow.Sleep(ctx, 1*time.Second); err != nil {
		return err
	}

	// After Elasticsearch index is updated we can query it.
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		HeartbeatTimeout:    time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	query := fmt.Sprintf(
		"%s=2 AND %s STARTS_WITH 'Keyword fields'",
		CustomIntField.GetName(),
		CustomKeywordField.GetName(),
	)
	var listResults []*workflowpb.WorkflowExecutionInfo
	err = workflow.ExecuteActivity(ctx, ListExecutions, query).Get(ctx, &listResults)
	if err != nil {
		logger.Error("Failed to list workflow executions.", "Error", err)
		return err
	}

	// In case this workflow is run multiples times, the first execution returned in the list should be the current
	// one because they are ordered by CustomDatetimeField (which is set as time.Now).
	lastExecution := listResults[0]

	info := workflow.GetInfo(ctx)
	logger.Info("WorkflowID must be the same", "info.WorkflowID", info.WorkflowExecution.ID, "lastExecution.WorkflowId", lastExecution.GetExecution().GetWorkflowId())
	logger.Info("RunID must be the same", "info.RunID", info.WorkflowExecution.RunID, "lastExecution.RunId", lastExecution.GetExecution().GetRunId())

	// No change should be expected since the last print call.
	printSearchAttributes(ctx)

	logger.Info("Workflow completed.")
	return nil
}

func printSearchAttributes(ctx workflow.Context) {
	logger := workflow.GetLogger(ctx)
	sas := workflow.GetTypedSearchAttributes(ctx)
	for k, v := range sas.GetUntypedValues() {
		logger.Info("Current search attribute value", k.GetName(), v)
	}
}

func ListExecutions(ctx context.Context, query string) ([]*workflowpb.WorkflowExecutionInfo, error) {
	logger := activity.GetLogger(ctx)
	c := activity.GetClient(ctx)
	info := activity.GetInfo(ctx)

	logger.Info("List executions.", "query", query)

	var executions []*workflowpb.WorkflowExecutionInfo
	var nextPageToken []byte
	var seenCurrentExecution bool

	// Wait until we've seen our current execution surface as the first result and we've listed all previous
	// executions.
	// Waiting is required since the visibility store is eventually consistent and may take time to index the
	// current execution.
	for hasMore := true; hasMore; hasMore = !seenCurrentExecution || len(nextPageToken) > 0 {
		resp, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     info.WorkflowNamespace,
			PageSize:      10,
			NextPageToken: nextPageToken,
			Query:         query,
		})
		if err != nil {
			return nil, err
		}

		seenCurrentExecution = seenCurrentExecution || slices.ContainsFunc(resp.Executions, func(exec *workflowpb.WorkflowExecutionInfo) bool {
			return exec.Execution.RunId == info.WorkflowExecution.RunID
		})
		if seenCurrentExecution {
			executions = append(executions, resp.Executions...)
			nextPageToken = resp.NextPageToken
		}

		activity.RecordHeartbeat(ctx, nextPageToken)
	}

	return executions, nil
}
