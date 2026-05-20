package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	externalstorage "github.com/temporalio/samples-go/external-storage"
	"go.temporal.io/sdk/client"
)

func main() {
	var orderCount int
	flag.IntVar(&orderCount, "orders", 200, "Number of orders in the batch")
	flag.Parse()

	c, err := externalstorage.NewClient(context.Background(), client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	runID := time.Now().Format("20060102-150405")
	workflowID := "external-storage-" + runID
	request := externalstorage.OrderBatchRequest{
		BatchID:    "BATCH-" + runID,
		OrderCount: orderCount,
	}

	fmt.Printf("Starting workflow %s (batch_id=%s, order_count=%d)\n", workflowID, request.BatchID, request.OrderCount)

	we, err := c.ExecuteWorkflow(context.Background(),
		client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: externalstorage.TaskQueue,
		},
		externalstorage.ProcessOrderBatchWorkflow,
		request,
	)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	var summary externalstorage.BatchSummary
	if err := we.Get(context.Background(), &summary); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}

	fmt.Printf("\nBatch %s: %d orders processed\n", summary.BatchID, summary.OrderCount)
	fmt.Printf("  Total shipping cost: $%.2f\n", summary.TotalShippingCostUSD)
	fmt.Printf("  Total weight:        %.1f kg\n", summary.TotalWeightKg)
	fmt.Printf("  Avg delivery:        %.1f days\n", summary.AvgDeliveryDays)
}
