package handler

import (
	"context"
	"fmt"
	"time"

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

var allLanguages = []service.Language{
	service.Arabic, service.Chinese, service.English,
	service.French, service.Hindi, service.Portuguese, service.Spanish,
}

// GetWorkflowID returns the workflow ID for a given user ID.
func GetWorkflowID(userID string) string {
	return WorkflowIDPrefix + userID
}

// GetLanguagesOperation queries a workflow for the supported languages.
var GetLanguagesOperation = temporalnexus.MustNewTemporalOperation(
	temporalnexus.TemporalOperationOptions[service.GetLanguagesInput, service.GetLanguagesOutput]{
		Name: service.GetLanguagesOperationName,
		Start: func(ctx context.Context, nc temporalnexus.NexusClient, input service.GetLanguagesInput, options temporalnexus.StartTemporalOperationOptions) (temporalnexus.TemporalOperationResult[service.GetLanguagesOutput], error) {
			var zero temporalnexus.TemporalOperationResult[service.GetLanguagesOutput]

			encodedVal, err := nc.GetWorkflowClient().QueryWorkflow(ctx, GetWorkflowID(input.UserID), "", queryGetLanguages, input.IncludeUnsupported)
			if err != nil {
				return zero, fmt.Errorf("failed to query workflow: %w", err)
			}
			var output service.GetLanguagesOutput
			if err := encodedVal.Get(&output); err != nil {
				return zero, fmt.Errorf("failed to decode query result: %w", err)
			}
			return temporalnexus.NewSyncResult(output), nil
		},
	},
)

// GetLanguageOperation queries a workflow for the current language.
var GetLanguageOperation = temporalnexus.MustNewTemporalOperation(
	temporalnexus.TemporalOperationOptions[service.GetLanguageInput, service.Language]{
		Name: service.GetLanguageOperationName,
		Start: func(ctx context.Context, nc temporalnexus.NexusClient, input service.GetLanguageInput, options temporalnexus.StartTemporalOperationOptions) (temporalnexus.TemporalOperationResult[service.Language], error) {
			var zero temporalnexus.TemporalOperationResult[service.Language]

			encodedVal, err := nc.GetWorkflowClient().QueryWorkflow(ctx, GetWorkflowID(input.UserID), "", queryGetLanguage)
			if err != nil {
				return zero, fmt.Errorf("failed to query workflow: %w", err)
			}
			var lang service.Language
			if err := encodedVal.Get(&lang); err != nil {
				return zero, fmt.Errorf("failed to decode query result: %w", err)
			}
			return temporalnexus.NewSyncResult(lang), nil
		},
	},
)

// SetLanguageOperation updates a workflow's language.
//
// Starting the update through the Nexus client makes this an asynchronous Nexus operation: the
// caller receives an operation token and the update result is delivered later over the Nexus
// completion callback. If the update happens to be complete by the time the start call returns,
// the result comes back synchronously instead.
var SetLanguageOperation = temporalnexus.MustNewTemporalOperation(
	temporalnexus.TemporalOperationOptions[service.SetLanguageInput, service.Language]{
		Name: service.SetLanguageOperationName,
		Start: func(ctx context.Context, nc temporalnexus.NexusClient, input service.SetLanguageInput, options temporalnexus.StartTemporalOperationOptions) (temporalnexus.TemporalOperationResult[service.Language], error) {
			return temporalnexus.StartUpdateWorkflow[service.Language](ctx, nc, client.UpdateWorkflowOptions{
				WorkflowID:   GetWorkflowID(input.UserID),
				UpdateName:   updateSetLanguage,
				Args:         []interface{}{input.Language},
				WaitForStage: client.WorkflowUpdateStageAccepted,
			})
		},
	},
)

// ApproveOperation signals a workflow to approve.
var ApproveOperation = temporalnexus.MustNewTemporalOperation(
	temporalnexus.TemporalOperationOptions[service.ApproveInput, service.ApproveOutput]{
		Name: service.ApproveOperationName,
		Start: func(ctx context.Context, nc temporalnexus.NexusClient, input service.ApproveInput, options temporalnexus.StartTemporalOperationOptions) (temporalnexus.TemporalOperationResult[service.ApproveOutput], error) {
			var zero temporalnexus.TemporalOperationResult[service.ApproveOutput]

			if err := nc.GetWorkflowClient().SignalWorkflow(ctx, GetWorkflowID(input.UserID), "", signalApprove, input.Name); err != nil {
				return zero, fmt.Errorf("failed to signal workflow: %w", err)
			}
			return temporalnexus.NewSyncResult(service.ApproveOutput{}), nil
		},
	},
)

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
			return service.GetLanguagesOutput{Languages: append([]service.Language(nil), allLanguages...)}, nil
		}
		supported := make([]service.Language, 0, len(initialGreetings))
		for _, lang := range allLanguages {
			if _, ok := initialGreetings[lang]; ok {
				supported = append(supported, lang)
			}
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
				for _, lang := range allLanguages {
					if greeting, ok := greetingsMap[lang]; ok {
						initialGreetings[lang] = greeting
					}
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
