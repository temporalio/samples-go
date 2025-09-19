### Steps to run this sample:

1. Run a Temporal dev server[Temporal Server](Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).

  ```
  temporal server start-dev \
    --search-attribute CustomIntField=Int \
    --search-attribute CustomKeywordField=Keyword \
    --search-attribute CustomBoolField=Bool \
    --search-attribute CustomDoubleField=Double \
    --search-attribute CustomDatetimeField=Datetime  \
    --search-attribute CustomKeywordListField=KeywordList
  ```

1. Run the following command to start the worker:

  ```
  go run searchattributes/worker/main.go
  ```

1. Run the following command to start the example:

  ```
  go run searchattributes/starter/main.go
  ```

1. Observe search attributes in the worker log:

  ```
  ...
  2025/09/05 11:44:22 INFO  WorkflowID must be the same Namespace default TaskQueue search-attributes WorkerID 63326@local@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf RunID 01991b31-9fd1-7854-813d-0b49a53b5b1d Attempt 1 info.WorkflowID search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf lastExecution.WorkflowId search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf
  2025/09/05 11:44:22 INFO  RunID must be the same Namespace default TaskQueue search-attributes WorkerID 63326@local@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf RunID 01991b31-9fd1-7854-813d-0b49a53b5b1d Attempt 1 info.RunID 01991b31-9fd1-7854-813d-0b49a53b5b1d lastExecution.RunId 01991b31-9fd1-7854-813d-0b49a53b5b1d
  2025/09/05 11:44:22 INFO  Current search attribute value Namespace default TaskQueue search-attributes WorkerID 63326@local@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf RunID 01991b31-9fd1-7854-813d-0b49a53b5b1d Attempt 1 CustomDatetimeField 2025-09-05 18:44:21.075067 +0000 UTC
  2025/09/05 11:44:22 INFO  Current search attribute value Namespace default TaskQueue search-attributes WorkerID 63326@local@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf RunID 01991b31-9fd1-7854-813d-0b49a53b5b1d Attempt 1 CustomKeywordListField [value1 value2]
  2025/09/05 11:44:22 INFO  Current search attribute value Namespace default TaskQueue search-attributes WorkerID 63326@local@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf RunID 01991b31-9fd1-7854-813d-0b49a53b5b1d Attempt 1 CustomKeywordField Keyword fields supports prefix search
  2025/09/05 11:44:22 INFO  Current search attribute value Namespace default TaskQueue search-attributes WorkerID 63326@local@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf RunID 01991b31-9fd1-7854-813d-0b49a53b5b1d Attempt 1 CustomBoolField true
  2025/09/05 11:44:22 INFO  Current search attribute value Namespace default TaskQueue search-attributes WorkerID 63326@local@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf RunID 01991b31-9fd1-7854-813d-0b49a53b5b1d Attempt 1 CustomIntField 2
  2025/09/05 11:44:22 INFO  Workflow completed. Namespace default TaskQueue search-attributes WorkerID 63326@local@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_02243e18-58c0-4f7d-a893-49f4d30e8ccf RunID 01991b31-9fd1-7854-813d-0b49a53b5b1d Attempt 1
  ```
