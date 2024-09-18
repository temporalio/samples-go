# Atomic message handlers

This sample shows off important techniques for handling signals and updates, aka messages.  In particular, it illustrates how message handlers can interleave or not be completed before the workflow completes, and how you can manage that.

* Here, using workflow.Await, signal and update handlers will only operate when the workflow is within a certain state--between cluster_started and cluster_shutdown.
* You can run start_workflow with an initializer signal that you want to run before anything else other than the workflow's constructor.  This pattern is known as "signal-with-start."
* Message handlers can block and their actions can be interleaved with one another and with the main workflow.  This can easily cause bugs, so you can use a lock to protect shared state from interleaved access.
* An "Entity" workflow, i.e. a long-lived workflow, periodically "continues as new".  It must do this to prevent its history from growing too large, and it passes its state to the next workflow.  You can check `workflow.GetInfo().GetContinueAsNewSuggested()` to see when it's time. 
* Most people want their message handlers to finish before the workflow run completes or continues as new.  Use `workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) }` to achieve this.
* Message handlers can be made idempotent.  See update `ClusterManager.assign_nodes_to_job`.

To run, first see [README.md](../../README.md) for prerequisites.

Then, run the following from this directory to run the worker:

    go run safe_message_handler/worker/main.go

Then, in another terminal, run the following to execute the workflow:

    go run safe_message_handler/starter/main.go

This will start a worker to run your workflow and activities, then start a ClusterManagerWorkflow and put it through its paces.
