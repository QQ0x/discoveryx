// Package constants provides centralized constants for the entire application
package constants

import "time"

// Player movement constants
const (
	// RotationPerSecond is the rotation speed in radians per second
	RotationPerSecond = -4.5
	// MaxAcceleration is the maximum acceleration value
	MaxAcceleration = 3.5

	// RotationSmoothingMin is the smoothing factor at full speed
	RotationSmoothingMin = 0.25
	// RotationSmoothingMax is the smoothing factor when standing still
	RotationSmoothingMax = 0.45
	// VelocitySmoothingFactor for velocity changes (reduced for more gradual acceleration)
	VelocitySmoothingFactor = 0.08
	// MinSwipeDuration is the minimum duration for a swipe to be considered valid
	MinSwipeDuration = 200 * time.Millisecond
	// CurvePower controls how much the turning radius is affected by speed (higher = tighter turns at low speeds)
	CurvePower = 1.7
)
