package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go.temporal.io/temporal/activity"
	"strconv"
	"time"

	"github.com/temporalio/temporal-go-samples/cmd/samples/common"
	"go.temporal.io/temporal/.gen/go/shared"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This sample shows how to use search attributes. (Note this feature only work with ElasticSearch)
 */

// ApplicationName is the task list for this sample
const ApplicationName = "searchAttributesGroup"

// ClientKey is the key for lookup
type ClientKey int

const (
	// DomainName used for this sample
	DomainName = "samples-domain"
	// CadenceClientKey for retrieving cadence client from context
	CadenceClientKey ClientKey = iota
)

var (
	// ErrCadenceClientNotFound when cadence client is not found on context
	ErrCadenceClientNotFound = errors.New("failed to retrieve cadence client from context")
)

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	workflow.Register(SearchAttributesWorkflow)
	activity.Register(listExecutions)
}

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
	workflow.UpsertSearchAttributes(ctx, attributes)

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
	workflow.UpsertSearchAttributes(ctx, attributes)

	// print current search attributes
	info = workflow.GetInfo(ctx)
	err = printSearchAttributes(info.SearchAttributes, logger)
	if err != nil {
		return err
	}

	workflow.Sleep(ctx, 2*time.Second) // wait update reflected on ElasticSearch

	// list workflow
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 2 * time.Minute,
		StartToCloseTimeout:    2 * time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	query := "CustomIntField=2 and CustomKeywordField='Update2' order by CustomDatetimeField DESC"
	var listResults []*shared.WorkflowExecutionInfo
	err = workflow.ExecuteActivity(ctx, listExecutions, query).Get(ctx, &listResults)
	if err != nil {
		logger.Error("Failed to list workflow executions.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.String("Execution", listResults[0].String()))

	return nil
}

func printSearchAttributes(searchAttributes *shared.SearchAttributes, logger *zap.Logger) error {
	buf := new(bytes.Buffer)
	for k, v := range searchAttributes.IndexedFields {
		var currentVal interface{}
		err := client.NewValue(v).Get(&currentVal)
		if err != nil {
			logger.Error(fmt.Sprintf("Get search attribute for key %s failed", k), zap.Error(err))
			return err
		}
		fmt.Fprintf(buf, "%s=%v\n", k, currentVal)
	}
	logger.Info(fmt.Sprintf("Current Search Attributes: \n%s", buf.String()))
	return nil
}

func listExecutions(ctx context.Context, query string) ([]*shared.WorkflowExecutionInfo, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("List executions.", zap.String("Query", query))

	cadenceClient, err := getCadenceClientFromContext(ctx)
	if err != nil {
		logger.Error("Error when get cadence client")
		return nil, err
	}

	var executions []*shared.WorkflowExecutionInfo
	var nextPageToken []byte
	for hasMore := true; hasMore; hasMore = len(nextPageToken) > 0 {
		resp, err := cadenceClient.ListWorkflow(ctx, &shared.ListWorkflowExecutionsRequest{
			Domain:        common.StringPtr(DomainName),
			PageSize:      common.Int32Ptr(10),
			NextPageToken: nextPageToken,
			Query:         common.StringPtr(query),
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.Executions {
			executions = append(executions, r)
		}

		nextPageToken = resp.NextPageToken
		activity.RecordHeartbeat(ctx, nextPageToken)
	}

	return executions, nil
}

func getCadenceClientFromContext(ctx context.Context) (client.Client, error) {
	logger := activity.GetLogger(ctx)
	cadenceClient := ctx.Value(CadenceClientKey).(client.Client)
	if cadenceClient == nil {
		logger.Error("Could not retrieve cadence client from context.")
		return nil, ErrCadenceClientNotFound
	}

	return cadenceClient, nil
}
