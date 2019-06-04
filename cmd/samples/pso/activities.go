package main

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

/**
 * Sample activities used by file processing sample workflow.
 */
const (
	initParticleActivityName   = "initParticleActivityName"
	updateParticleActivityName = "updateParticleActivityName"
)

var rng *rand.Rand

// This is registration process where you register all your activity handlers.
func init() {
	// initialize the RNG
	// WARNING: the randomness of activity scheduling with multiple workers makes random number generation truly random and not repeatable in debugging
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	activity.RegisterWithOptions(
		initParticleActivity,
		activity.RegisterOptions{Name: initParticleActivityName},
	)
	activity.RegisterWithOptions(
		updateParticleActivity,
		activity.RegisterOptions{Name: updateParticleActivityName},
	)
}

func initParticleActivity(ctx context.Context, swarm Swarm) (Particle, error) {
	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
	logger.Info("initParticleActivity started.")

	particle := NewParticle(&swarm, rng)
	particle.UpdateFitness(&swarm)

	return *particle, nil
}

func updateParticleActivity(ctx context.Context, swarm Swarm, particleIdx int) (Particle, error) {
	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
	logger.Info("updateParticleActivity started.")

	particle := swarm.Particles[particleIdx]
	particle.UpdateLocation(&swarm, rng)
	particle.UpdateFitness(&swarm)

	return *particle, nil
}
