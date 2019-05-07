package main

//import "go.uber.org/cadence/workflow"

type Vector []float64

type Position struct {
	Location Vector
	Fitness  float64
	settings *SwarmSettings
}

func NewPosition(settings *SwarmSettings) *Position {
	loc := make([]float64, settings.Function.dim)
	return &Position{
		Location: loc,
		// Fitness:  EvaluateFunction(settings.Function.Evaluate, loc),
		settings: settings,
	}
}

func RandomPosition(settings *SwarmSettings) *Position {
	pos := NewPosition(settings)
	x_lo := settings.Function.x_lo
	x_hi := settings.Function.x_hi
	for i := 0; i < len(pos.Location); i++ {
		pos.Location[i] = x_lo + (x_hi-x_lo)*settings.rng.Float64()
	}
	// pos.Fitness = EvaluateFunction(settings.Function.Evaluate, pos.Location)
	return pos
}

// func (position *Position) UpdateFitness() {

// 	//position.Fitness = EvaluateFunction(position.settings.Function.Evaluate, position.Location)
// 	//Call/Execute activity
// 	err = workflow.ExecuteActivity(hCtx, EvaluateFitnessActivityName, position.settings.Function.Evaluate, position.Location).Get(ctx, &position.Fitness)
// 	if err != nil {
// 		return err
// 	}

// }

// func EvaluateFunction(ctx workflow.Context, f func(vec []float64) float64, location Vector) float64 {
// 	fitness float64
// 	err = workflow.ExecuteActivity(ctx, EvaluateFitnessActivityName, f, location).Get(ctx, &fitness
// 	if err != nil {
// 		return fitness, err
// 	}
// }

func (position *Position) Copy() *Position {
	newPosition := NewPosition(position.settings)
	copy(newPosition.Location, position.Location)
	newPosition.Fitness = position.Fitness
	return newPosition
}

func (position *Position) IsBetterThan(other *Position) bool {
	return position.Fitness < other.Fitness
}
