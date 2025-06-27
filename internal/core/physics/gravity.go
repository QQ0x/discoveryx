package physics

import (
	"discoveryx/internal/utils/math"
)

// Constants for gravity
const (
	// GravityForce is the strength of the gravity force
	GravityForce = 0.4

	// LowVelocityThreshold is the threshold below which gravity is applied
	// When the player's velocity is above this threshold, no gravity is applied
	LowVelocityThreshold = 10.0
)

// ApplyGravity applies a gravity force to the given position vector
// but only if the current velocity is below the LowVelocityThreshold
//
// Parameters:
//   - position: The position vector to apply gravity to
//   - velocity: The current velocity magnitude
//
// Returns:
//   - The updated position vector after applying gravity
func ApplyGravity(position math.Vector, velocity float64) math.Vector {
	// Only apply gravity if velocity is below the threshold
	if velocity < LowVelocityThreshold {
		// Apply gravity force in the downward direction (positive Y)
		position.Y += GravityForce
	}

	return position
}
