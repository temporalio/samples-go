package pso

import (
	"errors"
	"fmt"

	"go.temporal.io/sdk/workflow"
)

type ParticleResult struct {
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
	swarm.Gbest = NewPosition(swarm.Settings.function.dim)
	swarm.Gbest.Fitness = 1e20

	// initialize particles in parallel
	chunkResultChannel := workflow.NewChannel(ctx)
	swarm.Particles = make([]*Particle, settings.Size)
	for i := 0; i < swarm.Settings.Size; i++ {
		particleIdx := i
		workflow.Go(ctx, func(ctx workflow.Context) {
			var particle Particle
			err := workflow.ExecuteActivity(ctx, InitParticleActivityName, swarm).Get(ctx, &particle)
			if err == nil {
				swarm.Particles[particleIdx] = &particle
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

func (swarm *Swarm) Run(ctx workflow.Context, step int) (ParticleResult, error) {
	logger := workflow.GetLogger(ctx)

	// Setup query handler for query type "iteration"
	var iterationMessage string
	err := workflow.SetQueryHandler(ctx, "iteration", func(input []byte) (string, error) {
		return iterationMessage, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return ParticleResult{}, err
	}

	// the algorithm goes here
	chunkResultChannel := workflow.NewChannel(ctx)
	for step <= swarm.Settings.Steps {
		logger.Info("Iteration ", "step", step)
		// Update particles in parallel
		for i := 0; i < swarm.Settings.Size; i++ {
			particleIdx := i
			workflow.Go(ctx, func(ctx workflow.Context) {
				var particle Particle
				err := workflow.ExecuteActivity(ctx, UpdateParticleActivityName, *swarm, particleIdx).Get(ctx, &particle)
				if err == nil {
					swarm.Particles[particleIdx] = &particle
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
					return ParticleResult{
						Position: *swarm.Gbest,
						Step:     step,
					}, r
				}
			}
		}

		logger.Debug("Iteration Update Swarm Best", "step", step)

		swarm.updateBest()

		// Check if the goal has reached then stop early
		if swarm.Gbest.Fitness < swarm.Settings.function.Goal {
			logger.Debug("Iteration New Swarm Best", "step", step)
			return ParticleResult{
				Position: *swarm.Gbest,
				Step:     step,
			}, nil
		}

		iterationMessage = fmt.Sprintf("Step %d :: min err=%.5e\n", step, swarm.Gbest.Fitness)
		if step%swarm.Settings.PrintEvery == 0 {
			logger.Info(iterationMessage)
		}

		// Finished all iterations
		if step == swarm.Settings.Steps {
			break
		}

		// Not finished yet, just continue as new to reduce history size
		if step%swarm.Settings.ContinueAsNewEvery == 0 {
			return ParticleResult{
				Position: *swarm.Gbest,
				Step:     step,
			}, errors.New(ContinueAsNewStr)
		}

		step++
	}

	return ParticleResult{
		Position: *swarm.Gbest,
		Step:     step,
	}, nil
}
