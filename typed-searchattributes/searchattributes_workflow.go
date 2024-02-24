package typedsearchattributes

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

/*
 * This sample shows how to use the typed search attributes API.
 */

var (
	// CustomIntKey is the key for a custom int search attribute.
	CustomIntKey = temporal.NewSearchAttributeKeyInt64("CustomIntField")
	// CustomKeyword is the key for a custom keyword search attribute.
	CustomKeyword = temporal.NewSearchAttributeKeyString("CustomKeywordField")
	// CustomBool is the key for a custom bool search attribute.
	CustomBool = temporal.NewSearchAttributeKeyBool("CustomBoolField")
	// CustomDouble is the key for a custom double search attribute.
	CustomDouble = temporal.NewSearchAttributeKeyFloat64("CustomDoubleField")
	// CustomStringField is the key for a custom string search attribute.
	CustomStringField = temporal.NewSearchAttributeKeyString("CustomStringField")
	// CustomDatetimeField is the key for a custom datetime search attribute.
	CustomDatetimeField = temporal.NewSearchAttributeKeyTime("CustomDatetimeField")
)

// SearchAttributesWorkflow workflow definition
func SearchAttributesWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("SearchAttributes workflow started")

	// Get search attributes that were provided when workflow was started.
	searchAttributes := workflow.GetTypedSearchAttributes(ctx)
	currentIntValue, ok := searchAttributes.GetInt64(CustomIntKey)
	if !ok {
		return errors.New("CustomIntField is not set")
	}
	logger.Info("Current search attribute value.", "CustomIntField", currentIntValue)

	// Upsert search attributes.

	// This won't persist search attributes on server because commands are not sent to server,
	// but local cache will be updated.
	err := workflow.UpsertTypedSearchAttributes(ctx,
		CustomIntKey.ValueSet(2),
		CustomKeyword.ValueSet("Update1"),
		CustomBool.ValueSet(true),
		CustomDouble.ValueSet(3.14),
		CustomDatetimeField.ValueSet(workflow.Now(ctx).UTC()),
		CustomStringField.ValueSet("String field is for text. When query, it will be tokenized for partial match."),
	)
	if err != nil {
		return err
	}

	// Print current search attributes with modifications above.
	searchAttributes = workflow.GetTypedSearchAttributes(ctx)
	err = printSearchAttributes(searchAttributes, logger)
	if err != nil {
		return err
	}

	// Update search attributes again.
	err = workflow.UpsertTypedSearchAttributes(ctx,
		CustomKeyword.ValueSet("Update2"),
		CustomIntKey.ValueUnset(),
	)
	if err != nil {
		return err
	}

	// Sleep to allow update to be visible in search.
	err = workflow.Sleep(ctx, 1*time.Second)
	if err != nil {
		return err
	}

	// Print current search attributes.
	searchAttributes = workflow.GetTypedSearchAttributes(ctx)
	err = printSearchAttributes(searchAttributes, logger)
	if err != nil {
		return err
	}

	logger.Info("Workflow completed.")
	return nil
}

func printSearchAttributes(searchAttributes temporal.SearchAttributes, logger log.Logger) error {
	if searchAttributes.Size() == 0 {
		logger.Info("Current search attributes are empty.")
		return nil
	}

	var builder strings.Builder
	//workflowcheck:ignore Only iterates for logging reasons
	for k, v := range searchAttributes.GetUntypedValues() {
		builder.WriteString(fmt.Sprintf("%s=%v\n", k.GetName(), v))
	}
	logger.Info(fmt.Sprintf("Current search attribute values:\n%s", builder.String()))
	return nil
}
