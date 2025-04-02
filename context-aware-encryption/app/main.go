package main

import (
	"context"
	"fmt"
	contextawareencryption "github.com/temporalio/samples-go/context-aware-encryption"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type startable interface {
	Start(context.Context) error
	Shutdown(context.Context)
}

func main() {
	ctx, done := context.WithCancel(context.Background())

	c := contextawareencryption.MustGetDefaultTemporalClient(ctx, nil)
	defer c.Close()
	g, ctx := errgroup.WithContext(ctx)

	// set up signal listener
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)

	startables := []startable{
		&Worker{tclient: c},
		&App{tclient: c, maxCount: 1},
	}

	for _, s := range startables {
		var current = s
		g.Go(func() error {
			if err := current.Start(ctx); err != nil {
				return err
			}
			return nil
		})
	}

	select {
	case <-quit:
		break
	case <-ctx.Done():
		break
	}

	// shutdown the things
	done()
	// limit how long we'll wait for
	timeoutCtx, timeoutCancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer timeoutCancel()

	for _, s := range startables {
		s.Shutdown(timeoutCtx)
	}

	// wait for shutdown
	if err := g.Wait(); err != nil {
		panic("shutdown was not clean" + err.Error())
	}
}

type Worker struct {
	tclient sdkclient.Client
	worker  worker.Worker
}

func (w *Worker) Start(ctx context.Context) error {
	wrk := worker.New(w.tclient, "encryption", worker.Options{})

	wrk.RegisterWorkflow(contextawareencryption.TenantWorkflow)
	wrk.RegisterActivity(contextawareencryption.TenantActivity)

	return wrk.Run(worker.InterruptCh())
}
func (w *Worker) Shutdown(ctx context.Context) {
	w.worker.Stop()
}

type App struct {
	tclient  sdkclient.Client
	maxCount int
}

func (a *App) Shutdown(ctx context.Context) {

}
func (a *App) Start(ctx context.Context) error {
	if a.maxCount == 0 {
		return fmt.Errorf("You must at least one run Workflow")
	}
	dt := time.Now().UTC().String()
	count := 0
	for tenant, keyId := range contextawareencryption.TenantKeysByOrganization {
		wid := fmt.Sprintf("tenant_%s-%s", tenant, dt)
		workflowOptions := sdkclient.StartWorkflowOptions{
			ID:        wid,
			TaskQueue: "encryption",
		}

		// If you are using a ContextPropagator and varying keys per workflow you need to set
		// the KeyID to use for this workflow in the context:
		fmt.Println(fmt.Sprintf("Setting encryption key for '%s' with value '%s'", tenant, keyId))
		ctx = context.WithValue(ctx,
			contextawareencryption.PropagateKey,
			contextawareencryption.CryptContext{KeyID: keyId})

		// The workflow input tenant will be encrypted by the DataConverter before being sent to Temporal
		we, err := a.tclient.ExecuteWorkflow(
			ctx,
			workflowOptions,
			contextawareencryption.TenantWorkflow,
			"workflowargument for "+tenant,
		)
		if err != nil {
			log.Fatalln("Unable to execute workflow", err)
		}

		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

		// Synchronously wait for the workflow completion.
		var result string
		err = we.Get(context.Background(), &result)
		if err != nil {
			log.Fatalln("Unable get workflow result", err)
		}
		log.Println("TenantWorkflow result:", result)
		count++
		if count >= a.maxCount {
			break
		}
	}
	return nil
}
