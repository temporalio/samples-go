package main

import (
	"fmt"

	"go.uber.org/cadence/workflow"
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
	swarm.particles = make([]*Particle, settings.Function.dim)
	for i := 0; i < swarm.settings.Size; i++ {
		swarm.particles[i] = NewParticle(settings)

		swarm.particles[i].UpdateFitness(swarm.ctx)
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

func (swarm *Swarm) Run() Result {
	// the algorithm goes here
	var step int
	for step = 0; step < swarm.settings.Steps; step++ {
		for _, particle := range swarm.particles {
			particle.UpdateLocation(swarm.Gbest)
			particle.UpdateFitness(swarm.ctx)
		}

		// TODO (ali): we may need to put fitness updating of all particles in a separate
		// work flow to be able to evaluate fitness particles in parallel
		swarm.updateBest()

		// Check if the goal has reached then stop early
		if swarm.Gbest.Fitness < swarm.settings.Function.Goal {
			return Result{
				Position: *swarm.Gbest,
				Step:     step,
			}
		}

		if step%swarm.settings.PrintEvery == 0 {
			fmt.Printf("Step %d :: min err=%.5e\n", step, swarm.Gbest.Fitness)
		}
	}
	return Result{
		Position: *swarm.Gbest,
		Step:     step,
	}

}
