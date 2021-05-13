package recovery

import (
	"context"
	"errors"
	"time"

	"github.com/pborman/uuid"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	filterpb "go.temporal.io/api/filter/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/recovery/cache"
)

type (
	// Params is the input parameters to RecoveryWorkflow
	Params struct {
		ID          string
		Type        string
		Concurrency int
	}

	// ListOpenExecutionsResult is the result returned from listOpenExecutions activity
	ListOpenExecutionsResult struct {
		ID     string
		Count  int
		HostID string
	}

	// RestartParams are parameters extracted from StartWorkflowExecution history event
	RestartParams struct {
		Options client.StartWorkflowOptions
		State   UserState
	}

	// SignalParams are the parameters extracted from SignalWorkflowExecution history event
	SignalParams struct {
		Name string
		Data TripEvent
	}
)

// ClientKey is the key for lookup
type ClientKey int

const (
	// TemporalClientKey for retrieving client from context
	TemporalClientKey ClientKey = iota
	// WorkflowExecutionCacheKey for retrieving executions cache from context
	WorkflowExecutionCacheKey
)

// HostID - Use a new uuid just for demo so we can run 2 host specific activity workers on same machine.
// In real world case, you would use a hostname or ip address as HostID.
var HostID = "recovery_" + uuid.New()

var (
	// ErrClientNotFound when client is not found on context
	ErrClientNotFound = errors.New("failed to retrieve client from context")
	// ErrExecutionCacheNotFound when executions cache is not found on context
	ErrExecutionCacheNotFound = errors.New("failed to retrieve cache from context")
)

// RecoverWorkflow is the workflow implementation to recover TripWorkflow executions
func RecoverWorkflow(ctx workflow.Context, params Params) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Recover workflow started.")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result ListOpenExecutionsResult
	err := workflow.ExecuteActivity(ctx, ListOpenExecutions, params.Type).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to list open workflow executions.", "Error", err)
		return err
	}

	concurrency := 1
	if params.Concurrency > 0 {
		concurrency = params.Concurrency
	}

	if result.Count < concurrency {
		concurrency = result.Count
	}

	batchSize := result.Count / concurrency
	if result.Count%concurrency != 0 {
		batchSize++
	}

	// Setup retry policy for recovery activity
	info := workflow.GetInfo(ctx)
	expiration := info.WorkflowExecutionTimeout
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    10 * time.Second,
		MaximumAttempts:    100,
	}
	ao = workflow.ActivityOptions{
		StartToCloseTimeout: expiration,
		HeartbeatTimeout:    30 * time.Second,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	doneCh := workflow.NewChannel(ctx)
	for i := 0; i < concurrency; i++ {
		startIndex := i * batchSize

		workflow.Go(ctx, func(ctx workflow.Context) {
			err = workflow.ExecuteActivity(ctx, RecoverExecutions, result.ID, startIndex, batchSize).Get(ctx, nil)
			if err != nil {
				logger.Error("Recover executions failed.", "StartIndex", startIndex, "Error", err)
			} else {
				logger.Info("Recover executions completed.", "StartIndex", startIndex)
			}

			doneCh.Send(ctx, "done")
		})
	}

	for i := 0; i < concurrency; i++ {
		doneCh.Receive(ctx, nil)
	}

	logger.Info("Workflow completed.", "Result", result.Count)

	return nil
}

func ListOpenExecutions(ctx context.Context, workflowType string) (*ListOpenExecutionsResult, error) {
	key := uuid.New()
	logger := activity.GetLogger(ctx)
	logger.Info("List all open executions of type.",
		"WorkflowType", workflowType,
		"HostID", HostID)

	c, err := getClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	executionsCache := ctx.Value(WorkflowExecutionCacheKey).(cache.Cache)
	if executionsCache == nil {
		logger.Error("Could not retrieve cache from context.")
		return nil, ErrExecutionCacheNotFound
	}

	openExecutions, err := getAllExecutionsOfType(ctx, c, workflowType)
	if err != nil {
		return nil, err
	}

	executionsCache.Put(key, openExecutions)
	return &ListOpenExecutionsResult{
		ID:     key,
		Count:  len(openExecutions),
		HostID: HostID,
	}, nil
}

func RecoverExecutions(ctx context.Context, key string, startIndex, batchSize int) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting execution recovery.",
		"HostID", HostID,
		"Key", key,
		"StartIndex", startIndex,
		"BatchSize", batchSize)

	executionsCache := ctx.Value(WorkflowExecutionCacheKey).(cache.Cache)
	if executionsCache == nil {
		logger.Error("Could not retrieve cache from context.")
		return ErrExecutionCacheNotFound
	}

	openExecutions := executionsCache.Get(key).([]*commonpb.WorkflowExecution)
	endIndex := startIndex + batchSize

	// Check if this activity has previous heartbeat to retrieve progress from it
	if activity.HasHeartbeatDetails(ctx) {
		var finishedIndex int
		if err := activity.GetHeartbeatDetails(ctx, &finishedIndex); err == nil {
			// we have finished progress
			startIndex = finishedIndex + 1 // start from next one.
		}
	}

	for index := startIndex; index < endIndex && index < len(openExecutions); index++ {
		execution := openExecutions[index]
		if err := recoverSingleExecution(ctx, execution.GetWorkflowId()); err != nil {
			logger.Error("Failed to recover execution.",
				"WorkflowID", execution.GetWorkflowId(),
				"Error", err)
			return err
		}

		// Record a heartbeat after each recovery of execution
		activity.RecordHeartbeat(ctx, index)
	}

	return nil
}

func recoverSingleExecution(ctx context.Context, workflowID string) error {
	logger := activity.GetLogger(ctx)
	c, err := getClientFromContext(ctx)
	if err != nil {
		return err
	}

	execution := &commonpb.WorkflowExecution{
		WorkflowId: workflowID,
	}
	history, err := getHistory(ctx, execution)
	if err != nil {
		return err
	}

	if len(history) == 0 {
		// Nothing to recover
		return nil
	}

	firstEvent := history[0]
	lastEvent := history[len(history)-1]

	// Extract information from StartWorkflowExecution parameters so we can start a new run
	params, err := extractStateFromEvent(workflowID, firstEvent)
	if err != nil {
		return err
	}

	// Parse the entire history and extract all signals so they can be replayed back to new run
	signals, err := extractSignals(history)
	if err != nil {
		return err
	}

	// First terminate existing run if already running
	if !isExecutionCompleted(lastEvent) {
		err := c.TerminateWorkflow(ctx, execution.GetWorkflowId(), execution.GetRunId(), "Recover", nil)
		if err != nil {
			return err
		}
	}

	// Start new execution run
	newRun, err := c.ExecuteWorkflow(ctx, params.Options, "TripWorkflow", params.State)
	if err != nil {
		return err
	}

	// re-inject all signals to new run
	for _, s := range signals {
		_ = c.SignalWorkflow(ctx, execution.GetWorkflowId(), newRun.GetRunID(), s.Name, s.Data)
	}

	logger.Info("Successfully restarted workflow.",
		"WorkflowID", execution.GetWorkflowId(),
		"NewRunID", newRun.GetRunID())

	return nil
}

func extractStateFromEvent(workflowID string, event *historypb.HistoryEvent) (*RestartParams, error) {
	switch event.GetEventType() {
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
		attr := event.GetWorkflowExecutionStartedEventAttributes()
		state, err := deserializeUserState(attr.GetInput())
		if err != nil {
			// Corrupted Workflow Execution State
			return nil, err
		}
		return &RestartParams{
			Options: client.StartWorkflowOptions{
				ID:                  workflowID,
				TaskQueue:           attr.TaskQueue.GetName(),
				WorkflowTaskTimeout: *attr.GetWorkflowTaskTimeout(),
				// RetryPolicy: attr.RetryPolicy,
			},
			State: state,
		}, nil
	default:
		return nil, errors.New("unknown event type")
	}
}

func extractSignals(events []*historypb.HistoryEvent) ([]*SignalParams, error) {
	var signals []*SignalParams
	for _, event := range events {
		if event.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED {
			attr := event.GetWorkflowExecutionSignaledEventAttributes()
			if attr.GetSignalName() == TripSignalName && attr.GetInput() != nil {
				signalData, err := deserializeTripEvent(attr.GetInput())
				if err != nil {
					// Corrupted Signal Payload
					return nil, err
				}

				signal := &SignalParams{
					Name: attr.GetSignalName(),
					Data: signalData,
				}
				signals = append(signals, signal)
			}
		}
	}

	return signals, nil
}

func isExecutionCompleted(event *historypb.HistoryEvent) bool {
	switch event.GetEventType() {
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED, enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED,
		enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED, enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
		enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		return true
	default:
		return false
	}
}

func getAllExecutionsOfType(ctx context.Context, c client.Client, workflowType string) ([]*commonpb.WorkflowExecution, error) {
	var openExecutions []*commonpb.WorkflowExecution
	var nextPageToken []byte
	for hasMore := true; hasMore; hasMore = len(nextPageToken) > 0 {
		zeroTime := time.Time{}
		now := time.Now()
		resp, err := c.ListOpenWorkflow(ctx, &workflowservice.ListOpenWorkflowExecutionsRequest{
			Namespace:       client.DefaultNamespace,
			MaximumPageSize: 10,
			NextPageToken:   nextPageToken,
			StartTimeFilter: &filterpb.StartTimeFilter{
				EarliestTime: &zeroTime,
				LatestTime:   &now,
			},
			Filters: &workflowservice.ListOpenWorkflowExecutionsRequest_TypeFilter{TypeFilter: &filterpb.WorkflowTypeFilter{
				Name: workflowType,
			}},
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.Executions {
			openExecutions = append(openExecutions, r.Execution)
		}

		nextPageToken = resp.NextPageToken
		activity.RecordHeartbeat(ctx, nextPageToken)
	}

	return openExecutions, nil
}

func getHistory(ctx context.Context, execution *commonpb.WorkflowExecution) ([]*historypb.HistoryEvent, error) {
	c, err := getClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	iter := c.GetWorkflowHistory(ctx, execution.GetWorkflowId(), execution.GetRunId(), false, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	var events []*historypb.HistoryEvent
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil
}

func getClientFromContext(ctx context.Context) (client.Client, error) {
	logger := activity.GetLogger(ctx)
	temporalClient := ctx.Value(TemporalClientKey).(client.Client)
	if temporalClient == nil {
		logger.Error("Could not retrieve temporal client from context.")
		return nil, ErrClientNotFound
	}

	return temporalClient, nil
}
