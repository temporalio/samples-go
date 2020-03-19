package pso

import "math/rand"

type Vector []float64

type Position struct {
	Location Vector
	Fitness  float64
}

func NewPosition(dim int) *Position {
	loc := make([]float64, dim)
	return &Position{
		Location: loc,
		// Fitness:  EvaluateFunction(settings.Function.Evaluate, loc),
	}
}

func RandomPosition(function ObjectiveFunction, rng *rand.Rand) *Position {
	pos := NewPosition(function.dim)
	xLo := function.xLo
	xHi := function.xHi
	for i := 0; i < len(pos.Location); i++ {
		pos.Location[i] = xLo + (xHi-xLo)*rng.Float64()
	}
	// pos.Fitness = EvaluateFunction(settings.Function.Evaluate, pos.Location)
	return pos
}

func (position *Position) Copy() *Position {
	newPosition := NewPosition(len(position.Location))
	copy(newPosition.Location, position.Location)
	newPosition.Fitness = position.Fitness
	return newPosition
}

func (position *Position) IsBetterThan(other *Position) bool {
	return position.Fitness < other.Fitness
}
