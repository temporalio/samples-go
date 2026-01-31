package dataconverter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	dclp "github.com/temporalio/samples-go/data-converter-large-payloads"
)

// testSequentialIDGenerator is a predictable ID generator, useful in tests to
// validate the files being created by the data converter.
type testSequentialIDGenerator struct {
	start int
}

// NewString returns a file name with an integer suffix that is incremented
// every call.
func (t *testSequentialIDGenerator) NewString() string {
	t.start++
	return fmt.Sprintf("file-%02d", t.start)
}

// withSequentialIDGenerator is a helper function used to set the id
// generator to a predictable one in tests.
func withSequentialIDGenerator() LargeSizePayloadConverterOption {
	return func(c *LargeSizePayloadConverter) {
		c.idGenerator = &testSequentialIDGenerator{start: 0}
	}
}

type LargePayloadsDataConverterTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func TestLargePayloadsDataConverterTestSuite(t *testing.T) {
	suite.Run(t, new(LargePayloadsDataConverterTestSuite))
}

func (s *LargePayloadsDataConverterTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *LargePayloadsDataConverterTestSuite) TearDownTest() {
	// remove all the files that were created by the data converter
	dir, err := os.ReadDir(".")
	s.Require().NoError(err)
	for _, d := range dir {
		s.T().Logf("removig %s", d.Name())
		if strings.HasPrefix(d.Name(), "file-") {
			s.Require().NoError(os.Remove(d.Name()))
		}
	}
}

func (s *LargePayloadsDataConverterTestSuite) Test_DoesNotUseDataConverter_WhenNotExceedingThreshold() {

	// Ensure threshold will never be exceeded.
	s.env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		NewLargeSizePayloadConverter(WithThreshold(1000*1000*1000), withSequentialIDGenerator()),
		converter.NewJSONPayloadConverter(),
	))

	dummyActivity := func(ctx context.Context, ainput *dclp.ValueContainer) (*dclp.ValueContainer, error) {
		return ainput, nil
	}
	dummyWorkflow := func(ctx workflow.Context, winput *dclp.ValueContainer) (*dclp.ValueContainer, error) {
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)
		var result *dclp.ValueContainer
		err := workflow.ExecuteActivity(ctx, dummyActivity, winput).Get(ctx, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	s.env.RegisterWorkflow(dummyWorkflow)
	s.env.RegisterActivity(dummyActivity)

	input := dclp.ValueContainer{
		Values: []dclp.CustomData{
			{
				Index: 1,
				Name:  "name",
			},
		},
	}
	s.env.ExecuteWorkflow(dummyWorkflow, &input)

	// then
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	s.checkNoFilesWereCreated()
}

func (s *LargePayloadsDataConverterTestSuite) Test_UsesDataConverter_WhenExceedingThreshold() {

	// Ensure threshold will always be exceeded.
	s.env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		NewLargeSizePayloadConverter(WithThreshold(1), withSequentialIDGenerator()),
		converter.NewJSONPayloadConverter(),
	))

	workflowInput := dclp.ValueContainer{
		Values: []dclp.CustomData{
			{
				Name:  "workflow input",
				Index: 0,
			},
		},
	}
	activityInput := dclp.ValueContainer{
		Values: []dclp.CustomData{
			{
				Name:  "activity input",
				Index: 1,
			},
		},
	}
	workflowResult := dclp.ValueContainer{
		Values: []dclp.CustomData{
			{
				Name:  "workflow result",
				Index: 2,
			},
		},
	}
	activityResult := dclp.ValueContainer{
		Values: []dclp.CustomData{
			{
				Name:  "activity result",
				Index: 3,
			},
		},
	}

	dummyActivity := func(ctx context.Context, ainput *dclp.ValueContainer) (*dclp.ValueContainer, error) {
		return &activityResult, nil
	}
	dummyWorkflow := func(ctx workflow.Context, winput *dclp.ValueContainer) (*dclp.ValueContainer, error) {
		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
		}
		ctx = workflow.WithActivityOptions(ctx, ao)
		err := workflow.ExecuteActivity(ctx, dummyActivity, &activityInput).Get(ctx, nil)
		if err != nil {
			return nil, err
		}
		return &workflowResult, nil
	}

	s.env.RegisterWorkflow(dummyWorkflow)
	s.env.RegisterActivity(dummyActivity)

	s.env.ExecuteWorkflow(dummyWorkflow, &workflowInput)

	// then
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())

	// The following files should be created:
	// file-01 for calling the workflow
	// file-02 for calling the activity
	// file-03 for the activity return
	// file-04 for the workflow return
	s.checkOnlyFilesWereCreated("file-01", "file-02", "file-03", "file-04")
}

func (s *LargePayloadsDataConverterTestSuite) checkNoFilesWereCreated() {
	s.checkOnlyFilesWereCreated()
}

// checkOnlyFilesWereCreated validates that only the file names passed as
// argument were created by the data converter.
func (s *LargePayloadsDataConverterTestSuite) checkOnlyFilesWereCreated(files ...string) {
	s.T().Helper()

	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		s.T().Log("reading file", info.Name())

		// skip current directory and go files in this directory
		if path == "." || path[len(path)-3:] == ".go" {
			return nil
		}

		if len(files) == 0 || !slices.Contains(files, path) {
			return fmt.Errorf("unexpected path %s encountered", path)
		}

		return nil
	})
	s.Require().NoError(err)
}

// checkFileContent checks if the correct data is written to the file
func (s *LargePayloadsDataConverterTestSuite) checkFileContent(file string, expected *dclp.ValueContainer) {
	s.T().Helper()

	f, err := os.Open(file)
	s.Require().NoError(err)
	defer f.Close()

	d := json.NewDecoder(f)
	var content *dclp.ValueContainer
	err = d.Decode(&content)
	s.Require().NoError(err)

	s.Require().NotNil(content)
	s.Require().Equal(len(expected.Values), len(content.Values))

	for i := range len(expected.Values) {
		s.Assert().Equal(expected.Values[i], content.Values[i])
	}
}
