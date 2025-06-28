// Package constants provides centralized constants for the entire application
package constants

import "time"

// Player movement constants
const (
	// RotationPerSecond is the rotation speed in radians per second
	RotationPerSecond = -4.5
	// MaxAcceleration is the maximum acceleration value
	MaxAcceleration = 50.0

	// RotationSmoothingMin is the smoothing factor at full speed
	RotationSmoothingMin = 0.06
	// RotationSmoothingMax is the smoothing factor when standing still
	RotationSmoothingMax = 0.4
	// VelocitySmoothingFactor for faster velocity changes
	VelocitySmoothingFactor = 0.25
	// MinSwipeDuration is the minimum duration for a swipe to be considered valid
	MinSwipeDuration = 200 * time.Millisecond
)