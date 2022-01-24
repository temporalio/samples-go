package interceptor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/samples-go/interceptor"
	sdkinterceptor "go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func TestWorkflow(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	// Set our capturing logger
	var capturingLogger capturingLogger
	suite.SetLogger(&capturingLogger)
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterActivity(interceptor.Activity)

	// Add our interceptor to put a custom tag on the logs
	env.SetWorkerOptions(worker.Options{
		Interceptors: []sdkinterceptor.WorkerInterceptor{interceptor.NewWorkerInterceptor(interceptor.InterceptorOptions{
			GetExtraLogTagsForWorkflow: func(workflow.Context) []interface{} {
				return []interface{}{"workflow-tag", "workflow-value"}
			},
			GetExtraLogTagsForActivity: func(context.Context) []interface{} {
				return []interface{}{"activity-tag", "activity-value"}
			},
		})},
	})

	// Run workflow
	env.ExecuteWorkflow(interceptor.Workflow, "Temporal")
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Hello Temporal!", result)

	// Confirm logs
	require.True(t, capturingLogger.hasEntry("INFO", "HelloWorld workflow started", "workflow-tag", "workflow-value"))
	require.True(t, capturingLogger.hasEntry("INFO", "Activity", "activity-tag", "activity-value"))
}

type capturingLogger struct {
	entries []*logEntry
}

type logEntry struct {
	level string
	msg   string
	tags  []interface{}
}

func (c *capturingLogger) addEntry(level, msg string, tags []interface{}) {
	c.entries = append(c.entries, &logEntry{level, msg, tags})
}

func (c *capturingLogger) Debug(msg string, tags ...interface{}) { c.addEntry("DEBUG", msg, tags) }
func (c *capturingLogger) Info(msg string, tags ...interface{})  { c.addEntry("INFO", msg, tags) }
func (c *capturingLogger) Warn(msg string, tags ...interface{})  { c.addEntry("WARN", msg, tags) }
func (c *capturingLogger) Error(msg string, tags ...interface{}) { c.addEntry("ERROR", msg, tags) }

func (c *capturingLogger) hasEntry(level, msg string, atLeastTags ...interface{}) bool {
	for _, entry := range c.entries {
		if entry.level == level && entry.msg == msg {
			allTagsFound := true
			for i := 0; i < len(atLeastTags); i += 2 {
				if i+1 >= len(atLeastTags) || !entry.hasTag(atLeastTags[i], atLeastTags[i+1]) {
					allTagsFound = false
					break
				}
			}
			if allTagsFound {
				return true
			}
		}
	}
	return false
}

func (l *logEntry) hasTag(k, v interface{}) bool {
	for i := 0; i < len(l.tags); i += 2 {
		if l.tags[i] == k && i+1 < len(l.tags) && l.tags[i+1] == v {
			return true
		}
	}
	return false
}
