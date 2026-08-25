package main

import (
	"context"
	"log"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	ubertally "github.com/uber-go/tally/v4"
	"github.com/uber-go/tally/v4/prometheus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	sdktally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/worker"

	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/model/gemini"

	"go.temporal.io/sdk/contrib/googleadk"

	metrics "github.com/temporalio/samples-go/googleadk/metrics"
)

func main() {
	options := envconfig.MustLoadDefaultClientOptions()
	reporter, err := (prometheus.Configuration{
		ListenAddress: "127.0.0.1:9090",
		TimerType:     "histogram",
	}).NewReporter(prometheus.ConfigurationOptions{
		Registry: prom.NewRegistry(),
		OnError:  func(err error) { log.Println("Prometheus reporter error", err) },
	})
	if err != nil {
		log.Fatalln("Unable to create Prometheus reporter", err)
	}
	scope, _ := ubertally.NewRootScope(ubertally.ScopeOptions{
		CachedReporter:  reporter,
		Separator:       prometheus.DefaultSeparator,
		SanitizeOptions: &sdktally.PrometheusSanitizeOptions,
	}, time.Second)
	options.MetricsHandler = sdktally.NewMetricsHandler(sdktally.NewPrometheusNamingScope(scope))
	c, err := client.Dial(options)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	plugin, err := googleadk.NewPlugin(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			metrics.ModelName: func(ctx context.Context, name string) (model.LLM, error) {
				return gemini.NewModel(ctx, name, nil)
			},
		},
	})
	if err != nil {
		log.Fatalln("Unable to build googleadk plugin", err)
	}

	w := worker.New(c, metrics.TaskQueue, worker.Options{Plugins: []worker.Plugin{plugin}})
	w.RegisterWorkflow(metrics.AgentWorkflow)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
