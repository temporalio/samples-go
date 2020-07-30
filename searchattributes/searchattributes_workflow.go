package searchattributes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"go.temporal.io/api/common/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

/**
 * This sample shows how to use search attributes. (Note this feature only work with ElasticSearch)
 */

// ClientKey is the key for lookup
type ClientKey int

const (
	// Namespace used for this sample. "default" namespace always exists on the server.
	Namespace = "default"
	// TemporalClientKey for retrieving client from context
	TemporalClientKey ClientKey = iota
)

var (
	// ErrClientNotFound when client is not found on context
	ErrClientNotFound = errors.New("failed to retrieve client from context")
)

// SearchAttributesWorkflow workflow definition
func SearchAttributesWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("SearchAttributes workflow started")

	// get search attributes that provided when start workflow
	info := workflow.GetInfo(ctx)
	val := info.SearchAttributes.IndexedFields["CustomIntField"]
	var currentIntValue int
	err := converter.GetDefaultDataConverter().FromPayload(val, &currentIntValue)
	if err != nil {
		logger.Error("Get search attribute failed", "Error", err)
		return err
	}
	logger.Info("Current Search Attributes: ", "CustomIntField", currentIntValue)

	// upsert search attributes
	attributes := map[string]interface{}{
		"CustomIntField":      2, // update CustomIntField from 1 to 2, then insert other fields
		"CustomKeywordField":  "Update1",
		"CustomBoolField":     true,
		"CustomDoubleField":   3.14,
		"CustomDatetimeField": time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local),
		"CustomStringField":   "String field is for text. When query, it will be tokenized for partial match. StringTypeField cannot be used in Order By",
	}
	_ = workflow.UpsertSearchAttributes(ctx, attributes)

	// print current search attributes
	info = workflow.GetInfo(ctx)
	err = printSearchAttributes(info.SearchAttributes, logger)
	if err != nil {
		return err
	}

	// update search attributes again
	attributes = map[string]interface{}{
		"CustomKeywordField": "Update2",
	}
	_ = workflow.UpsertSearchAttributes(ctx, attributes)

	// print current search attributes
	info = workflow.GetInfo(ctx)
	err = printSearchAttributes(info.SearchAttributes, logger)
	if err != nil {
		return err
	}

	_ = workflow.Sleep(ctx, 2*time.Second) // wait update reflected on ElasticSearch

	// list workflow
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 2 * time.Minute,
		StartToCloseTimeout:    2 * time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	query := "CustomIntField=2 and CustomKeywordField='Update2' order by CustomDatetimeField DESC"
	var listResults []*workflowpb.WorkflowExecutionInfo
	err = workflow.ExecuteActivity(ctx, ListExecutions, query).Get(ctx, &listResults)
	if err != nil {
		logger.Error("Failed to list workflow executions.", "Error", err)
		return err
	}

	logger.Info("Workflow completed.", "Execution", listResults[0].String())

	return nil
}

func printSearchAttributes(searchAttributes *common.SearchAttributes, logger log.Logger) error {
	buf := new(bytes.Buffer)
	for k, v := range searchAttributes.IndexedFields {
		var currentVal interface{}
		err := converter.GetDefaultDataConverter().FromPayload(v, &currentVal)
		if err != nil {
			logger.Error(fmt.Sprintf("Get search attribute for key %s failed", k), "Error", err)
			return err
		}
		_, _ = fmt.Fprintf(buf, "%s=%v\n", k, currentVal)
	}
	logger.Info(fmt.Sprintf("Current Search Attributes: \n%s", buf.String()))
	return nil
}

func ListExecutions(ctx context.Context, query string) ([]*workflowpb.WorkflowExecutionInfo, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("List executions.", "Query", query)

	c, err := getClientFromContext(ctx)
	if err != nil {
		logger.Error("Error when get client")
		return nil, err
	}

	var executions []*workflowpb.WorkflowExecutionInfo
	var nextPageToken []byte
	for hasMore := true; hasMore; hasMore = len(nextPageToken) > 0 {
		resp, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Namespace:     Namespace,
			PageSize:      10,
			NextPageToken: nextPageToken,
			Query:         query,
		})
		if err != nil {
			return nil, err
		}

		executions = append(executions, resp.Executions...)

		nextPageToken = resp.NextPageToken
		activity.RecordHeartbeat(ctx, nextPageToken)
	}

	return executions, nil
}

func getClientFromContext(ctx context.Context) (client.Client, error) {
	logger := activity.GetLogger(ctx)
	c := ctx.Value(TemporalClientKey).(client.Client)
	if c == nil {
		logger.Error("Could not retrieve client from context.")
		return nil, ErrClientNotFound
	}

	return c, nil
}
