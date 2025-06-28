// Package constants provides centralized constants for the entire application
package constants

import "time"

// Input-related constants
const (
	// SwipeThreshold is the minimum distance for swipe detection in pixels
	SwipeThreshold = 30.0
	// SwipeDuration is the maximum duration for swipe detection
	SwipeDuration = 200 * time.Millisecond
)