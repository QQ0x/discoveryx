// Package physics implements game physics simulations and interactions.
// It provides functionality for gravity, collision detection, and other
// physical forces that affect entities in the game world.
//
// The physics system is designed to be:
// - Lightweight and efficient for mobile devices
// - Frame-rate independent using deltaTime
// - Configurable through constants
// - Selective in application (e.g., gravity only affects slow-moving objects)
//
// This package works closely with the gameplay systems to create
// realistic movement and interactions between game objects.
package physics

import (
	"discoveryx/internal/constants"
	"discoveryx/internal/utils/math"
)

// Physics constants are now defined in the constants package

// ApplyGravity applies a gravity force to the given position vector
// but only if the current velocity is below the LowVelocityThreshold.
//
// This function implements a simplified gravity model where:
// - Fast-moving objects (above LowVelocityThreshold) ignore gravity
//   to simulate momentum overcoming gravitational pull
// - Slow or stationary objects are gradually pulled downward
// - The effect scales with deltaTime for frame-rate independence
//
// This selective gravity application creates interesting gameplay dynamics:
// - Players must maintain sufficient speed to overcome gravity wells
// - Stationary objects will slowly drift downward
// - Projectiles with high velocity will maintain straight trajectories
//
// Parameters:
//   - position: The position vector to apply gravity to
//   - velocity: The current velocity magnitude (used to determine if gravity applies)
//   - deltaTime: The time elapsed since the last update in seconds
//
// Returns:
//   - The updated position vector after applying gravity
func ApplyGravity(position math.Vector, velocity float64, deltaTime float64) math.Vector {
	// Only apply gravity if velocity is below the threshold
	if velocity < constants.LowVelocityThreshold {
		// Apply gravity force in the downward direction (positive Y)
		// Scale by 60.0 to maintain original speed at 60 FPS
		position.Y += constants.GravityForce * deltaTime * 60.0
	}

	return position
}
