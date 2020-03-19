package pso

import (
	"math"
)

func CalculateSwarmSize(dim, max_size int) int {
	s := 10. + 2.*math.Sqrt(float64(dim))
	size := int(math.Floor(s + 0.5))
	if size > max_size {
		return max_size
	} else {
		return size
	}
}
