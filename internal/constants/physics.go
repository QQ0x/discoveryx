// Package constants provides centralized constants for the entire application
package constants

// Physics-related constants
const (
	// GravityForce is the strength of the gravity force
	GravityForce = 0.4

	// LowVelocityThreshold is the threshold below which gravity is applied
	// When the player's velocity is above this threshold, no gravity is applied
	LowVelocityThreshold = 10.0
)
