package physics

import (
	"discoveryx/internal/constants"
	"discoveryx/internal/utils/math"
)

// Physics constants are now defined in the constants package

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
	if velocity < constants.LowVelocityThreshold {
		// Apply gravity force in the downward direction (positive Y)
		position.Y += constants.GravityForce
	}

	return position
}
