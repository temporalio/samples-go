package earlyreturn

import (
	"fmt"
	"testing"

	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_CompleteTransaction(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	tx := Transaction{ID: uuid.New(), SourceAccount: "Bob", TargetAccount: "Alice", Amount: 100}
	env.RegisterActivity(tx.InitTransaction)
	env.RegisterActivity(tx.CompleteTransaction)

	uc := &updateCallback{}
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), uc)
	}, 0)
	env.ExecuteWorkflow(Workflow, tx)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.NoError(t, uc.completeErr)
}

func Test_CompleteTransaction_Fails(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	tx := Transaction{ID: uuid.New(), SourceAccount: "Bob", TargetAccount: "Alice", Amount: 100}
	env.RegisterActivity(tx.InitTransaction)
	env.RegisterActivity(tx.CompleteTransaction)

	env.OnActivity(tx.CompleteTransaction, mock.Anything).Return(fmt.Errorf("crash"))

	uc := &updateCallback{}
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), uc)
	}, 0)
	env.ExecuteWorkflow(Workflow, tx)

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, env.GetWorkflowError(), "crash")
}

func Test_CancelTransaction(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	tx := Transaction{ID: uuid.New(), SourceAccount: "Bob", TargetAccount: "Alice", Amount: -1} // invalid!
	env.RegisterActivity(tx.InitTransaction)
	env.RegisterActivity(tx.CancelTransaction)

	uc := &updateCallback{}
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), uc)
	}, 0)
	env.ExecuteWorkflow(Workflow, tx)

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, uc.completeErr, "invalid Amount")
	require.ErrorContains(t, env.GetWorkflowError(), "invalid Amount")
}

func Test_CancelTransaction_Fails(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	tx := Transaction{ID: uuid.New(), SourceAccount: "Bob", TargetAccount: "Alice", Amount: -1} // invalid!
	env.RegisterActivity(tx.InitTransaction)
	env.RegisterActivity(tx.CancelTransaction)

	env.OnActivity(tx.CancelTransaction, mock.Anything).Return(fmt.Errorf("crash"))

	uc := &updateCallback{}
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), uc)
	}, 0)
	env.ExecuteWorkflow(Workflow, tx)

	require.True(t, env.IsWorkflowCompleted())
	require.ErrorContains(t, uc.completeErr, "invalid Amount")
	require.ErrorContains(t, env.GetWorkflowError(), "crash")
}

type updateCallback struct {
	completeErr error
}

func (uc *updateCallback) Accept() {}

func (uc *updateCallback) Reject(err error) {}

func (uc *updateCallback) Complete(success interface{}, err error) {
	uc.completeErr = err
}
