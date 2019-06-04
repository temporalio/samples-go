package main

import "math/rand"

type Particle struct {
	Position *Position
	Pbest    *Position
	Velocity Vector
}

func NewParticle(swarm *Swarm, rng *rand.Rand) *Particle {
	particle := new(Particle)
	particle.Position = RandomPosition(swarm.Settings.Function, rng)

	particle.Pbest = particle.Position.Copy()
	particle.Pbest.Fitness = 1e20

	particle.Velocity = make([]float64, swarm.Settings.Function.dim)
	xLo := swarm.Settings.Function.xLo
	xHi := swarm.Settings.Function.xHi
	for i := 0; i < swarm.Settings.Function.dim; i++ {
		a := xLo + (xHi-xLo)*rng.Float64()
		b := xLo + (xHi-xLo)*rng.Float64()
		particle.Velocity[i] = (a - b) / 2.0
	}
	return particle
}

func (particle *Particle) UpdateLocation(swarm *Swarm, rng *rand.Rand) {
	for i := 0; i < swarm.Settings.Function.dim; i++ {
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
	particle.Position.Fitness = swarm.Settings.Function.Evaluate(particle.Position.Location)

	//logger.Info(fmt.Sprintf("Particle fitness=%f", particle.position.Fitness))

	if particle.Position.IsBetterThan(particle.Pbest) {
		particle.Pbest = particle.Position.Copy()
	}
}
