package accumulator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

// test 0: send nothing, verify it times out but is successful
func (s *UnitTestSuite) Test_WorkflowTimeout() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(ComposeGreeting)
	bucket := "blue"
	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	// Workflow times out
	s.NoError(env.GetWorkflowError())

	var result string

	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello")
	s.Contains(result, "(0)")
}

// test 1: start workflow, send one signal, make sure one is accepted
func (s *UnitTestSuite) Test_Signal() {
	env := s.NewTestWorkflowEnvironment()

	bucket := "purple"
	env.RegisterActivity(ComposeGreeting)
	env.RegisterDelayedCallback(func() {
		suzieGreeting := new(AccumulateGreeting)
		suzieGreeting.GreetingText = "Suzie Robot"
		suzieGreeting.Bucket = bucket
		suzieGreeting.GreetingKey = "11235813"
		env.SignalWorkflow("greeting", suzieGreeting)
	}, time.Second*5)

	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello (1)")
	s.Contains(result, "Suzie Robot")
}

// test 2: just send an exit signal, should end quickly and return empty string
func (s *UnitTestSuite) Test_Exit() {
	env := s.NewTestWorkflowEnvironment()
	bucket := "purple"
	env.RegisterActivity(ComposeGreeting)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("exit", "")
	}, time.Second*5)

	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello")
	s.Contains(result, "(0)")
}

// test 3: send multiple greetings, should get them all
func (s *UnitTestSuite) Test_Multiple_Signals() {
	env := s.NewTestWorkflowEnvironment()

	bucket := "purple"
	env.RegisterActivity(ComposeGreeting)
	env.RegisterDelayedCallback(func() {
		suzieGreeting := new(AccumulateGreeting)
		suzieGreeting.GreetingText = "Suzie Robot"
		suzieGreeting.Bucket = bucket
		suzieGreeting.GreetingKey = "11235813"
		env.SignalWorkflow("greeting", suzieGreeting)
	}, time.Second*5)
	env.RegisterDelayedCallback(func() {
		hezekiahGreeting := new(AccumulateGreeting)
		hezekiahGreeting.GreetingText = "Hezekiah Robot"
		hezekiahGreeting.Bucket = bucket
		hezekiahGreeting.GreetingKey = "11235"
		env.SignalWorkflow("greeting", hezekiahGreeting)
	}, time.Second*6)

	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello (2)")
	s.Contains(result, "Suzie Robot")
	s.Contains(result, "Hezekiah Robot")
}

// test 4: send greetings with duplicate keys, should only get one
func (s *UnitTestSuite) Test_Duplicate_Signals() {
	env := s.NewTestWorkflowEnvironment()

	bucket := "purple"
	env.RegisterActivity(ComposeGreeting)
	env.RegisterDelayedCallback(func() {
		suzieGreeting := new(AccumulateGreeting)
		suzieGreeting.GreetingText = "Suzie Robot"
		suzieGreeting.Bucket = bucket
		suzieGreeting.GreetingKey = "11235813"
		env.SignalWorkflow("greeting", suzieGreeting)
	}, time.Second*5)
	env.RegisterDelayedCallback(func() {
		hezekiahGreeting := new(AccumulateGreeting)
		hezekiahGreeting.GreetingText = "Hezekiah Robot"
		hezekiahGreeting.Bucket = bucket
		hezekiahGreeting.GreetingKey = "11235813"
		env.SignalWorkflow("greeting", hezekiahGreeting)
	}, time.Second*6)

	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello (1)")
	s.Contains(result, "Suzie Robot")
	s.NotContains(result, "Hezekiah Robot")
}

// test 5: test sent with a bad bucket key
func (s *UnitTestSuite) Test_Bad_Bucket() {
	env := s.NewTestWorkflowEnvironment()

	bucket := "purple"
	env.RegisterActivity(ComposeGreeting)
	env.RegisterDelayedCallback(func() {
		suzieGreeting := new(AccumulateGreeting)
		suzieGreeting.GreetingText = "Jake Robot"
		suzieGreeting.Bucket = "blue"
		suzieGreeting.GreetingKey = "11235813"
		env.SignalWorkflow("greeting", suzieGreeting)
	}, time.Second*5)
	env.RegisterDelayedCallback(func() {
		hezekiahGreeting := new(AccumulateGreeting)
		hezekiahGreeting.GreetingText = "Hezekiah Robot"
		hezekiahGreeting.Bucket = bucket
		hezekiahGreeting.GreetingKey = "112358"
		env.SignalWorkflow("greeting", hezekiahGreeting)
	}, time.Second*6)

	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello (1)")
	s.NotContains(result, "Jake Robot")
	s.Contains(result, "Hezekiah Robot")
}

// test 6: test signal with start
func (s *UnitTestSuite) Test_Signal_With_Start() {
	env := s.NewTestWorkflowEnvironment()

	bucket := "purple"
	env.RegisterActivity(ComposeGreeting)
	env.RegisterDelayedCallback(func() {
		androssGreeting := new(AccumulateGreeting)
		androssGreeting.GreetingText = "Andross Robot"
		androssGreeting.Bucket = bucket
		androssGreeting.GreetingKey = "1123"
		env.SignalWorkflow("greeting", androssGreeting)
	}, time.Second*0)

	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello (1)")
	s.Contains(result, "Andross Robot")
}

// test 7: signal with start, wait too long for first workflow so it times out, send another signal, should be just one greeting
func (s *UnitTestSuite) Test_Signal_With_Start_Wait_Too_Long() {
	env := s.NewTestWorkflowEnvironment()

	bucket := "purple"
	env.RegisterActivity(ComposeGreeting)

	env.RegisterDelayedCallback(func() {
		johnGreeting := new(AccumulateGreeting)
		johnGreeting.GreetingText = "John Robot"
		johnGreeting.Bucket = bucket
		johnGreeting.GreetingKey = "112"
		env.SignalWorkflow("greeting", johnGreeting)
	}, time.Second*0)
	env.RegisterDelayedCallback(func() {
		targeGreeting := new(AccumulateGreeting)
		targeGreeting.GreetingText = "Targe Robot"
		targeGreeting.Bucket = bucket
		targeGreeting.GreetingKey = "11"
		env.SignalWorkflow("greeting", targeGreeting)
	}, fromStartTimeout+time.Second*1)

	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello (1)")
	s.Contains(result, "John Robot")
}

// test 8: signal with start, don't wait too long for first workflow so it times out, send another signal, should be two greetings
func (s *UnitTestSuite) Test_Signal_With_Start_Wait_Too_Short() {
	env := s.NewTestWorkflowEnvironment()

	bucket := "purple"
	env.RegisterActivity(ComposeGreeting)

	env.RegisterDelayedCallback(func() {
		johnGreeting := new(AccumulateGreeting)
		johnGreeting.GreetingText = "John Robot"
		johnGreeting.Bucket = bucket
		johnGreeting.GreetingKey = "112"
		env.SignalWorkflow("greeting", johnGreeting)
	}, time.Second*0)
	env.RegisterDelayedCallback(func() {
		targeGreeting := new(AccumulateGreeting)
		targeGreeting.GreetingText = "Targe Robot"
		targeGreeting.Bucket = bucket
		targeGreeting.GreetingKey = "11"
		env.SignalWorkflow("greeting", targeGreeting)
	}, signalToSignalTimeout-time.Second*1)

	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello (2)")
	s.Contains(result, "John Robot")
	s.Contains(result, "Targe Robot")
}

// test 9: signal with start, send exit, then signal with start, should get one signal
func (s *UnitTestSuite) Test_Signal_After_Exit() {
	env := s.NewTestWorkflowEnvironment()

	bucket := "purple"
	env.RegisterActivity(ComposeGreeting)

	env.RegisterDelayedCallback(func() {
		johnGreeting := new(AccumulateGreeting)
		johnGreeting.GreetingText = "John Robot"
		johnGreeting.Bucket = bucket
		johnGreeting.GreetingKey = "112"
		env.SignalWorkflow("greeting", johnGreeting)
	}, time.Second*0)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("exit", "")
	}, time.Second*5)

	env.RegisterDelayedCallback(func() {
		targeGreeting := new(AccumulateGreeting)
		targeGreeting.GreetingText = "Targe Robot"
		targeGreeting.Bucket = bucket
		targeGreeting.GreetingKey = "11"
		env.SignalWorkflow("greeting", targeGreeting)
	}, time.Second*5+time.Millisecond*1)

	greetings := GreetingsInfo{
		BucketKey:          bucket,
		GreetingsList:      []AccumulateGreeting{},
		UniqueGreetingKeys: make(map[string]bool),
	}
	env.ExecuteWorkflow(AccumulateSignalsWorkflow, greetings)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result string
	s.NoError(env.GetWorkflowResult(&result))
	s.Contains(result, "Hello (2)")
	s.Contains(result, "John Robot")
	s.Contains(result, "Targe Robot")
}
