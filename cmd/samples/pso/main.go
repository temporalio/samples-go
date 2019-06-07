package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"reflect"
	"time"

	"github.com/pborman/uuid"
	"github.com/uber/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/worker"
)

// myDataConverter implements encoded.DataConverter using gob for Swarm and Particle
// WARGNING: Make sure all struct members are public (Capital letter) otherwise serialization does not work!
// TODO: consider storing blobs in external DB or Forge
type myDataConverter struct {
}

func (dc *myDataConverter) ToData(value ...interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	var err error
	for i, obj := range value {
		switch t := obj.(type) {
		case Swarm:
			err = enc.Encode(*t.Settings)
			err = enc.Encode(*t.Gbest)
			if t.Settings.Size > 0 {
				for _, particle := range t.Particles {
					if particle == nil {
						particle = new(Particle)
					}
					err = enc.Encode(*particle)
				}
			}
		default:
			err = enc.Encode(obj)
		}
		if err != nil {
			return nil, fmt.Errorf(
				"unable to encode argument: %d, %v, with gob error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return buf.Bytes(), nil
	// TODO: store buf.Bytes() in DB/Forge and get key
	// return key, nil
}

func (dc *myDataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	// TODO: convert input into key in DB/Forge and retrieve bytes
	//dec := gob.NewDecoder(bytes)
	dec := gob.NewDecoder(bytes.NewBuffer(input))
	var err error
	for i, obj := range valuePtr {
		switch t := obj.(type) {
		case *Swarm:
			t.Settings = new(SwarmSettings)
			err = dec.Decode(t.Settings)
			t.Settings.Function = FunctionFactory(t.Settings.FunctionName)
			t.Gbest = NewPosition(t.Settings.Function.dim)
			err = dec.Decode(t.Gbest)
			t.Particles = make([]*Particle, t.Settings.Size)
			for index := 0; index < t.Settings.Size; index++ {
				t.Particles[index] = new(Particle)
				err = dec.Decode(t.Particles[index])
			}
		default:
			err = dec.Decode(obj)
		}
		if err != nil {
			return fmt.Errorf(
				"unable to decode argument: %d, %v, with gob error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}

func newMyDataConverter() encoded.DataConverter {
	return &myDataConverter{}
}

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope:          h.Scope,
		Logger:                h.Logger,
		EnableLoggingInReplay: true,
		DataConverter:         h.DataConverter,
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)

	// Host Specific activities processing case
	workerOptions.DisableWorkflowWorker = true
	h.StartWorkers(h.Config.DomainName, HostID, workerOptions)
}

func startWorkflow(h *common.SampleHelper, functionName string) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "PSO_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute * 60,
		DecisionTaskStartToCloseTimeout: time.Minute, // measure of responsiveness of the worker to various server signals apart from start workflow
	}
	h.StartWorkflow(workflowOptions, PSOWorkflow, functionName)
}

func main() {
	var mode string
	var functionName string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.StringVar(&functionName, "f", "sphere", "One of [sphere, rosenbrock, griewank]")
	flag.Parse()

	gob.Register(Vector{})
	gob.Register(Position{})
	gob.Register(Particle{})
	gob.Register(ObjectiveFunction{})
	gob.Register(SwarmSettings{})
	gob.Register(Swarm{})

	var h common.SampleHelper
	h.DataConverter = newMyDataConverter()
	h.SetupServiceConfig() // This configures DataConverter

	switch mode {
	case "worker":
		startWorkers(&h)

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger":
		startWorkflow(&h, functionName)
	}
}
