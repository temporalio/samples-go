package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/queue"
)

// define command line parameters
var (
	testMode = flag.String("test", "basic", "test mode: basic=basic competition test, query=query function test, pool=inspect internal resource pool")
	poolID   = flag.String("poolid", "", "specify the resource pool ID to query, format: 'resource-pool:{namespace}:{resourceID}'")
)

func main() {
	// parse command line parameters
	flag.Parse()

	// create temporal client
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalln("can't create client", err)
	}
	defer c.Close()

	// execute different tests based on test mode
	switch *testMode {
	case "basic":
		runBasicTest(c)
	case "query":
		runQueryTest(c)
	case "pool":
		if *poolID == "" {
			log.Fatalln("when using pool mode, you must specify the resource pool ID through the -poolid parameter")
		}
		runPoolInspectorTest(c, *poolID)
	default:
		log.Fatalln("unknown test mode:", *testMode)
	}
}

// runBasicTest 运行基本的资源竞争测试
func runBasicTest(c client.Client) {
	// 生成一个随机的资源ID，此ID可以是业务逻辑标识符
	resourceID := "resource-" + uuid.New()

	// 准备资源池相关信息
	resourcePoolWorkflowID := "resource-pool:ResourcePoolDemo:" + resourceID

	// start multiple workflow instances that will compete for the same resource
	// each workflow will attempt to acquire the resource, complete work, and then release the resource

	// configure first workflow
	workflow1ID := "QueueWorkflow1_" + uuid.New()
	workflow1Options := client.StartWorkflowOptions{
		ID:        workflow1ID,
		TaskQueue: "queue-sample",
	}

	// configure second workflow
	workflow2ID := "QueueWorkflow2_" + uuid.New()
	workflow2Options := client.StartWorkflowOptions{
		ID:        workflow2ID,
		TaskQueue: "queue-sample",
	}

	// configure third workflow
	workflow3ID := "QueueWorkflow3_" + uuid.New()
	workflow3Options := client.StartWorkflowOptions{
		ID:        workflow3ID,
		TaskQueue: "queue-sample",
	}

	// configure fourth workflow
	workflow4ID := "QueueWorkflow4_" + uuid.New()
	workflow4Options := client.StartWorkflowOptions{
		ID:        workflow4ID,
		TaskQueue: "queue-sample",
	}

	// start first workflow (processing time 5 seconds, cancelable)
	we, err := c.ExecuteWorkflow(context.Background(), workflow1Options,
		queue.SampleWorkflowWithResourcePool, resourceID, 5*time.Second, true)
	if err != nil {
		log.Fatalln("can't execute workflow1", err)
	}
	log.Println("workflow1 started", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// start second workflow (processing time 7 seconds, cancelable)
	we, err = c.ExecuteWorkflow(context.Background(), workflow2Options,
		queue.SampleWorkflowWithResourcePool, resourceID, 7*time.Second, true)
	if err != nil {
		log.Fatalln("failed to execute workflow2", err)
	}
	log.Println("workflow2 started", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// start third workflow (processing time 3 seconds, cancelable)
	we, err = c.ExecuteWorkflow(context.Background(), workflow3Options,
		queue.SampleWorkflowWithResourcePool, resourceID, 3*time.Second, true)
	if err != nil {
		log.Fatalln("failed to execute workflow3", err)
	}
	log.Println("workflow3 started", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// wait 2 seconds, there should be one workflow in progress and two waiting workflows in the resource pool
	time.Sleep(2 * time.Second)

	// execute the following operations in a separate goroutine to avoid blocking the main process
	go func() {
		// wait 3 seconds, one of the first or second workflows should have acquired the resource and started processing
		time.Sleep(3 * time.Second)

		// try to cancel the resource request of the third workflow
		cancelWorkflow, err := c.ExecuteWorkflow(context.Background(),
			client.StartWorkflowOptions{
				ID:        "CancelResourceRequest_" + uuid.New(),
				TaskQueue: "queue-sample",
			},
			cancelResourceRequest, resourcePoolWorkflowID, workflow3ID)
		if err != nil {
			log.Println("failed to cancel resource request workflow", err)
		} else {
			log.Println("cancel resource request workflow started", cancelWorkflow.GetID())
		}

		// wait 2 seconds and then increase the resource pool size
		time.Sleep(2 * time.Second)
		updateResourcePoolWorkflow, err := c.ExecuteWorkflow(context.Background(),
			client.StartWorkflowOptions{
				ID:        "UpdateResourcePool_" + uuid.New(),
				TaskQueue: "queue-sample",
			},
			updateResourcePoolSize, resourcePoolWorkflowID, 3) // 增加到3个资源
		if err != nil {
			log.Println("failed to update resource pool size workflow", err)
		} else {
			log.Println("update resource pool size workflow started", updateResourcePoolWorkflow.GetID())
		}

		// wait 2 seconds and then start the fourth workflow
		time.Sleep(2 * time.Second)
		we, err = c.ExecuteWorkflow(context.Background(), workflow4Options,
			queue.SampleWorkflowWithResourcePool, resourceID, 4*time.Second, true)
		if err != nil {
			log.Println("failed to execute workflow4", err)
		} else {
			log.Println("workflow4 started", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
		}

		// wait 5 seconds and then reduce the resource pool size
		time.Sleep(5 * time.Second)
		updateResourcePoolWorkflow, err = c.ExecuteWorkflow(context.Background(),
			client.StartWorkflowOptions{
				ID:        "UpdateResourcePool2_" + uuid.New(),
				TaskQueue: "queue-sample",
			},
			updateResourcePoolSize, resourcePoolWorkflowID, 1) // reduce to 1 resource
		if err != nil {
			log.Println("failed to update resource pool size workflow", err)
		} else {
			log.Println("update resource pool size workflow started", updateResourcePoolWorkflow.GetID())
		}
	}()

	log.Println("all workflows have been started, they will compete for the same resource. please check the workflow logs for resource competition details.")

	// main thread waits long enough for all examples to complete
	time.Sleep(60 * time.Second)
}

// runQueryTest 运行资源池查询测试
func runQueryTest(c client.Client) {
	// 生成一个随机的资源ID，此ID可以是业务逻辑标识符
	resourceID := "resource-" + uuid.New()

	// 准备资源池相关信息
	resourcePoolWorkflowID := "resource-pool:ResourcePoolDemo:" + resourceID

	// start multiple workflow instances that will compete for the same resource
	log.Println("start resource pool query test, resource ID:", resourceID)
	log.Println("resource pool workflow ID:", resourcePoolWorkflowID)

	// start 5 workflows that will compete for the resource
	workflowIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		// configure workflow
		workflowID := fmt.Sprintf("QueryTestWorkflow%d_%s", i+1, uuid.New())
		workflowIDs[i] = workflowID
		workflowOptions := client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: "queue-sample",
		}

		// processing time increases with index
		processingTime := time.Duration(i+5) * time.Second

		// start workflow
		we, err := c.ExecuteWorkflow(context.Background(), workflowOptions,
			queue.SampleWorkflowWithResourcePool, resourceID, processingTime, true)
		if err != nil {
			log.Fatalln("failed to execute workflow", i+1, err)
		}
		log.Printf("workflow%d started, WorkflowID: %s, RunID: %s", i+1, we.GetID(), we.GetRunID())

		// wait a short time after starting the workflow to better observe the query results
		time.Sleep(500 * time.Millisecond)
	}

	// wait 2 seconds, there should be one workflow in progress and four waiting workflows in the resource pool
	time.Sleep(2 * time.Second)

	// execute first query - initial state, default resource pool size is 1
	queryAndPrintResourcePool(c, resourcePoolWorkflowID, "initial state")

	// update resource pool size to 3
	updatePoolWorkflow, err := c.ExecuteWorkflow(context.Background(),
		client.StartWorkflowOptions{
			ID:        "UpdateResourcePoolForQuery_" + uuid.New(),
			TaskQueue: "queue-sample",
		},
		updateResourcePoolSize, resourcePoolWorkflowID, 3)
	if err != nil {
		log.Println("failed to update resource pool size workflow", err)
	} else {
		log.Println("update resource pool size workflow started", updatePoolWorkflow.GetID())
	}

	// wait 3 seconds for the resource pool to update
	time.Sleep(3 * time.Second)

	// execute second query - resource pool size is 3
	queryAndPrintResourcePool(c, resourcePoolWorkflowID, "expanded to 3 resources")

	// try to cancel the resource request of the fifth workflow
	cancelWorkflow, err := c.ExecuteWorkflow(context.Background(),
		client.StartWorkflowOptions{
			ID:        "CancelResourceRequestForQuery_" + uuid.New(),
			TaskQueue: "queue-sample",
		},
		cancelResourceRequest, resourcePoolWorkflowID, workflowIDs[4])
	if err != nil {
		log.Println("failed to cancel resource request workflow", err)
	} else {
		log.Println("cancel resource request workflow started", cancelWorkflow.GetID())
	}

	// wait 2 seconds
	time.Sleep(2 * time.Second)

	// execute third query - after canceling one request
	queryAndPrintResourcePool(c, resourcePoolWorkflowID, "after canceling one request")

	// wait a while for some workflows to complete
	time.Sleep(10 * time.Second)

	// execute fourth query - after some workflows have completed
	queryAndPrintResourcePool(c, resourcePoolWorkflowID, "after some workflows have completed")

	// update resource pool size to 1
	updatePoolWorkflow, err = c.ExecuteWorkflow(context.Background(),
		client.StartWorkflowOptions{
			ID:        "ShrinkResourcePoolForQuery_" + uuid.New(),
			TaskQueue: "queue-sample",
		},
		updateResourcePoolSize, resourcePoolWorkflowID, 1)
	if err != nil {
		log.Println("failed to update resource pool size workflow", err)
	} else {
		log.Println("update resource pool size workflow started", updatePoolWorkflow.GetID())
	}

	// wait 3 seconds for the resource pool to update
	time.Sleep(3 * time.Second)

	// execute fifth query - after shrinking back to 1 resource
	queryAndPrintResourcePool(c, resourcePoolWorkflowID, "after shrinking back to 1 resource")

	// main thread waits long enough for all workflows to complete
	log.Println("waiting for all workflows to complete...")
	time.Sleep(30 * time.Second)

	// execute last query - after all workflows have completed
	queryAndPrintResourcePool(c, resourcePoolWorkflowID, "after all workflows have completed")

	log.Println("query test completed!")
}

// runPoolInspectorTest 运行资源池内部状态检查测试
func runPoolInspectorTest(c client.Client, poolID string) {
	ctx := context.Background()

	log.Printf("start checking the internal state of resource pool %s\n", poolID)

	// query resource pool basic status
	statusResp, err := c.QueryWorkflow(ctx, poolID, "", queue.GetResourcePoolStatusQuery)
	if err != nil {
		log.Fatalf("failed to query resource pool status: %v", err)
	}

	var status queue.ResourcePoolStatus
	if err := statusResp.Get(&status); err != nil {
		log.Fatalf("failed to parse resource pool status: %v", err)
	}

	// print resource pool basic information
	log.Println("==== resource pool basic information ====")
	log.Printf("resource ID: %s\n", status.ResourceID)
	log.Printf("total resources: %d\n", status.TotalResources)
	log.Printf("available resources: %d\n", status.AvailableCount)
	log.Printf("allocated resources: %d\n", status.AllocatedCount)
	log.Printf("waiting queue length: %d\n", status.WaitingCount)

	// query resource pool detailed allocation details
	allocationResp, err := c.QueryWorkflow(ctx, poolID, "", queue.GetResourceAllocationQuery)
	if err != nil {
		log.Fatalf("failed to query resource allocation details: %v", err)
	}

	var allocation queue.ResourceAllocation
	if err := allocationResp.Get(&allocation); err != nil {
		log.Fatalf("failed to parse resource allocation details: %v", err)
	}

	// print resource pool internal state JSON
	allocationJSON, err := json.MarshalIndent(allocation, "", "  ")
	if err != nil {
		log.Fatalf("failed to format JSON: %v", err)
	}
	log.Println("\n==== resource pool internal state (JSON) ====")
	log.Println(string(allocationJSON))

	// print resource details table
	log.Println("\n==== resource details ====")
	log.Printf("%-20s %-10s %-10s %-30s %-25s\n", "resource ID", "index", "status", "assigned to", "acquired time")
	log.Println(strings.Repeat("-", 100))

	for _, res := range allocation.Resources {
		status := "available"
		assignedTo := "-"
		acquiredTime := "-"

		if !res.Available {
			status = "allocated"
			assignedTo = res.AssignedTo
			acquiredTime = res.AcquiredTime.Format("2006-01-02 15:04:05")
		}

		log.Printf("%-20s %-10d %-10s %-30s %-25s\n",
			res.ResourceID,
			res.ResourceIndex,
			status,
			assignedTo,
			acquiredTime)
	}

	// print waiting queue
	if len(allocation.WaitingQueue) > 0 {
		log.Println("\n==== waiting queue ====")
		log.Printf("%-5s %-30s %-10s\n", "position", "workflow ID", "priority")
		log.Println(strings.Repeat("-", 50))

		for _, req := range allocation.WaitingQueue {
			log.Printf("%-5d %-30s %-10d\n",
				req.QueuePosition,
				req.WorkflowID,
				req.Priority)
		}
	} else {
		log.Println("\nwaiting queue is empty")
	}

	// start monitoring resource pool changes (update every 2 seconds, press Ctrl+C to stop)
	log.Println("\n==== start monitoring resource pool changes (update every 2 seconds, press Ctrl+C to stop) ====")
	log.Println("time                    total resources   available   allocated   waiting queue")
	log.Println(strings.Repeat("-", 70))

	startTime := time.Now()
	for time.Since(startTime) < 60*time.Second {
		// query resource pool status
		statusResp, err := c.QueryWorkflow(ctx, poolID, "", queue.GetResourcePoolStatusQuery)
		if err != nil {
			log.Printf("[%s] failed to query: %v\n", time.Now().Format("15:04:05"), err)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := statusResp.Get(&status); err != nil {
			log.Printf("[%s] failed to parse: %v\n", time.Now().Format("15:04:05"), err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("%-23s %-8d %-8d %-8d %-8d\n",
			time.Now().Format("2006-01-02 15:04:05"),
			status.TotalResources,
			status.AvailableCount,
			status.AllocatedCount,
			status.WaitingCount)

		time.Sleep(2 * time.Second)
	}
}

// queryAndPrintResourcePool 查询资源池并打印结果
func queryAndPrintResourcePool(c client.Client, resourcePoolWorkflowID string, stage string) {
	ctx := context.Background()

	log.Printf("\n===== query stage: %s =====\n", stage)

	// query resource pool status
	statusResp, err := c.QueryWorkflow(ctx, resourcePoolWorkflowID, "", queue.GetResourcePoolStatusQuery)
	if err != nil {
		log.Printf("failed to query resource pool status: %v\n", err)
		return
	}

	var status queue.ResourcePoolStatus
	if err := statusResp.Get(&status); err != nil {
		log.Printf("failed to parse resource pool status: %v\n", err)
		return
	}

	// print basic status information
	log.Println("resource pool status:")
	log.Printf("   resource ID: %s\n", status.ResourceID)
	log.Printf("   total resources: %d\n", status.TotalResources)
	log.Printf("   available resources: %d\n", status.AvailableCount)
	log.Printf("   allocated resources: %d\n", status.AllocatedCount)
	log.Printf("   waiting queue length: %d\n", status.WaitingCount)

	// query resource allocation details
	allocationResp, err := c.QueryWorkflow(ctx, resourcePoolWorkflowID, "", queue.GetResourceAllocationQuery)
	if err != nil {
		log.Printf("failed to query resource allocation details: %v\n", err)
		return
	}

	var allocation queue.ResourceAllocation
	if err := allocationResp.Get(&allocation); err != nil {
		log.Printf("failed to parse resource allocation details: %v\n", err)
		return
	}

	// print resource details
	log.Println("\nresource details:")
	for i, res := range allocation.Resources {
		log.Printf("   resource #%d:\n", i+1)
		log.Printf("    ID: %s\n", res.ResourceID)
		log.Printf("    index: %d\n", res.ResourceIndex)
		log.Printf("    available: %v\n", res.Available)

		// if the resource is allocated, display the allocation details
		if !res.Available {
			log.Printf("    assigned to: %s\n", res.AssignedTo)
			log.Printf("    acquired time: %s\n", res.AcquiredTime.Format("2006-01-02 15:04:05"))
		}

		// display metadata
		if len(res.Metadata) > 0 {
			metadataJSON, _ := json.MarshalIndent(res.Metadata, "    ", "  ")
			log.Printf("    metadata: %s\n", string(metadataJSON))
		}
	}

	// if there is a waiting queue, display the waiting queue information
	if len(allocation.WaitingQueue) > 0 {
		log.Println("\nwaiting queue:")
		for i, req := range allocation.WaitingQueue {
			log.Printf("   request #%d:\n", i+1)
			log.Printf("    workflow ID: %s\n", req.WorkflowID)
			log.Printf("    priority: %d\n", req.Priority)
			log.Printf("    queue position: %d\n", req.QueuePosition)
		}
	}

	log.Println("")
}

// cancelResourceRequest 取消资源请求的辅助工作流
func cancelResourceRequest(ctx workflow.Context, resourcePoolWorkflowID string, targetWorkflowID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("start cancel resource request workflow",
		"resourcePoolWorkflowID", resourcePoolWorkflowID,
		"targetWorkflowID", targetWorkflowID)

	// create resource pool workflow execution reference
	execution := workflow.Execution{
		ID: resourcePoolWorkflowID,
	}

	// send cancel command
	err := queue.CancelResourceRequest(ctx, execution, targetWorkflowID)
	if err != nil {
		logger.Error("failed to cancel resource request", "Error", err)
		return err
	}

	logger.Info("successfully canceled resource request")
	return nil
}

// updateResourcePoolSize 更新资源池大小的辅助工作流
func updateResourcePoolSize(ctx workflow.Context, resourcePoolWorkflowID string, newSize int) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("start update resource pool size workflow",
		"resourcePoolWorkflowID", resourcePoolWorkflowID,
		"newSize", newSize)

	// create resource pool workflow execution reference
	execution := workflow.Execution{
		ID: resourcePoolWorkflowID,
	}

	// send update command
	err := queue.UpdateResourcePool(ctx, execution, newSize)
	if err != nil {
		logger.Error("failed to update resource pool size", "Error", err)
		return err
	}

	logger.Info("successfully updated resource pool size")
	return nil
}
