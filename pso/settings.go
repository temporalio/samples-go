package pso

const pso_max_size int = 100
const pso_inertia float64 = 0.7298 // default value of w (see clerc02)

type SwarmSettings struct {
	FunctionName string
	function     ObjectiveFunction // lower case to avoid data converter export
	// swarm size (number of particles)
	Size int
	// ... N steps (set to 0 for no output)
	PrintEvery int
	// Steps after issuing a ContinueAsNew, to reduce history size
	ContinueAsNewEvery int
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

	Inertia float64 // current inertia weight value
}

func FunctionFactory(functionName string) ObjectiveFunction {
	var function ObjectiveFunction
	switch functionName {
	case "sphere":
		function = Sphere
	case "rosenbrock":
		function = Rosenbrock
	case "griewank":
		function = Griewank
	}
	return function
}

func PSODefaultSettings(functionName string) *SwarmSettings {
	settings := new(SwarmSettings)

	settings.FunctionName = functionName
	settings.function = FunctionFactory(functionName)

	settings.Size = CalculateSwarmSize(settings.function.dim, pso_max_size)
	settings.PrintEvery = 10
	settings.ContinueAsNewEvery = 10
	settings.Steps = 100000
	settings.C1 = 1.496
	settings.C2 = 1.496
	settings.InertiaMax = pso_inertia
	settings.InertiaMin = 0.3
	settings.Inertia = settings.InertiaMax

	settings.ClampPosition = true

	return settings
}
