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
	xLo := settings.Function.xLo
	xHi := settings.Function.xHi
	for i := 0; i < len(pos.Location); i++ {
		pos.Location[i] = xLo + (xHi-xLo)*settings.rng.Float64()
	}
	// pos.Fitness = EvaluateFunction(settings.Function.Evaluate, pos.Location)
	return pos
}

func (position *Position) Copy() *Position {
	newPosition := NewPosition(position.settings)
	copy(newPosition.Location, position.Location)
	newPosition.Fitness = position.Fitness
	return newPosition
}

func (position *Position) IsBetterThan(other *Position) bool {
	return position.Fitness < other.Fitness
}
