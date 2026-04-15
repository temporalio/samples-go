package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus-messaging/callerpattern/service"
)

const (
	HandlerTaskQueue = "my-handler-task-queue"
	WorkflowIDPrefix = "GreetingWorkflow_for_"

	queryGetLanguages = "getLanguages"
	queryGetLanguage  = "getLanguage"
	updateSetLanguage = "setLanguage"
	signalApprove     = "approve"
)

// GetWorkflowID returns the workflow ID for a given user ID.
func GetWorkflowID(userID string) string {
	return WorkflowIDPrefix + userID
}

// GetLanguagesOperation queries a workflow for the supported languages.
var GetLanguagesOperation = nexus.NewSyncOperation(service.GetLanguagesOperationName, func(ctx context.Context, input service.GetLanguagesInput, options nexus.StartOperationOptions) (service.GetLanguagesOutput, error) {
	c := temporalnexus.GetClient(ctx)
	workflowID := GetWorkflowID(input.UserID)

	encodedVal, err := c.QueryWorkflow(ctx, workflowID, "", queryGetLanguages, input.IncludeUnsupported)
	if err != nil {
		return service.GetLanguagesOutput{}, fmt.Errorf("failed to query workflow: %w", err)
	}
	var output service.GetLanguagesOutput
	if err := encodedVal.Get(&output); err != nil {
		return service.GetLanguagesOutput{}, fmt.Errorf("failed to decode query result: %w", err)
	}
	return output, nil
})

// GetLanguageOperation queries a workflow for the current language.
var GetLanguageOperation = nexus.NewSyncOperation(service.GetLanguageOperationName, func(ctx context.Context, input service.GetLanguageInput, options nexus.StartOperationOptions) (service.Language, error) {
	c := temporalnexus.GetClient(ctx)
	workflowID := GetWorkflowID(input.UserID)

	encodedVal, err := c.QueryWorkflow(ctx, workflowID, "", queryGetLanguage)
	if err != nil {
		return 0, fmt.Errorf("failed to query workflow: %w", err)
	}
	var lang service.Language
	if err := encodedVal.Get(&lang); err != nil {
		return 0, fmt.Errorf("failed to decode query result: %w", err)
	}
	return lang, nil
})

// SetLanguageOperation updates a workflow's language.
var SetLanguageOperation = nexus.NewSyncOperation(service.SetLanguageOperationName, func(ctx context.Context, input service.SetLanguageInput, options nexus.StartOperationOptions) (service.Language, error) {
	c := temporalnexus.GetClient(ctx)
	workflowID := GetWorkflowID(input.UserID)

	handle, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
		WorkflowID:   workflowID,
		UpdateName:   updateSetLanguage,
		Args:         []interface{}{input.Language},
		WaitForStage: client.WorkflowUpdateStageCompleted,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to update workflow: %w", err)
	}
	var prevLang service.Language
	if err := handle.Get(ctx, &prevLang); err != nil {
		return 0, fmt.Errorf("failed to get update result: %w", err)
	}
	return prevLang, nil
})

// ApproveOperation signals a workflow to approve.
var ApproveOperation = nexus.NewSyncOperation(service.ApproveOperationName, func(ctx context.Context, input service.ApproveInput, options nexus.StartOperationOptions) (service.ApproveOutput, error) {
	c := temporalnexus.GetClient(ctx)
	workflowID := GetWorkflowID(input.UserID)

	if err := c.SignalWorkflow(ctx, workflowID, "", signalApprove, input.Name); err != nil {
		return service.ApproveOutput{}, fmt.Errorf("failed to signal workflow: %w", err)
	}
	return service.ApproveOutput{}, nil
})

// GreetingWorkflow is a long-running workflow that supports queries, updates, and signals.
func GreetingWorkflow(ctx workflow.Context, userID string) (string, error) {
	logger := workflow.GetLogger(ctx)

	language := service.English
	approved := false
	approvedBy := ""
	lock := workflow.NewMutex(ctx)

	initialGreetings := map[service.Language]string{
		service.Chinese: "你好，世界",
		service.English: "Hello, world",
	}

	// Register query: getLanguages
	if err := workflow.SetQueryHandler(ctx, queryGetLanguages, func(includeUnsupported bool) (service.GetLanguagesOutput, error) {
		if includeUnsupported {
			return service.GetLanguagesOutput{
				Languages: []service.Language{
					service.Arabic, service.Chinese, service.English,
					service.French, service.Hindi, service.Portuguese, service.Spanish,
				},
			}, nil
		}
		supported := make([]service.Language, 0, len(initialGreetings))
		for lang := range initialGreetings {
			supported = append(supported, lang)
		}
		return service.GetLanguagesOutput{Languages: supported}, nil
	}); err != nil {
		return "", err
	}

	// Register query: getLanguage
	if err := workflow.SetQueryHandler(ctx, queryGetLanguage, func() (service.Language, error) {
		return language, nil
	}); err != nil {
		return "", err
	}

	// Register update: setLanguage (with validator)
	if err := workflow.SetUpdateHandlerWithOptions(ctx, updateSetLanguage,
		func(ctx workflow.Context, newLang service.Language) (service.Language, error) {
			if err := lock.Lock(ctx); err != nil {
				return 0, err
			}
			defer lock.Unlock()

			prevLang := language

			// If the language is not in the initial greetings map, call the activity to fetch it.
			if _, ok := initialGreetings[newLang]; !ok {
				actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
					StartToCloseTimeout: 10 * time.Second,
					RetryPolicy: &temporal.RetryPolicy{
						MaximumAttempts: 3,
					},
				})
				var greetingsMap map[service.Language]string
				if err := workflow.ExecuteActivity(actCtx, GreetingActivity).Get(actCtx, &greetingsMap); err != nil {
					return 0, fmt.Errorf("activity failed: %w", err)
				}
				for lang, greeting := range greetingsMap {
					initialGreetings[lang] = greeting
				}
			}

			language = newLang
			logger.Info("Language updated", "from", prevLang, "to", newLang)
			return prevLang, nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx workflow.Context, newLang service.Language) error {
				if newLang < service.Arabic || newLang > service.Spanish {
					return fmt.Errorf("unsupported language: %d", newLang)
				}
				return nil
			},
		},
	); err != nil {
		return "", err
	}

	// Handle approve signal.
	approveCh := workflow.GetSignalChannel(ctx, signalApprove)
	workflow.Go(ctx, func(ctx workflow.Context) {
		var name string
		approveCh.Receive(ctx, &name)
		approved = true
		approvedBy = name
		logger.Info("Workflow approved", "by", name)
	})

	// Wait for approve signal and all handlers to finish.
	if err := workflow.Await(ctx, func() bool {
		return approved && workflow.AllHandlersFinished(ctx)
	}); err != nil {
		return "", err
	}

	greeting, ok := initialGreetings[language]
	if !ok {
		return "", fmt.Errorf("no greeting for language %s", language)
	}
	return fmt.Sprintf("%s (approved by %s)", greeting, approvedBy), nil
}

// GreetingActivity returns a map of all supported language greetings.
func GreetingActivity(_ context.Context) (map[service.Language]string, error) {
	return map[service.Language]string{
		service.Arabic:     "مرحبا بالعالم",
		service.Chinese:    "你好，世界",
		service.English:    "Hello, world",
		service.French:     "Bonjour, monde",
		service.Hindi:      "नमस्ते दुनिया",
		service.Portuguese: "Olá, mundo",
		service.Spanish:    "Hola, mundo",
	}, nil
}
