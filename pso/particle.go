package pso

import (
	"math/rand"
	"time"
)

type Particle struct {
	Position *Position
	Pbest    *Position
	Velocity Vector
}

func NewParticle(swarm *Swarm) *Particle {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	particle := new(Particle)
	particle.Position = RandomPosition(swarm.Settings.function, rng)

	particle.Pbest = particle.Position.Copy()
	particle.Pbest.Fitness = 1e20

	particle.Velocity = make([]float64, swarm.Settings.function.dim)
	xLo := swarm.Settings.function.xLo
	xHi := swarm.Settings.function.xHi
	for i := 0; i < swarm.Settings.function.dim; i++ {
		a := xLo + (xHi-xLo)*rng.Float64()
		b := xLo + (xHi-xLo)*rng.Float64()
		particle.Velocity[i] = (a - b) / 2.0
	}
	return particle
}

func (particle *Particle) UpdateLocation(swarm *Swarm) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < swarm.Settings.function.dim; i++ {
		// calculate stochastic coefficients
		rho1 := swarm.Settings.C1 * rng.Float64()
		rho2 := swarm.Settings.C2 * rng.Float64()
		// update velocity
		particle.Velocity[i] =
			swarm.Settings.Inertia*particle.Velocity[i] +
				rho1*(particle.Pbest.Location[i]-particle.Position.Location[i]) +
				rho2*(swarm.Gbest.Location[i]-particle.Position.Location[i])

		particle.Position.Location[i] += particle.Velocity[i]
	}
}

func (particle *Particle) UpdateFitness(swarm *Swarm) {
	particle.Position.Fitness = swarm.Settings.function.Evaluate(particle.Position.Location)

	if particle.Position.IsBetterThan(particle.Pbest) {
		particle.Pbest = particle.Position.Copy()
	}
}
