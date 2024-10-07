package main

import (
	"context"
	"log"

	message "github.com/temporalio/samples-go/message-passing-intro"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "message-passing-intro-workflow-ID",
		TaskQueue: "message-passing-intro",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, message.GreetingWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	supportedLangResult, err := c.QueryWorkflow(context.Background(), we.GetID(), we.GetRunID(), message.GetLanguagesQuery, message.GetLanguagesInput{IncludeUnsupported: false})
	if err != nil {
		log.Fatalf("Unable to query workflow: %v", err)
	}
	var supportedLang []message.Language
	err = supportedLangResult.Get(&supportedLang)
	if err != nil {
		log.Fatalf("Unable to get query result: %v", err)
	}
	log.Println("Supported languages:", supportedLang)

	langResult, err := c.QueryWorkflow(context.Background(), we.GetID(), we.GetRunID(), message.GetLanguageQuery, message.GetLanguagesInput{})
	if err != nil {
		log.Fatalf("Unable to query workflow: %v", err)
	}
	var currentLang message.Language
	err = langResult.Get(&currentLang)
	if err != nil {
		log.Fatalf("Unable to get query result: %v", err)
	}
	log.Println("Current language:", currentLang)

	updateHandle, err := c.UpdateWorkflow(context.Background(), client.UpdateWorkflowOptions{
		WorkflowID:   we.GetID(),
		RunID:        we.GetRunID(),
		UpdateName:   message.SetLanguageUpdate,
		WaitForStage: client.WorkflowUpdateStageAccepted,
		Args:         []interface{}{message.Chinese},
	})
	if err != nil {
		log.Fatalf("Unable to update workflow: %v", err)
	}
	var previousLang message.Language
	err = updateHandle.Get(context.Background(), &previousLang)
	if err != nil {
		log.Fatalf("Unable to get update result: %v", err)
	}

	langResult, err = c.QueryWorkflow(context.Background(), we.GetID(), we.GetRunID(), message.GetLanguageQuery, message.GetLanguagesInput{})
	if err != nil {
		log.Fatalf("Unable to query workflow: %v", err)
	}
	err = langResult.Get(&currentLang)
	if err != nil {
		log.Fatalf("Unable to get query result: %v", err)
	}
	log.Printf("Language changed: %s -> %s", previousLang, currentLang)

	updateHandle, err = c.UpdateWorkflow(context.Background(), client.UpdateWorkflowOptions{
		WorkflowID:   we.GetID(),
		RunID:        we.GetRunID(),
		UpdateName:   message.SetLanguageUpdate,
		WaitForStage: client.WorkflowUpdateStageAccepted,
		Args:         []interface{}{message.English},
	})
	if err != nil {
		log.Fatalf("Unable to update workflow: %v", err)
	}
	err = updateHandle.Get(context.Background(), &previousLang)
	if err != nil {
		log.Fatalf("Unable to get update result: %v", err)
	}

	langResult, err = c.QueryWorkflow(context.Background(), we.GetID(), we.GetRunID(), message.GetLanguageQuery, message.GetLanguagesInput{})
	if err != nil {
		log.Fatalf("Unable to query workflow: %v", err)
	}
	err = langResult.Get(&currentLang)
	if err != nil {
		log.Fatalf("Unable to get query result: %v", err)
	}
	log.Printf("Language changed: %s -> %s", previousLang, currentLang)

	err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), message.ApproveSignal, message.ApproveInput{Name: ""})
	if err != nil {
		log.Fatalf("Unable to signal workflow: %v", err)
	}

	var wfresult string
	if err = we.Get(context.Background(), &wfresult); err != nil {
		log.Fatalf("unable get workflow result: %v", err)
	}
	log.Println("workflow result:", wfresult)
}
