// @@@SNIPSTART go-samples-your-workflow-definition
package yourapp

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type YourWorkflowParam struct {
	WorkflowParamX string
	WorkflowParamY int
}

type YourWorkflowResultObject struct {
	ResultFieldX string
	ResultFieldY int
}

// Workflow is a Hello World workflow definition.
func YourWorkflowDefinition(ctx workflow.Context, param YourWorkflowParam) (YourWorkflowResultObject, error) {
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
	// Use a nil struct pointer to call Activities that are part of a struct
	var a *YourActivityObject
	err := workflow.ExecuteActivity(ctx, a.PrintSharedSate).Get(ctx, nil)
	if err != nil {
		return YourWorkflowResultObject{}, err
	}
	// Execute the Activity and wait for the result.
	var activityResult YourActivityResultObject
	err = workflow.ExecuteActivity(ctx, a.YourActivityDefinition, activityParam).Get(ctx, &activityResult)
	if err != nil {
		return YourWorkflowResultObject{}, err
	}
	err = workflow.ExecuteActivity(ctx, a.PrintSharedSate).Get(ctx, nil)
	if err != nil {
		return YourWorkflowResultObject{}, err
	}
	// Make the results of the Workflow Execution available.
	workflowResult := YourWorkflowResultObject {
		ResultFieldX: activityResult.ResultFieldX,
		ResultFieldY: activityResult.ResultFieldY,
	}
	return workflowResult, nil
}

func YourSimpleWorkflowDefinition(ctx workflow.Context) error {
	return nil
}
// @@@SNIPEND