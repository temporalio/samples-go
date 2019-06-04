package main

import (
	"errors"
	"fmt"
	"strconv"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type Result struct {
	Position
	Step int
}

type Swarm struct {
	Settings  *SwarmSettings
	Gbest     *Position
	Particles []*Particle
}

func NewSwarm(ctx workflow.Context, settings *SwarmSettings) (*Swarm, error) {
	var swarm Swarm
	// store settings
	swarm.Settings = settings
	// initialize gbest
	swarm.Gbest = NewPosition(swarm.Settings.Function.dim)
	swarm.Gbest.Fitness = 1e20

	// initialize particles in parallel
	chunkResultChannel := workflow.NewChannel(ctx)
	swarm.Particles = make([]*Particle, settings.Size)
	for i := 0; i < swarm.Settings.Size; i++ {
		particleIdx := i
		workflow.Go(ctx, func(ctx workflow.Context) {
			var particle Particle
			err := workflow.ExecuteActivity(ctx, initParticleActivityName, swarm).Get(ctx, &particle)
			if err == nil {
				swarm.Particles[particleIdx] = &particle
			} else {
				//FATAL ERROR
			}
			chunkResultChannel.Send(ctx, err)
		})
	}

	// wait for all particles to be initialized
	for i := 0; i < swarm.Settings.Size; i++ {
		var v interface{}
		chunkResultChannel.Receive(ctx, &v)
		switch r := v.(type) {
		case error:
			if r != nil {
				return &swarm, r
			}
		}
	}

	swarm.updateBest()

	return &swarm, nil
}

func (swarm *Swarm) updateBest() {
	for i := 0; i < swarm.Settings.Size; i++ {
		if swarm.Particles[i].Pbest.IsBetterThan(swarm.Gbest) {
			swarm.Gbest = swarm.Particles[i].Pbest.Copy()
		}
	}
}

func (swarm *Swarm) Run(ctx workflow.Context, step int) (Result, error) {
	logger := workflow.GetLogger(ctx)

	// the algorithm goes here
	chunkResultChannel := workflow.NewChannel(ctx)
	for step <= swarm.Settings.Steps {
		logger.Info("Iteration ", zap.String("step", strconv.Itoa(step)))
		// Update particles in parallel
		for i := 0; i < swarm.Settings.Size; i++ {
			particleIdx := i
			workflow.Go(ctx, func(ctx workflow.Context) { // Use an activity for this whole block
				var particle Particle
				err := workflow.ExecuteActivity(ctx, updateParticleActivityName, *swarm, particleIdx).Get(ctx, &particle)
				if err == nil {
					swarm.Particles[particleIdx] = &particle
				} else {
					//FATAL ERROR
				}
				chunkResultChannel.Send(ctx, err)
			})
		}

		// Wait for all particles to be updated
		for i := 0; i < swarm.Settings.Size; i++ {
			var v interface{}
			chunkResultChannel.Receive(ctx, &v)
			switch r := v.(type) {
			case error:
				if r != nil {
					return Result{
						Position: *swarm.Gbest,
						Step:     step,
					}, r
				}
			}
		}

		logger.Debug("Iteration Update Swarm Best", zap.String("step", strconv.Itoa(step)))

		swarm.updateBest()

		// Check if the goal has reached then stop early
		if swarm.Gbest.Fitness < swarm.Settings.Function.Goal {
			logger.Debug("Iteration New Swarm Best", zap.String("step", strconv.Itoa(step)))
			return Result{
				Position: *swarm.Gbest,
				Step:     step,
			}, nil
		}

		if step%swarm.Settings.PrintEvery == 0 {
			msg := fmt.Sprintf("Step %d :: min err=%.5e\n", step, swarm.Gbest.Fitness)
			logger.Info(msg)
		}

		// Finished all iterations
		if step == swarm.Settings.Steps {
			break
		}

		// Not finished yet, just continue as new to reduce history size
		if step%swarm.Settings.ContinueAsNewEvery == 0 {
			return Result{
				Position: *swarm.Gbest,
				Step:     step,
			}, errors.New("CONTINUEASNEW")
		}

		step++
	}

	return Result{
		Position: *swarm.Gbest,
		Step:     step,
	}, nil
}
