package pso

import (
	"context"

	"go.temporal.io/sdk/activity"
)

/**
 * Sample activities used by file processing sample workflow.
 */
const (
	InitParticleActivityName   = "initParticleActivityName"
	UpdateParticleActivityName = "updateParticleActivityName"
)

func InitParticleActivity(ctx context.Context, swarm Swarm) (Particle, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("initParticleActivity started.")

	particle := NewParticle(&swarm)
	particle.UpdateFitness(&swarm)

	return *particle, nil
}

func UpdateParticleActivity(ctx context.Context, swarm Swarm, particleIdx int) (Particle, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("updateParticleActivity started.")

	particle := swarm.Particles[particleIdx]
	particle.UpdateLocation(&swarm)
	particle.UpdateFitness(&swarm)

	return *particle, nil
}
