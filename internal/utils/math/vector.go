// Package math provides mathematical utilities for game development.
// It includes vector operations, geometric calculations, and other
// mathematical functions commonly used in 2D game development.
//
// This package is designed to be lightweight and focused on the specific
// mathematical operations needed by the game, rather than providing a
// comprehensive math library.
package math

import "math"

// Vector represents a 2D vector with X and Y components.
// It is used throughout the game for:
// - Representing positions in the game world
// - Storing movement directions and velocities
// - Calculating physics interactions
// - Defining offsets and dimensions
//
// The Vector struct is designed to be simple and efficient, with
// methods that provide common vector operations needed for game development.
type Vector struct {
	X float64 // X component (horizontal axis)
	Y float64 // Y component (vertical axis)
}

// Normalize returns a normalized version of the vector (unit vector).
// A normalized vector has the same direction as the original vector
// but a magnitude (length) of 1.0.
//
// This is commonly used for:
// - Representing pure direction without magnitude
// - Ensuring consistent movement speeds
// - Calculating reflection angles
// - Preparing vectors for further calculations
//
// Note: This method does not modify the original vector but returns a new one.
// If the vector has zero magnitude, this will result in NaN values.
func (v Vector) Normalize() Vector {
	magnitude := math.Sqrt(v.X*v.X + v.Y*v.Y)
	return Vector{v.X / magnitude, v.Y / magnitude}
}
