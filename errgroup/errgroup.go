package main

import (
	"errors"
	"log"
	"time"

	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context) (string, error) {
	g, cc := WithContext(ctx)

	for i := 0; i < 3; i++ {
		g.Go(cc, func(ctx workflow.Context) error {
			workflow.Sleep(ctx, 3*time.Second)
			if ctx.Err() != nil {
				println("ctx error", ctx.Err().Error())
			} else {
				panic("shouldn't be here")
			}
			return nil
		})
	}
	g.Go(cc, func(ctx workflow.Context) error {
		return errors.New("foo error")
	})

	err := g.Wait(cc)
	if err == nil {
		return "", errors.New("expected error")
	}
	return "expected error received: " + err.Error(), nil
}

func main() {
	s := &testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(Workflow)
	err := env.GetWorkflowError()
	if err != nil {
		log.Fatalf("[ERROR] not expecting workflow error: %v", err)
	}
	var result interface{}
	err = env.GetWorkflowResult(&result)
	if err != nil {
		log.Fatalf("[ERROR] failed to get workflow result: %v", err)
	}
	log.Printf("result: %v", result)
}

// adapted from https://cs.opensource.google/go/x/sync/+/036812b2:errgroup/errgroup.go
type ErrGroup struct {
	wg     workflow.WaitGroup
	cancel func()
	err    error
}

func WithContext(ctx workflow.Context) (*ErrGroup, workflow.Context) {
	cc, cancel := workflow.WithCancel(ctx)
	eg := &ErrGroup{
		cancel: cancel,
		wg:     workflow.NewWaitGroup(ctx),
	}
	return eg, cc
}

func (g *ErrGroup) Wait(ctx workflow.Context) error {
	g.wg.Wait(ctx)
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

func (g *ErrGroup) Go(ctx workflow.Context, f func(workflow.Context) error) {
	g.wg.Add(1)
	workflow.Go(ctx, func(ctx workflow.Context) {
		defer g.wg.Done()
		if err := f(ctx); err != nil {
			if g.err == nil {
				g.err = err
			}
			if g.cancel != nil {
				g.cancel()
			}
		}
	})
}
