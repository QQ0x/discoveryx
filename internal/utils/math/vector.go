package math

import "math"

// Vector represents a 2D vector with X and Y components
type Vector struct {
	X float64
	Y float64
}

// Normalize returns a normalized version of the vector (unit vector)
func (v Vector) Normalize() Vector {
	magnitude := math.Sqrt(v.X*v.X + v.Y*v.Y)
	return Vector{v.X / magnitude, v.Y / magnitude}
}