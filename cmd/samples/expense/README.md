This sample workflow process an expense request. The key part of this sample is to show how to complete an activity asynchronously.

There are 3 steps in this expense processing workflow:
1) Create a new expense report.
2) Wait for the expense report to be approved. This could take an arbitrary amount of time. So the activity's Execute method has to return before it is actually approved. This is done by returning a special error so the framework knows the activity is not completed yet. 
    2.1) When the expense is approved (or rejected), somewhere in the world needs to be notified, and it will need to call WorkflowClient.CompleteActivity() to tell cadence service that that activity is now completed. In this sample case, the dummy server do this job. In real world, you will need to register some listener to the expense system or you will need to have your own pulling agent to check for the expense status periodic. 
3) After the wait activity is completed, it did the payment for the expense. (dummy step in this sample case)

This sample rely on an a dummy expense server to work.

Steps to run this sample:
1) You need a cadence service running. See cmd/samples/README.md for more details.
2) Run "./dummy" to start the dummy server. If dummy is not found, run make to build it.
3) Run "./expense -m worker" to start workers.
4) Run "./expense -m trigger" to kick off an expense workflow.
5) When you see the console print out the expense is created, go to [localhost](http://localhost:8080/list) to approve the expense.
6) You should see the workflow complete after you approve the expense. You can also reject the expense.
