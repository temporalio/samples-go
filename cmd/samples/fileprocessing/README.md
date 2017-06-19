This sample workflow demos a file processing process. The key part is to show how to use host specific tasklist. 

The workflow first starts an activity to download a requested resource file from web and store it locally on the host where it runs the download activity. Then, the workflow will start more activities to process the downloaded resource file. The key part is the following activities have to be run on the same host as the initial downloading activity. This is achieved by using host specific task list.

Steps to run this sample: 
1) You need a cadence service running. See details in cmd/samples/README.md
2) Run "./fileprocessing -m worker" multiple times on different console window. This is to simulate running workers on multiple different machines.
3) Run "./fileprocessing -m trigger" to submit a start request for this fileprocessing workflow.

You should see that all activities for one particular workflow execution are scheduled to run on one console window.
