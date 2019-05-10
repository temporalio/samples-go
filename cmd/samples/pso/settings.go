package main

import (
	"math/rand"
	"time"
)

const pso_max_size int = 100
const pso_inertia float64 = 0.7298 // default value of w (see clerc02)

type SwarmSettings struct {
	Function     ObjectiveFunction
	FunctionName string
	// swarm size (number of particles)
	Size int
	// ... N steps (set to 0 for no output)
	PrintEvery int
	// maximum number of iterations
	Steps int
	// cognitive coefficient
	C1 float64
	// social coefficient
	C2 float64
	// max inertia weight value
	InertiaMax float64
	// min inertia weight value
	InertiaMin float64
	// whether to keep particle position within defined bounds (TRUE)
	// or apply periodic boundary conditions (FALSE)
	ClampPosition bool

	inertia float64 // current inertia weight value
	step    int     // current PSO step

	rng *rand.Rand // the random number generator
}

func PSODefaultSettings(functionName string) *SwarmSettings {
	settings := new(SwarmSettings)

	settings.FunctionName = functionName
	switch settings.FunctionName {
	case "sphere":
		settings.Function = Sphere
	case "rosenbrock":
		settings.Function = Rosenbrock
	case "griewank":
		settings.Function = Griewank
	}

	settings.Size = CalculateSwarmSize(settings.Function.dim, pso_max_size)
	settings.PrintEvery = 10
	settings.Steps = 100000
	settings.C1 = 1.496
	settings.C2 = 1.496
	settings.InertiaMax = pso_inertia
	settings.InertiaMin = 0.3
	settings.inertia = settings.InertiaMax

	settings.ClampPosition = true

	// initialize the RNG
	settings.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	return settings
}
