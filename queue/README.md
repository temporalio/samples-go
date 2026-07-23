# Resource Competition and Queue Scheduling Example

This example demonstrates how to implement resource competition and queue scheduling mechanisms using Temporal, providing a flexible resource pool management system that supports dynamic adjustment of resource pool size, cancellation of resource requests, and monitoring functions.

## Overview

This example implements a resource pool and queue system based on Temporal, featuring the following characteristics:

1. **Parallel Execution**: Supports the parallel execution of multiple workflow instances, making full use of system resources.
2. **Shared Resource Pool**: Provides a limited number of shared resources for workflows to use.
3. **Preemptive Resource Acquisition**: Workflows acquire resources through a preemptive mechanism to ensure efficient utilization.
4. **Waiting Queue**: When resources are unavailable, workflows enter a waiting queue and continue execution after automatically acquiring resources.
5. **Dynamic Adjustment of Resource Pool**: Supports dynamic adjustment of the resource pool size (expansion or reduction) at runtime.
6. **Cancellation of Resource Requests**: Supports cancellation of resource requests in the waiting queue, terminating workflows that are no longer needed.
7. **Real-time Monitoring**: Provides real-time monitoring capabilities to track the status and changes of the resource pool.

## Component Description

- **ResourcePool**: The structure of the resource pool, providing resource acquisition, release, and management functions.
- **ResourcePoolWorkflow**: The workflow that manages resource allocation and queues.
- **SampleWorkflowWithResourcePool**: An example workflow that uses the resource pool, demonstrating resource acquisition and release.
- **UpdateResourcePool**: The functionality to dynamically adjust the size of the resource pool.
- **CancelResourceRequest**: Cancels resource requests in the waiting queue.
- **ResourcePoolInitializer**: An interface for customizing resource pool initialization and scaling behavior.

## Key Features

- **Resource Allocation**: The resource pool workflow manages the allocation of limited resources to ensure efficient utilization.
- **Queuing Mechanism**: Workflows that have not acquired resources enter a waiting queue and continue execution after automatically acquiring resources.
- **Dynamic Scaling**: Supports adjusting the size of the resource pool at runtime to meet varying load demands.
- **Request Cancellation**: Supports terminating resource requests in the waiting queue to avoid unnecessary resource occupation.
- **Signal Communication**: Uses Temporal's signal mechanism for communication between workflows.
- **Persistence**: Even if the system crashes, the state of waiting workflows and resources can be restored.
- **Delayed Scaling Down**: Intelligently waits for resources to be released before completing scaling down, without affecting resources currently in use.

# Usage Example

### Start Workflow

You can start the workflow with the following command:

```bash
go run queue/starter/main.go -test=basic
```

### Query Resource Pool Status

To query the status of the resource pool, you can use the following command：

```bash
go run queue/starter/main.go -test=pool -poolid=resource-pool:{namespace}:{resourceID}
```

### Running Tests

The project includes unit tests, and you can run the tests using the following command:

```bash
go test ./...
```

## Dependencies

- [Temporal Go SDK](https://github.com/temporalio/sdk-go)
- [Testify](https://github.com/stretchr/testify)

## Contribution

Contributions of any kind are welcome! Please submit issues or pull requests.

## License

This project is licensed under the MIT License. For more details, please see the LICENSE file.
