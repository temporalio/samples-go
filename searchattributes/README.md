### Steps to run this sample:
1) Before running this, Temporal Server need to run with advanced visibility store.
See [https://docs.temporal.io/docs/server/elasticsearch-setup](https://docs.temporal.io/docs/server/elasticsearch-setup)
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
2021/03/02 14:56:14 INFO  Current search attributes:
BinaryChecksums=[cf3ef64750699d978e45ad654d47377e]
CustomBoolField=true
CustomDatetimeField=2019-08-22T00:00:00-07:00
CustomDoubleField=3.14
CustomIntField=2
CustomKeywordField=Update2
CustomStringField=String field is for text. When query, it will be tokenized for partial match. StringTypeField cannot be used in Order By
 Namespace default TaskQueue search-attributes WorkerID 1317283@server WorkflowType SearchAttributesWorkflow WorkflowID search_attributes_552c5869-f716-4918-a67e-a7e12ff3f774 RunID a746c151-46ae-4d88-96fc-30fd6457348e Attempt 1
...
```
