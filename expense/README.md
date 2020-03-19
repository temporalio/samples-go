# Expense
This sample workflow process an expense request. The key part of this sample is to show how to complete an activity asynchronously.

# Sample Description
* Create a new expense report.
* Wait for the expense report to be approved. This could take an arbitrary amount of time. So the activity's Execute method has to return before it is actually approved. This is done by returning a special error so the framework knows the activity is not completed yet. 
  * When the expense is approved (or rejected), somewhere in the world needs to be notified, and it will need to call WorkflowClient.CompleteActivity() to tell cadence service that that activity is now completed. In this sample case, the dummy server do this job. In real world, you will need to register some listener to the expense system or you will need to have your own pulling agent to check for the expense status periodic. 
* After the wait activity is completed, it did the payment for the expense. (dummy step in this sample case)

This sample rely on an a dummy expense server to work.

# Steps To Run Sample
* You need a cadence service running. See https://github.com/uber/cadence/blob/master/README.md for more details.
* Start the dummy server 
```
./bin/dummy
```
If dummy is not found, run make to build it.
* Start workflow and activity workers
```
./bin/expense -m worker
```
* Start expanse workflow execution
```
./bin/expense -m trigger
```
* When you see the console print out the expense is created, go to [localhost:8099/list](http://localhost:8099/list) to approve the expense.
* You should see the workflow complete after you approve the expense. You can also reject the expense.
* If you see the workflow failed, try to change to a different port number in dummy.go and workflow.go. Then rebuild everything.
