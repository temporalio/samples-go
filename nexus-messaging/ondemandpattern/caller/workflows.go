package caller

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus-messaging/ondemandpattern/service"
)

const (
	CallerTaskQueue = "nexus-messaging-caller-task-queue"
	endpointName    = "my-nexus-endpoint-name"
)

// CallerRemoteWorkflow starts two remote GreetingWorkflows via runFromRemote,
// interacts with both via queries/updates/signals, then waits for results.
func CallerRemoteWorkflow(ctx workflow.Context) ([]string, error) {
	var log []string
	c := workflow.NewNexusClient(endpointName, service.ServiceName)

	userID1 := "nexus-messaging-greeting-one"
	userID2 := "nexus-messaging-greeting-two"

	// Start both remote workflows asynchronously.
	fut1 := c.ExecuteOperation(ctx, service.RunFromRemoteOperationName, service.RunFromRemoteInput{
		UserID: userID1,
	}, workflow.NexusOperationOptions{})

	fut2 := c.ExecuteOperation(ctx, service.RunFromRemoteOperationName, service.RunFromRemoteInput{
		UserID: userID2,
	}, workflow.NexusOperationOptions{})

	// Wait for both workflows to start before interacting.
	var exec1, exec2 workflow.NexusOperationExecution
	if err := fut1.GetNexusOperationExecution().Get(ctx, &exec1); err != nil {
		return nil, fmt.Errorf("runFromRemote (one) start failed: %w", err)
	}
	log = append(log, fmt.Sprintf("started remote workflow one: %s", userID1))

	if err := fut2.GetNexusOperationExecution().Get(ctx, &exec2); err != nil {
		return nil, fmt.Errorf("runFromRemote (two) start failed: %w", err)
	}
	log = append(log, fmt.Sprintf("started remote workflow two: %s", userID2))

	// Query languages from workflow one.
	langsFut1 := c.ExecuteOperation(ctx, service.GetLanguagesOperationName, service.GetLanguagesInput{
		IncludeUnsupported: false,
		UserID:             userID1,
	}, workflow.NexusOperationOptions{})
	var langsOut1 service.GetLanguagesOutput
	if err := langsFut1.Get(ctx, &langsOut1); err != nil {
		return nil, fmt.Errorf("getLanguages (one) failed: %w", err)
	}
	log = append(log, fmt.Sprintf("getLanguages (one) returned %d languages", len(langsOut1.Languages)))

	// Query languages from workflow two.
	langsFut2 := c.ExecuteOperation(ctx, service.GetLanguagesOperationName, service.GetLanguagesInput{
		IncludeUnsupported: true,
		UserID:             userID2,
	}, workflow.NexusOperationOptions{})
	var langsOut2 service.GetLanguagesOutput
	if err := langsFut2.Get(ctx, &langsOut2); err != nil {
		return nil, fmt.Errorf("getLanguages (two) failed: %w", err)
	}
	log = append(log, fmt.Sprintf("getLanguages (two) with unsupported returned %d languages", len(langsOut2.Languages)))

	// Set language to French for workflow one.
	setFut1 := c.ExecuteOperation(ctx, service.SetLanguageOperationName, service.SetLanguageInput{
		Language: service.French,
		UserID:   userID1,
	}, workflow.NexusOperationOptions{})
	var prevLang1 service.Language
	if err := setFut1.Get(ctx, &prevLang1); err != nil {
		return nil, fmt.Errorf("setLanguage (one) failed: %w", err)
	}
	log = append(log, fmt.Sprintf("setLanguage(French) on one returned previous: %s", prevLang1))

	// Set language to Spanish for workflow two.
	setFut2 := c.ExecuteOperation(ctx, service.SetLanguageOperationName, service.SetLanguageInput{
		Language: service.Spanish,
		UserID:   userID2,
	}, workflow.NexusOperationOptions{})
	var prevLang2 service.Language
	if err := setFut2.Get(ctx, &prevLang2); err != nil {
		return nil, fmt.Errorf("setLanguage (two) failed: %w", err)
	}
	log = append(log, fmt.Sprintf("setLanguage(Spanish) on two returned previous: %s", prevLang2))

	// Confirm current language for workflow one.
	getLangFut1 := c.ExecuteOperation(ctx, service.GetLanguageOperationName, service.GetLanguageInput{
		UserID: userID1,
	}, workflow.NexusOperationOptions{})
	var currentLang1 service.Language
	if err := getLangFut1.Get(ctx, &currentLang1); err != nil {
		return nil, fmt.Errorf("getLanguage (one) failed: %w", err)
	}
	log = append(log, fmt.Sprintf("getLanguage (one) = %s", currentLang1))

	// Confirm current language for workflow two.
	getLangFut2 := c.ExecuteOperation(ctx, service.GetLanguageOperationName, service.GetLanguageInput{
		UserID: userID2,
	}, workflow.NexusOperationOptions{})
	var currentLang2 service.Language
	if err := getLangFut2.Get(ctx, &currentLang2); err != nil {
		return nil, fmt.Errorf("getLanguage (two) failed: %w", err)
	}
	log = append(log, fmt.Sprintf("getLanguage (two) = %s", currentLang2))

	// Approve workflow one.
	approveFut1 := c.ExecuteOperation(ctx, service.ApproveOperationName, service.ApproveInput{
		Name:   "CallerRemoteWorkflow",
		UserID: userID1,
	}, workflow.NexusOperationOptions{})
	var approveOut1 service.ApproveOutput
	if err := approveFut1.Get(ctx, &approveOut1); err != nil {
		return nil, fmt.Errorf("approve (one) failed: %w", err)
	}
	log = append(log, "approved workflow one")

	// Approve workflow two.
	approveFut2 := c.ExecuteOperation(ctx, service.ApproveOperationName, service.ApproveInput{
		Name:   "CallerRemoteWorkflow",
		UserID: userID2,
	}, workflow.NexusOperationOptions{})
	var approveOut2 service.ApproveOutput
	if err := approveFut2.Get(ctx, &approveOut2); err != nil {
		return nil, fmt.Errorf("approve (two) failed: %w", err)
	}
	log = append(log, "approved workflow two")

	// Wait for both remote workflows to complete.
	var result1 string
	if err := fut1.Get(ctx, &result1); err != nil {
		return nil, fmt.Errorf("remote workflow one failed: %w", err)
	}
	log = append(log, fmt.Sprintf("remote workflow one result: %s", result1))

	var result2 string
	if err := fut2.Get(ctx, &result2); err != nil {
		return nil, fmt.Errorf("remote workflow two failed: %w", err)
	}
	log = append(log, fmt.Sprintf("remote workflow two result: %s", result2))

	return log, nil
}
