package searchattributes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.temporal.io/temporal-proto/common"
	"go.temporal.io/temporal-proto/workflowservice"
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This sample shows how to use search attributes. (Note this feature only work with ElasticSearch)
 */

// ClientKey is the key for lookup
type ClientKey int

const (
	// DomainName used for this sample
	DomainName = "default"
	// TemporalClientKey for retrieving client from context
	TemporalClientKey ClientKey = iota
)

var (
	// ErrClientNotFound when client is not found on context
	ErrClientNotFound = errors.New("failed to retrieve client from context")
)

// SearchAttributesWorkflow workflow decider
func SearchAttributesWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("SearchAttributes workflow started")

	// get search attributes that provided when start workflow
	info := workflow.GetInfo(ctx)
	val := info.SearchAttributes.IndexedFields["CustomIntField"]
	var currentIntValue int
	err := client.NewValue(val).Get(&currentIntValue)
	if err != nil {
		logger.Error("Get search attribute failed", zap.Error(err))
		return err
	}
	logger.Info("Current Search Attributes: ", zap.String("CustomIntField", strconv.Itoa(currentIntValue)))

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
	var listResults []*common.WorkflowExecutionInfo
	err = workflow.ExecuteActivity(ctx, ListExecutions, query).Get(ctx, &listResults)
	if err != nil {
		logger.Error("Failed to list workflow executions.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.String("Execution", listResults[0].String()))

	return nil
}

func printSearchAttributes(searchAttributes *common.SearchAttributes, logger *zap.Logger) error {
	buf := new(bytes.Buffer)
	for k, v := range searchAttributes.IndexedFields {
		var currentVal interface{}
		err := client.NewValue(v).Get(&currentVal)
		if err != nil {
			logger.Error(fmt.Sprintf("Get search attribute for key %s failed", k), zap.Error(err))
			return err
		}
		_, _ = fmt.Fprintf(buf, "%s=%v\n", k, currentVal)
	}
	logger.Info(fmt.Sprintf("Current Search Attributes: \n%s", buf.String()))
	return nil
}

func ListExecutions(ctx context.Context, query string) ([]*common.WorkflowExecutionInfo, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("List executions.", zap.String("Query", query))

	c, err := getClientFromContext(ctx)
	if err != nil {
		logger.Error("Error when get client")
		return nil, err
	}

	var executions []*common.WorkflowExecutionInfo
	var nextPageToken []byte
	for hasMore := true; hasMore; hasMore = len(nextPageToken) > 0 {
		resp, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Domain:        DomainName,
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
