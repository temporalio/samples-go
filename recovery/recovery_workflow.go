package recovery

import (
	"context"
	"errors"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal"
	eventpb "go.temporal.io/temporal-proto/event"
	executionpb "go.temporal.io/temporal-proto/execution"
	filterpb "go.temporal.io/temporal-proto/filter"
	"go.temporal.io/temporal-proto/workflowservice"
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/recovery/cache"
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
		ScheduleToStartTimeout: 10 * time.Minute,
		StartToCloseTimeout:    10 * time.Minute,
		HeartbeatTimeout:       time.Second * 30,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result ListOpenExecutionsResult
	err := workflow.ExecuteActivity(ctx, ListOpenExecutions, params.Type).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to list open workflow executions.", zap.Error(err))
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
	expiration := time.Duration(info.ExecutionStartToCloseTimeoutSeconds) * time.Second
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    10 * time.Second,
		ExpirationInterval: expiration,
		MaximumAttempts:    100,
	}
	ao = workflow.ActivityOptions{
		ScheduleToStartTimeout: expiration,
		StartToCloseTimeout:    expiration,
		HeartbeatTimeout:       time.Second * 30,
		RetryPolicy:            retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	doneCh := workflow.NewChannel(ctx)
	for i := 0; i < concurrency; i++ {
		startIndex := i * batchSize

		workflow.Go(ctx, func(ctx workflow.Context) {
			err = workflow.ExecuteActivity(ctx, RecoverExecutions, result.ID, startIndex, batchSize).Get(ctx, nil)
			if err != nil {
				logger.Error("Recover executions failed.", zap.Int("StartIndex", startIndex), zap.Error(err))
			} else {
				logger.Info("Recover executions completed.", zap.Int("StartIndex", startIndex))
			}

			doneCh.Send(ctx, "done")
		})
	}

	for i := 0; i < concurrency; i++ {
		doneCh.Receive(ctx, nil)
	}

	logger.Info("Workflow completed.", zap.Int("Result", result.Count))

	return nil
}

func ListOpenExecutions(ctx context.Context, workflowType string) (*ListOpenExecutionsResult, error) {
	key := uuid.New()
	logger := activity.GetLogger(ctx)
	logger.Info("List all open executions of type.",
		zap.String("WorkflowType", workflowType),
		zap.String("HostID", HostID))

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
		zap.String("HostID", HostID),
		zap.String("Key", key),
		zap.Int("StartIndex", startIndex),
		zap.Int("BatchSize", batchSize))

	executionsCache := ctx.Value(WorkflowExecutionCacheKey).(cache.Cache)
	if executionsCache == nil {
		logger.Error("Could not retrieve cache from context.")
		return ErrExecutionCacheNotFound
	}

	openExecutions := executionsCache.Get(key).([]*executionpb.WorkflowExecution)
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
				zap.String("WorkflowID", execution.GetWorkflowId()),
				zap.Error(err))
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

	execution := &executionpb.WorkflowExecution{
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
		zap.String("WorkflowID", execution.GetWorkflowId()),
		zap.String("NewRunID", newRun.GetRunID()))

	return nil
}

func extractStateFromEvent(workflowID string, event *eventpb.HistoryEvent) (*RestartParams, error) {
	switch event.GetEventType() {
	case eventpb.EventType_WorkflowExecutionStarted:
		attr := event.GetWorkflowExecutionStartedEventAttributes()
		state, err := deserializeUserState(attr.Input)
		if err != nil {
			// Corrupted Workflow Execution State
			return nil, err
		}
		return &RestartParams{
			Options: client.StartWorkflowOptions{
				ID:                              workflowID,
				TaskList:                        attr.TaskList.GetName(),
				ExecutionStartToCloseTimeout:    time.Second * time.Duration(attr.GetExecutionStartToCloseTimeoutSeconds()),
				DecisionTaskStartToCloseTimeout: time.Second * time.Duration(attr.GetTaskStartToCloseTimeoutSeconds()),
				// RetryPolicy: attr.RetryPolicy,
			},
			State: state,
		}, nil
	default:
		return nil, errors.New("unknown event type")
	}
}

func extractSignals(events []*eventpb.HistoryEvent) ([]*SignalParams, error) {
	var signals []*SignalParams
	for _, event := range events {
		if event.GetEventType() == eventpb.EventType_WorkflowExecutionSignaled {
			attr := event.GetWorkflowExecutionSignaledEventAttributes()
			if attr.GetSignalName() == TripSignalName && attr.Input != nil && len(attr.Input) > 0 {
				signalData, err := deserializeTripEvent(attr.Input)
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

func isExecutionCompleted(event *eventpb.HistoryEvent) bool {
	switch event.GetEventType() {
	case eventpb.EventType_WorkflowExecutionCompleted, eventpb.EventType_WorkflowExecutionTerminated,
		eventpb.EventType_WorkflowExecutionCanceled, eventpb.EventType_WorkflowExecutionFailed,
		eventpb.EventType_WorkflowExecutionTimedOut:
		return true
	default:
		return false
	}
}

func getAllExecutionsOfType(ctx context.Context, c client.Client, workflowType string) ([]*executionpb.WorkflowExecution, error) {
	var openExecutions []*executionpb.WorkflowExecution
	var nextPageToken []byte
	for hasMore := true; hasMore; hasMore = len(nextPageToken) > 0 {
		resp, err := c.ListOpenWorkflow(ctx, &workflowservice.ListOpenWorkflowExecutionsRequest{
			Namespace:       client.DefaultNamespace,
			MaximumPageSize: 10,
			NextPageToken:   nextPageToken,
			StartTimeFilter: &filterpb.StartTimeFilter{
				EarliestTime: 0,
				LatestTime:   time.Now().UnixNano(),
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

func getHistory(ctx context.Context, execution *executionpb.WorkflowExecution) ([]*eventpb.HistoryEvent, error) {
	c, err := getClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	iter := c.GetWorkflowHistory(ctx, execution.GetWorkflowId(), execution.GetRunId(), false, filterpb.HistoryEventFilterType_AllEvent)
	var events []*eventpb.HistoryEvent
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
