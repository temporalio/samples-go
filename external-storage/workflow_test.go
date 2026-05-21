package externalstorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_ProcessOrderBatchWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	// Real activities, not mocks: generateOrders is deterministic per batchID
	// so the workflow's totals can be derived from the same source of truth.
	env.RegisterActivity(FetchOrders)
	env.RegisterActivity(ProcessOrders)

	request := OrderBatchRequest{BatchID: "BATCH-TEST", OrderCount: 10}
	env.ExecuteWorkflow(ProcessOrderBatchWorkflow, request)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var summary BatchSummary
	require.NoError(t, env.GetWorkflowResult(&summary))

	// Re-run the activity pipeline directly to compute the expected totals.
	// The workflow's BatchSummary must agree with this independent reduction.
	orders, err := FetchOrders(context.Background(), request)
	require.NoError(t, err)
	processed, err := ProcessOrders(context.Background(), orders)
	require.NoError(t, err)

	var expectedCost, expectedWeight float64
	var totalDays int
	for _, p := range processed {
		expectedCost += p.ShippingCostUSD
		expectedWeight += p.TotalWeightKg
		totalDays += p.EstimatedDeliveryDays
	}
	expectedAvg := float64(totalDays) / float64(len(processed))

	require.Equal(t, request.BatchID, summary.BatchID)
	require.Equal(t, request.OrderCount, summary.OrderCount)
	require.InDelta(t, expectedCost, summary.TotalShippingCostUSD, 0.01)
	require.InDelta(t, expectedWeight, summary.TotalWeightKg, 0.01)
	require.InDelta(t, expectedAvg, summary.AvgDeliveryDays, 0.1)
}
