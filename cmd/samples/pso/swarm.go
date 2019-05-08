package main

import (
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
	ctx       workflow.Context
	Gbest     *Position
	settings  *SwarmSettings
	particles []*Particle
}

func NewSwarm(ctx workflow.Context, settings *SwarmSettings) *Swarm {
	var swarm Swarm
	swarm.ctx = ctx
	// store settings
	swarm.settings = settings
	// initialize gbest
	swarm.Gbest = NewPosition(settings)
	swarm.Gbest.Fitness = 1e20
	// initialize particles
	swarm.particles = make([]*Particle, settings.Size)
	for i := 0; i < swarm.settings.Size; i++ {
		swarm.particles[i] = NewParticle(settings)

		swarm.particles[i].UpdateFitness(swarm.ctx)
		// if err != nil {
		// 	return Result{
		// 		Position: *swarm.Gbest,
		// 		Step:     step,
		// 	}, err
		// }
	}
	swarm.updateBest()

	return &swarm
}

func (swarm *Swarm) updateBest() {
	for i := 0; i < swarm.settings.Size; i++ {
		if swarm.particles[i].pbest.IsBetterThan(swarm.Gbest) {
			swarm.Gbest = swarm.particles[i].pbest.Copy()
		}
	}
}

func (swarm *Swarm) Run() (Result, error) {
	// the algorithm goes here
	var step int
	for step = 0; step < swarm.settings.Steps; step++ {
		workflow.GetLogger(swarm.ctx).Info("Iteration ", zap.String("step", strconv.Itoa(step)))
		for _, particle := range swarm.particles {
			particle.UpdateLocation(swarm.Gbest)
			err := particle.UpdateFitness(swarm.ctx)
			if err != nil {
				return Result{
					Position: *swarm.Gbest,
					Step:     step,
				}, err
			}
		}

		workflow.GetLogger(swarm.ctx).Debug("Iteration Update Swarm Best", zap.String("step", strconv.Itoa(step)))
		// TODO (ali): we may need to put fitness updating of all particles in a separate
		// work flow to be able to evaluate fitness particles in parallel
		swarm.updateBest()

		// Check if the goal has reached then stop early
		if swarm.Gbest.Fitness < swarm.settings.Function.Goal {
			workflow.GetLogger(swarm.ctx).Debug("Iteration New Swarm Best", zap.String("step", strconv.Itoa(step)))
			return Result{
				Position: *swarm.Gbest,
				Step:     step,
			}, nil
		}

		if step%swarm.settings.PrintEvery == 0 {
			msg := fmt.Sprintf("Step %d :: min err=%.5e\n", step, swarm.Gbest.Fitness)
			workflow.GetLogger(swarm.ctx).Info(msg)
		}
	}
	return Result{
		Position: *swarm.Gbest,
		Step:     step,
	}, nil
}
