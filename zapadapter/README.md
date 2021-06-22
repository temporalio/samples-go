### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run zaplogger/worker/main.go
```
3) Run the following command to start the example
```
go run zaplogger/starter/main.go
```
4) Check worker logs in colorful JSON format:
```json
2021-06-21T17:31:59.836-0700    INFO    internal/internal_worker.go:1001        Started Worker  {"Namespace": "default", "TaskQueue": "zap-logger", "WorkerID": "506926@MainWorker@"}
2021-06-21T17:32:02.246-0700    INFO    reflect/value.go:476    Logging from workflow   {"Namespace": "default", "TaskQueue": "zap-logger", "WorkerID": "506926@MainWorker@", "WorkflowType": "Workflow", "WorkflowID": "zap_logger_workflow_id", "RunID": "7f59f2a4-87a5-4a73-9020-076fdfa3c766", "Attempt": 1, "name": "<param to log>"}
2021-06-21T17:32:02.246-0700    DEBUG   internal/workflow.go:491        ExecuteActivity {"Namespace": "default", "TaskQueue": "zap-logger", "WorkerID": "506926@MainWorker@", "WorkflowType": "Workflow", "WorkflowID": "zap_logger_workflow_id", "RunID": "7f59f2a4-87a5-4a73-9020-076fdfa3c766", "Attempt": 1, "ActivityID": "5", "ActivityType": "Activity"}
2021-06-21T17:32:02.269-0700    INFO    reflect/value.go:337    Executing Activity.     {"Namespace": "default", "TaskQueue": "zap-logger", "WorkerID": "506926@MainWorker@", "ActivityID": "5", "ActivityType": "Activity", "Attempt": 1, "WorkflowType": "Workflow", "WorkflowID": "zap_logger_workflow_id", "RunID": "7f59f2a4-87a5-4a73-9020-076fdfa3c766", "name": "<param to log>"}
2021-06-21T17:32:02.269-0700    DEBUG   reflect/value.go:337    Debugging Activity.     {"Namespace": "default", "TaskQueue": "zap-logger", "WorkerID": "506926@MainWorker@", "ActivityID": "5", "ActivityType": "Activity", "Attempt": 1, "WorkflowType": "Workflow", "WorkflowID": "zap_logger_workflow_id", "RunID": "7f59f2a4-87a5-4a73-9020-076fdfa3c766", "value": "important debug data"}
2021-06-21T17:32:02.284-0700    DEBUG   internal/workflow.go:491        ExecuteActivity {"Namespace": "default", "TaskQueue": "zap-logger", "WorkerID": "506926@MainWorker@", "WorkflowType": "Workflow", "WorkflowID": "zap_logger_workflow_id", "RunID": "7f59f2a4-87a5-4a73-9020-076fdfa3c766", "Attempt": 1, "ActivityID": "11", "ActivityType": "ActivityError"}
2021-06-21T17:32:02.295-0700    WARN    reflect/value.go:337    Ignore next error message. It is just for demo purpose. {"Namespace": "default", "TaskQueue": "zap-logger", "WorkerID": "506926@MainWorker@", "ActivityID": "11", "ActivityType": "ActivityError", "Attempt": 1, "WorkflowType": "Workflow", "WorkflowID": "zap_logger_workflow_id", "RunID": "7f59f2a4-87a5-4a73-9020-076fdfa3c766"}
2021-06-21T17:32:02.295-0700    ERROR   reflect/value.go:337    Unable to execute ActivityError.        {"Namespace": "default", "TaskQueue": "zap-logger", "WorkerID": "506926@MainWorker@", "ActivityID": "11", "ActivityType": "ActivityError", "Attempt": 1, "WorkflowType": "Workflow", "WorkflowID": "zap_logger_workflow_id", "RunID": "7f59f2a4-87a5-4a73-9020-076fdfa3c766", "error": "random error"}
reflect.Value.Call
        /usr/local/go/src/reflect/value.go:337
go.temporal.io/sdk/internal.(*activityExecutor).Execute
        /home/user/go/pkg/mod/go.temporal.io/sdk@v1.7.0/internal/internal_worker.go:777
go.temporal.io/sdk/internal.(*activityTaskHandlerImpl).Execute
        /home/user/go/pkg/mod/go.temporal.io/sdk@v1.7.0/internal/internal_task_handlers.go:1816
go.temporal.io/sdk/internal.(*activityTaskPoller).ProcessTask
        /home/user/go/pkg/mod/go.temporal.io/sdk@v1.7.0/internal/internal_task_pollers.go:875
go.temporal.io/sdk/internal.(*baseWorker).processTask
        /home/user/go/pkg/mod/go.temporal.io/sdk@v1.7.0/internal/internal_worker_base.go:343
2021-06-21T17:32:02.310-0700    INFO    reflect/value.go:476    Workflow completed.     {"Namespace": "default", "TaskQueue": "zap-logger", "WorkerID": "506926@MainWorker@", "WorkflowType": "Workflow", "WorkflowID": "zap_logger_workflow_id", "RunID": "7f59f2a4-87a5-4a73-9020-076fdfa3c766", "Attempt": 1}
```