package safe_message_handler

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	StartCluster      = "StartCluster"
	ShutdownCluster   = "ShutdownCluster"
	AssignNodesToJobs = "AssignNodesToJobs"
	DeleteJob         = "DeleteJob"
)

type (
	// In workflows that continue-as-new, it's convenient to store all your state in one serializable structure
	// to make it easier to pass between runs
	ClusterManagerState struct {
		ClusterStarted  bool
		ClusterShutdown bool
		Nodes           map[string]string
		JobsAssigned    map[string]struct{}
	}

	ClusterManagerInput struct {
		State             *ClusterManagerState
		TestContinueAsNew bool
	}

	ClusterManagerResult struct {
		NumCurrentlyAssignedNodes int
		NumBadNodes               int
	}

	// Be in the habit of storing message inputs and outputs in serializable structures.
	// This makes it easier to add more over time in a backward-compatible way.
	ClusterManagerAssignNodesToJobInput struct {
		// If larger or smaller than previous amounts, will resize the job.
		TotalNumNodes int
		JobName       string
	}

	ClusterManagerDeleteJobInput struct {
		JobName string
	}

	ClusterManagerAssignNodesToJobResult struct {
		NodesAssigned map[string]struct{}
	}

	ClusterManager struct {
		state             ClusterManagerState
		nodeLock          workflow.Mutex
		logger            log.Logger
		sleepInterval     time.Duration
		maxHistoryLength  int
		startCh           workflow.ReceiveChannel
		shutdownCh        workflow.ReceiveChannel
		testContinueAsNew bool
	}
)

func newClusterManager(ctx workflow.Context, wfInput ClusterManagerInput) (*ClusterManager, error) {
	logger := workflow.GetLogger(ctx)
	sleepInterval := time.Second * 600
	maxHistoryLength := 0
	nodeLock := workflow.NewMutex(ctx)

	state := ClusterManagerState{
		Nodes:        make(map[string]string),
		JobsAssigned: make(map[string]struct{}),
	}
	if wfInput.State != nil {
		state = *wfInput.State
	}
	if wfInput.TestContinueAsNew {
		maxHistoryLength = 120
		sleepInterval = time.Second * 1
	}

	startCh := workflow.GetSignalChannel(ctx, StartCluster)
	shutdownCh := workflow.GetSignalChannel(ctx, ShutdownCluster)

	state.Nodes = make(map[string]string)
	for i := range 25 {
		state.Nodes[fmt.Sprint(i)] = ""
	}

	cm := &ClusterManager{
		state:             state,
		nodeLock:          nodeLock,
		logger:            logger,
		startCh:           startCh,
		shutdownCh:        shutdownCh,
		sleepInterval:     sleepInterval,
		maxHistoryLength:  maxHistoryLength,
		testContinueAsNew: wfInput.TestContinueAsNew,
	}

	err := workflow.SetUpdateHandler(ctx, AssignNodesToJobs, cm.AssignNodesToJobs)
	if err != nil {
		return nil, err
	}

	err = workflow.SetUpdateHandler(ctx, DeleteJob, cm.DeleteJob)
	if err != nil {
		return nil, err
	}
	return cm, nil
}

func (cm *ClusterManager) badNodes() map[string]struct{} {
	badNodes := make(map[string]struct{})
	for _, k := range workflow.DeterministicKeys(cm.state.Nodes) {
		if cm.state.Nodes[k] == "BAD!" {
			badNodes[k] = struct{}{}
		}
	}
	return badNodes
}

// This is an update as opposed to a signal because the client may want to wait for nodes to be allocated
// before sending work to those nodes.
// Returns the list of node names that were allocated to the job.
func (cm *ClusterManager) AssignNodesToJobs(ctx workflow.Context, input ClusterManagerAssignNodesToJobInput) (ClusterManagerAssignNodesToJobResult, error) {
	err := workflow.Await(ctx, func() bool {
		return cm.state.ClusterStarted
	})
	if err != nil {
		return ClusterManagerAssignNodesToJobResult{}, err
	}
	if cm.state.ClusterShutdown {
		// If you want the client to receive a error, either add an update validator and return the
		// error from there, or return an error.
		return ClusterManagerAssignNodesToJobResult{}, errors.New("cannot assign nodes to a job: Cluster is already shut down")
	}
	err = cm.nodeLock.Lock(ctx)
	if err != nil {
		return ClusterManagerAssignNodesToJobResult{}, err
	}
	defer cm.nodeLock.Unlock()
	// Idempotency guard.
	if _, ok := cm.state.JobsAssigned[input.JobName]; ok {
		return ClusterManagerAssignNodesToJobResult{
			NodesAssigned: cm.getAssignedNodes(&input.JobName),
		}, nil
	}
	unassignedNodes := cm.getUnassignedNodes()
	if len(unassignedNodes) < input.TotalNumNodes {
		return ClusterManagerAssignNodesToJobResult{}, errors.New("not enough nodes to assign to job")
	}
	nodesToAssign := unassignedNodes[:input.TotalNumNodes]

	// Assign the nodes to the job.
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 10,
	})
	// This get would be dangerous without holding nodeLock because it yields control and allows interleaving
	// with DeleteJob and performHealthChecks, which both modify cm.state.nodes.
	err = workflow.ExecuteActivity(activityCtx, AssignNodesToJobsActivity, AssignNodesToJobInput{
		Nodes:   nodesToAssign,
		JobName: input.JobName,
	}).Get(activityCtx, nil)
	if err != nil {
		return ClusterManagerAssignNodesToJobResult{}, err
	}
	for _, node := range nodesToAssign {
		cm.state.Nodes[node] = input.JobName
	}
	cm.state.JobsAssigned[input.JobName] = struct{}{}

	return ClusterManagerAssignNodesToJobResult{
		NodesAssigned: cm.getAssignedNodes(&input.JobName),
	}, nil
}

// Even though it returns nothing, this is an update because the client may want to track it, for example
// to wait for nodes to be unassigned before reassigning them.
func (cm *ClusterManager) DeleteJob(ctx workflow.Context, input ClusterManagerDeleteJobInput) error {
	err := workflow.Await(ctx, func() bool {
		return cm.state.ClusterStarted
	})
	if err != nil {
		return err
	}
	if cm.state.ClusterShutdown {
		// If you want the client to receive a error, either add an update validator and return the
		// error from there, or return an error.
		return errors.New("cannot delete a job: Cluster is already shut down")
	}
	err = cm.nodeLock.Lock(ctx)
	if err != nil {
		return err
	}
	defer cm.nodeLock.Unlock()

	nodesToUnassign := make([]string, 0)
	for _, k := range workflow.DeterministicKeys(cm.state.Nodes) {
		if cm.state.Nodes[k] == input.JobName {
			nodesToUnassign = append(nodesToUnassign, k)
		}
	}
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Second * 10,
	})
	err = workflow.ExecuteActivity(activityCtx, UnassignNodesForJobActivity, UnassignNodesForJobInput{
		Nodes:   nodesToUnassign,
		JobName: input.JobName,
	}).Get(activityCtx, nil)
	if err != nil {
		return err
	}
	for _, node := range nodesToUnassign {
		cm.state.Nodes[node] = ""
	}
	return nil
}

func (cm *ClusterManager) getUnassignedNodes() []string {
	var unassignedNodes []string
	for _, k := range workflow.DeterministicKeys(cm.state.Nodes) {
		if cm.state.Nodes[k] == "" {
			unassignedNodes = append(unassignedNodes, k)
		}
	}
	return unassignedNodes

}

func (cm *ClusterManager) performHealthCheck(ctx workflow.Context) {
	err := cm.nodeLock.Lock(ctx)
	if err != nil {
		cm.logger.Error("Failed to acquire lock", "error", err)
		return
	}
	defer cm.nodeLock.Unlock()
	assignedNodes := cm.getAssignedNodes(nil)
	// Do some activity on the nodes
	var badNodes map[string]struct{}
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Second * 10,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	})
	err = workflow.ExecuteActivity(activityCtx, FindBadNodesActivity, assignedNodes).Get(activityCtx, &badNodes)
	if err != nil {
		cm.logger.Error("Health check failed", "error", err)
	}
	for _, node := range workflow.DeterministicKeys(badNodes) {
		cm.state.Nodes[node] = "BAD!"
	}
}

func (cm *ClusterManager) getAssignedNodes(jobName *string) map[string]struct{} {
	assignedNodes := make(map[string]struct{})
	if jobName != nil {
		for _, k := range workflow.DeterministicKeys(cm.state.Nodes) {
			if cm.state.Nodes[k] == *jobName {
				assignedNodes[k] = struct{}{}
			}
		}
	} else {
		for _, k := range workflow.DeterministicKeys(cm.state.Nodes) {
			if cm.state.Nodes[k] != "BAD!" {
				assignedNodes[k] = struct{}{}
			}
		}
	}
	return assignedNodes
}

func (cm *ClusterManager) shouldContinueAsNew(ctx workflow.Context) bool {
	if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
		return true
	}
	if cm.maxHistoryLength > 0 && workflow.GetInfo(ctx).GetCurrentHistoryLength() > cm.maxHistoryLength {
		return true
	}
	return false
}

func (cm *ClusterManager) run(ctx workflow.Context) (ClusterManagerResult, error) {
	// Wait for the start signal.
	cm.startCh.Receive(ctx, nil)
	cm.state.ClusterStarted = true
	cm.logger.Info("Cluster started")
	for {
		selector := workflow.NewSelector(ctx)
		shouldShutdown := false
		selector.AddReceive(cm.shutdownCh, func(c workflow.ReceiveChannel, _ bool) {
			c.Receive(ctx, nil)
			shouldShutdown = true
		})
		selector.AddFuture(workflow.NewTimer(ctx, cm.sleepInterval), func(f workflow.Future) {
			cm.performHealthCheck(ctx)
		})
		selector.Select(ctx)
		if shouldShutdown {
			break
		}
		if cm.shouldContinueAsNew(ctx) {
			// We don't want to leave any job assignment or deletion handlers half-finished when we continue as new.
			err := workflow.Await(ctx, func() bool {
				return workflow.AllHandlersFinished(ctx)
			})
			if err != nil {
				cm.logger.Error("Failed to wait for all handlers to finish", "error", err)
				return ClusterManagerResult{}, err
			}
			cm.logger.Info("Continuing as new")
			return ClusterManagerResult{}, workflow.NewContinueAsNewError(ctx, ClusterManagerInput{
				State:             &cm.state,
				TestContinueAsNew: cm.testContinueAsNew,
			})
		}

	}
	// Make sure we finish off handlers such as deleting jobs before we complete the workflow.
	err := workflow.Await(ctx, func() bool {
		return workflow.AllHandlersFinished(ctx)
	})
	if err != nil {
		cm.logger.Error("Failed to wait for all handlers to finish", "error", err)
		return ClusterManagerResult{}, err
	}
	return ClusterManagerResult{
		NumCurrentlyAssignedNodes: len(cm.getAssignedNodes(nil)),
		NumBadNodes:               len(cm.badNodes()),
	}, nil
}

// ClusterManagerWorkflow keeps track of the assignments of a cluster of nodes.
// Via signals, the cluster can be started and shutdown.
// Via updates, clients can also assign jobs to nodes and delete jobs.
// These updates must run atomically.
func ClusterManagerWorkflow(ctx workflow.Context, wfInput ClusterManagerInput) (ClusterManagerResult, error) {
	cm, err := newClusterManager(ctx, wfInput)
	if err != nil {
		return ClusterManagerResult{}, err
	}
	return cm.run(ctx)
}
