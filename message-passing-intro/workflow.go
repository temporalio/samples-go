package update

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"golang.org/x/exp/maps"
)

type Language string

const Chinese Language = "chinese"
const English Language = "english"
const French Language = "french"
const Spanish Language = "spanish"
const Portuguese Language = "portuguese"

const GetLanguagesQuery = "get-languages"
const GetLanguageQuery = "get-language"
const SetLanguageUpdate = "set-language"
const ApproveSignal = "approve"

type ApproveInput struct {
	Name string
}

type GetLanguagesInput struct {
	IncludeUnsupported bool
}

func GreetingWorkflow(ctx workflow.Context) (string, error) {
	logger := workflow.GetLogger(ctx)
	approverName := ""
	language := English
	greeting := map[Language]string{English: "Hello", Chinese: "你好，世界"}
	err := workflow.SetQueryHandler(ctx, GetLanguagesQuery, func(input GetLanguagesInput) ([]Language, error) {
		// 👉 A Query handler returns a value: it can inspect but must not mutate the Workflow state.
		if input.IncludeUnsupported {
			return []Language{Chinese, English, French, Spanish, Portuguese}, nil
		} else {
			// Range over map is a nondeterministic operation.
			// It is OK to have a non-deterministic operation in a query function.
			//workflowcheck:ignore
			return maps.Keys(greeting), nil
		}
	})
	if err != nil {
		return "", err
	}

	err = workflow.SetQueryHandler(ctx, GetLanguageQuery, func(input GetLanguagesInput) (Language, error) {
		return language, nil
	})
	if err != nil {
		return "", err
	}

	err = workflow.SetUpdateHandlerWithOptions(ctx, SetLanguageUpdate, func(ctx workflow.Context, newLanguage Language) (Language, error) {
		// 👉 An Update handler can mutate the Workflow state and return a value.
		var previousLanguage Language
		previousLanguage, language = language, newLanguage
		return previousLanguage, nil
	}, workflow.UpdateHandlerOptions{
		Validator: func(ctx workflow.Context, newLanguage Language) error {
			if _, ok := greeting[newLanguage]; !ok {
				// 👉 In an Update validator you return any error to reject the Update.
				return fmt.Errorf("%s unsupported language", newLanguage)
			}
			return nil
		},
	})
	if err != nil {
		return "", err
	}
	// Block until the language is approved
	var approveInput ApproveInput
	workflow.GetSignalChannel(ctx, ApproveSignal).Receive(ctx, &approveInput)
	approverName = approveInput.Name
	logger.Info("Received approval", "Approver", approverName)

	return greeting[language], nil
}
