// @@@SNIPSTART go-samples-your-workflow-definition
package yourapp

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// YourWorkflowParam is the object passed to the Workflow.
type YourWorkflowParam struct {
	WorkflowParamX string
	WorkflowParamY int
}

// YourWorkflowResultObject is the object returned by the Workflow.
type YourWorkflowResultObject struct {
	WFResultFieldX string
	WFResultFieldY int
}

// YourWorkflowDefinition is your custom Workflow Definition.
func YourWorkflowDefinition(ctx workflow.Context, param YourWorkflowParam) (*YourWorkflowResultObject, error) {
	// Set the options for the Activity Execution.
	// Either StartToClose Timeout OR ScheduleToClose is required.
	// Not specifying a Task Queue will default to the parent Workflow Task Queue.
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	activityParam := YourActivityParam{
		ActivityParamX: param.WorkflowParamX,
		ActivityParamY: param.WorkflowParamY,
	}
	// Use a nil struct pointer to call Activities that are part of a struct.
	var a *YourActivityObject
	err := workflow.ExecuteActivity(ctx, a.PrintSharedSate).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	// Execute the Activity and wait for the result.
	var activityResult YourActivityResultObject
	err = workflow.ExecuteActivity(ctx, a.YourActivityDefinition, activityParam).Get(ctx, &activityResult)
	if err != nil {
		return nil, err
	}
	// Execute another Activity and wait for the result.
	err = workflow.ExecuteActivity(ctx, a.PrintSharedSate).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	// Make the results of the Workflow Execution available.
	workflowResult := &YourWorkflowResultObject{
		WFResultFieldX: activityResult.ResultFieldX,
		WFResultFieldY: activityResult.ResultFieldY,
	}
	return workflowResult, nil
}

// YourSimpleWorkflowDefintiion is the most basic Workflow Defintion.
func YourSimpleWorkflowDefinition(ctx workflow.Context) error {
	return nil
}

// @@@SNIPEND
