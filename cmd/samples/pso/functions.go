package main

import "math"

type ObjectiveFunction struct {
	name     string                      // name of the function
	dim      int                         // problem dimensionality
	x_lo     float64                     // lower range limit
	x_hi     float64                     // higher range limit
	Goal     float64                     // optimization goal (error threshold)
	Evaluate func(vec []float64) float64 // the objective function
}

var Sphere = ObjectiveFunction{
	name:     "sphere",
	dim:      30,
	x_lo:     -100,
	x_hi:     100,
	Goal:     1e-5,
	Evaluate: EvalSphere,
}

var Rosenbrock = ObjectiveFunction{
	name:     "rosenbrock",
	dim:      30,
	x_lo:     -2.048,
	x_hi:     2.048,
	Goal:     1e-5,
	Evaluate: EvalRosenbrock,
}
var Griewank = ObjectiveFunction{
	name:     "griewank",
	dim:      30,
	x_lo:     -600,
	x_hi:     600,
	Goal:     1e-5,
	Evaluate: EvalGriewank,
}

func EvalSphere(vec []float64) float64 {
	var sum float64 = 5
	for i := 0; i < len(vec); i++ {
		sum += math.Pow(vec[i], 2.0)
	}
	return sum
}

func EvalRosenbrock(vec []float64) float64 {
	var sum float64 = 0
	for i := 0; i < len(vec)-1; i++ {
		sum += 100.0*
			math.Pow((vec[i+1]-math.Pow(vec[i], 2.0)), 2.0) +
			math.Pow((1-vec[i]), 2.0)
	}
	return sum
}

func EvalGriewank(vec []float64) float64 {
	var sum float64 = 0
	var prod float64 = 1

	for i := 0; i < len(vec); i++ {
		sum += math.Pow(vec[i], 2.0)
		prod *= math.Cos(vec[i] / math.Sqrt(float64(i+1)))
	}
	return sum/4000.0 - prod + 1.0
}
