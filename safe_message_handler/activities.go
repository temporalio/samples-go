package safe_message_handler

import (
	"context"
	"strconv"
	"time"

	"go.temporal.io/sdk/activity"
)

type (
	AssignNodesToJobInput struct {
		Nodes   []string
		JobName string
	}

	FindBadNodesInput struct {
		NodesToCheck map[string]struct{}
	}

	UnassignNodesForJobInput struct {
		Nodes   []string
		JobName string
	}
)

func AssignNodesToJobsActivity(ctx context.Context, input AssignNodesToJobInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Deallocating nodes to job", "nodes", input.Nodes, "job", input.JobName)
	time.Sleep(time.Millisecond * 1)
	return nil
}

func UnassignNodesForJobActivity(ctx context.Context, input UnassignNodesForJobInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Unassigning nodes from job", "nodes", input.Nodes, "job", input.JobName)
	time.Sleep(time.Millisecond * 1)
	return nil
}

func FindBadNodesActivity(ctx context.Context, input FindBadNodesInput) (map[string]struct{}, error) {
	time.Sleep(time.Millisecond * 1)
	logger := activity.GetLogger(ctx)
	badNodes := make(map[string]struct{})
	for node := range input.NodesToCheck {
		i, err := strconv.Atoi(node)
		if err != nil {
			logger.Error("Failed to convert node to int", "node", node)
			return nil, err
		}
		if i%5 == 0 {
			badNodes[node] = struct{}{}
		}
	}

	if len(badNodes) > 0 {
		logger.Info("Found bad node", "nodes", badNodes)
	} else {
		logger.Info("No new bad nodes found.")
	}
	return badNodes, nil
}
