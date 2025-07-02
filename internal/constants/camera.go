// Package constants provides centralized constants for the entire application
package constants

// Camera constants
const (
	// CameraDeadZoneX is the horizontal deadzone width as a percentage of screen width
	CameraDeadZoneX = 0.1
	// CameraDeadZoneY is the vertical deadzone height as a percentage of screen height
	CameraDeadZoneY = 0.1
	// CameraFollowStrengthFactor controls how quickly the camera follows the player
	// when outside the deadzone (higher values = faster camera)
	CameraFollowStrengthFactor = 1.0
	// CameraInterpolationFactor controls how smoothly the camera moves
	// (lower values = smoother movement)
	CameraInterpolationFactor = 0.05
	// CameraMaxFollowStrength is the maximum follow strength when player is far from center
	CameraMaxFollowStrength = 0.5
	// CameraVelocityThreshold is the velocity threshold below which the camera starts centering on the player
	CameraVelocityThreshold = 0.5
	// CameraCenteringStrength controls how quickly the camera centers on the player when they're stopped
	// (higher values = faster centering)
	CameraCenteringStrength = 0.05
)
