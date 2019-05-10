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

func NewSwarm(ctx workflow.Context, settings *SwarmSettings) (*Swarm, error) {
	var swarm Swarm
	swarm.ctx = ctx
	// store settings
	swarm.settings = settings
	// initialize gbest
	swarm.Gbest = NewPosition(settings)
	swarm.Gbest.Fitness = 1e20

	// initialize particles in parallel
	chunkResultChannel := workflow.NewChannel(swarm.ctx)
	swarm.particles = make([]*Particle, settings.Size)
	for i := 0; i < swarm.settings.Size; i++ {
		particleIdx := i
		workflow.Go(swarm.ctx, func(ctx workflow.Context) {
			swarm.particles[particleIdx] = NewParticle(settings)
			err := swarm.particles[particleIdx].UpdateFitness(ctx)
			chunkResultChannel.Send(ctx, err)
		})
	}

	// wait for all particles to be initialized
	for i := 0; i < swarm.settings.Size; i++ {
		var v interface{}
		chunkResultChannel.Receive(swarm.ctx, &v)
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
	for i := 0; i < swarm.settings.Size; i++ {
		if swarm.particles[i].pbest.IsBetterThan(swarm.Gbest) {
			swarm.Gbest = swarm.particles[i].pbest.Copy()
		}
	}
}

func (swarm *Swarm) Run() (Result, error) {
	// the algorithm goes here
	chunkResultChannel := workflow.NewChannel(swarm.ctx)
	var step int
	for step = 0; step < swarm.settings.Steps; step++ {
		workflow.GetLogger(swarm.ctx).Info("Iteration ", zap.String("step", strconv.Itoa(step)))
		// Update particles in parallel
		for i := 0; i < swarm.settings.Size; i++ {
			particleIdx := i
			workflow.Go(swarm.ctx, func(ctx workflow.Context) {
				swarm.particles[particleIdx].UpdateLocation(swarm.Gbest)
				err := swarm.particles[particleIdx].UpdateFitness(ctx)
				chunkResultChannel.Send(ctx, err)
			})
		}

		// Wait for all particles to be updated
		for i := 0; i < swarm.settings.Size; i++ {
			var v interface{}
			chunkResultChannel.Receive(swarm.ctx, &v)
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

		workflow.GetLogger(swarm.ctx).Debug("Iteration Update Swarm Best", zap.String("step", strconv.Itoa(step)))

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
