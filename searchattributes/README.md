### Steps to run this sample:
1) Before running this, Temporal Server need to run with an Advanced Visibility store (Elasticsearch integrated).
If you are using the default `docker-compose` config, then Elasticsearch is already integrated.
If not, then you can [integrate Elasticsearch](https://docs.temporal.io/docs/content/how-to-integrate-elasticsearch-into-a-temporal-cluster) yourself.
2) Run the following command to start the worker:
```
go run searchattributes/worker/main.go
```
3) Run the following command to start the example:
```
go run searchattributes/starter/main.go
```
4) Observe search attributes in the worker log:
```
...
2021/05/24 15:21:34 INFO  WorkflowID must be the same Namespace default TaskQueue search-attributes WorkerID 53250@worker@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_b0ffc4cb-ecac-42f0-82df-a5afb8eae719 RunID a0e5c968-cef9-4cb3-aa31-e842566e20d3 Attempt 1 info.WorkflowID search_attributes_b0ffc4cb-ecac-42f0-82df-a5afb8eae719 lastExecution.WorkflowId search_attributes_b0ffc4cb-ecac-42f0-82df-a5afb8eae719
2021/05/24 15:21:34 INFO  RunID must be the same Namespace default TaskQueue search-attributes WorkerID 53250@worker@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_b0ffc4cb-ecac-42f0-82df-a5afb8eae719 RunID a0e5c968-cef9-4cb3-aa31-e842566e20d3 Attempt 1 info.RunID a0e5c968-cef9-4cb3-aa31-e842566e20d3 lastExecution.RunId a0e5c968-cef9-4cb3-aa31-e842566e20d3
2021/05/24 15:21:34 INFO  Current search attribute values:
BinaryChecksums=[b8a1e5078c8cadc19af61384ff23f712]
CustomBoolField=true
CustomDatetimeField=2021-05-24T22:21:33.480946932Z
CustomDoubleField=3.14
CustomIntField=2
CustomKeywordField=Update2
CustomStringField=String field is for text. When query, it will be tokenized for partial match. StringTypeField cannot be used in Order By
 Namespace default TaskQueue search-attributes WorkerID 53250@worker@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_b0ffc4cb-ecac-42f0-82df-a5afb8eae719 RunID a0e5c968-cef9-4cb3-aa31-e842566e20d3 Attempt 1
2021/05/24 15:21:34 INFO  Workflow completed. Namespace default TaskQueue search-attributes WorkerID 53250@worker@ WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_b0ffc4cb-ecac-42f0-82df-a5afb8eae719 RunID a0e5c968-cef9-4cb3-aa31-e842566e20d3 Attempt 1
```
