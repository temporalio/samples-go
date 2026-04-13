package caller

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus-messaging/callerpattern/service"
)

const (
	CallerTaskQueue = "nexus-messaging-caller-task-queue"
	endpointName    = "my-nexus-endpoint-name"
)

// CallerWorkflow calls the NexusGreetingService operations for a given userID
// and returns a log of actions taken.
func CallerWorkflow(ctx workflow.Context, userID string) ([]string, error) {
	var log []string
	c := workflow.NewNexusClient(endpointName, service.ServiceName)

	// 1. Get supported languages.
	fut := c.ExecuteOperation(ctx, service.GetLanguagesOperationName, service.GetLanguagesInput{
		IncludeUnsupported: false,
		UserID:             userID,
	}, workflow.NexusOperationOptions{})
	var langsOut service.GetLanguagesOutput
	if err := fut.Get(ctx, &langsOut); err != nil {
		return nil, fmt.Errorf("getLanguages failed: %w", err)
	}
	log = append(log, fmt.Sprintf("getLanguages returned %d languages", len(langsOut.Languages)))

	// 2. Set language to Arabic.
	fut = c.ExecuteOperation(ctx, service.SetLanguageOperationName, service.SetLanguageInput{
		Language: service.Arabic,
		UserID:   userID,
	}, workflow.NexusOperationOptions{})
	var prevLang service.Language
	if err := fut.Get(ctx, &prevLang); err != nil {
		return nil, fmt.Errorf("setLanguage failed: %w", err)
	}
	log = append(log, fmt.Sprintf("setLanguage(Arabic) returned previous language: %s", prevLang))

	// 3. Get current language (assert it is Arabic).
	fut = c.ExecuteOperation(ctx, service.GetLanguageOperationName, service.GetLanguageInput{
		UserID: userID,
	}, workflow.NexusOperationOptions{})
	var currentLang service.Language
	if err := fut.Get(ctx, &currentLang); err != nil {
		return nil, fmt.Errorf("getLanguage failed: %w", err)
	}
	if currentLang != service.Arabic {
		return nil, fmt.Errorf("expected Arabic, got %s", currentLang)
	}
	log = append(log, fmt.Sprintf("getLanguage returned: %s (confirmed Arabic)", currentLang))

	// 4. Approve the workflow.
	fut = c.ExecuteOperation(ctx, service.ApproveOperationName, service.ApproveInput{
		Name:   "CallerWorkflow",
		UserID: userID,
	}, workflow.NexusOperationOptions{})
	var approveOut service.ApproveOutput
	if err := fut.Get(ctx, &approveOut); err != nil {
		return nil, fmt.Errorf("approve failed: %w", err)
	}
	log = append(log, "approve sent successfully")

	return log, nil
}
