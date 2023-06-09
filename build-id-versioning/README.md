# Build ID Based Versioning
This sample illustrates how to use Build ID based versioning to help you appropriately roll out 
incompatible and compatible changes to workflow and activity code for the same task queue.

## Description
The sample shows you how to roll out both a compatible change and an incompatible change to a
workflow.

## Running
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run
    ```
    go run build-id-versioning/worker/main.go 
    ```
    to start the appropriate workers. It will print a task queue name to the console, which you
    will need to copy and paste when running the next step. This is to allow running the sample
    repeatedly without encountering issues due to Build IDs already existing on the queue.
   
3) Run
    ```
    go run build-id-versioning/starter/main.go <task queue name>
    ```
    to start the workflows.
