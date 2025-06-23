package player

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/ecs"
	"discoveryx/internal/input"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
	"time"
)

// Constants for player movement
const (
	rotationPerSecond = -3.0 // Rotation speed in radians per second
	maxAcceleration   = 10.0 // Maximum acceleration value

	// Constants for smooth movement
	rotationSmoothingFactor = 0.5                    // Higher values make rotation more responsive
	velocitySmoothingFactor = 0.25                   // Higher values make velocity changes more responsive
	minSwipeDuration        = 200 * time.Millisecond // Minimum duration for a swipe to be considered valid
)

// Player represents the player entity
type Player struct {
	sprite          *ebiten.Image
	world           ecs.World
	rotation        float64
	position        math.Vector
	playerVelocity  float64 // Player speed
	curAcceleration float64 // Current acceleration

	// Fields for smooth movement
	targetRotation float64   // Target rotation angle
	targetVelocity float64   // Target velocity
	lastSwipeTime  time.Time // Time of the last swipe
	lastSwipeAngle float64   // Angle of the last swipe
	isMoving       bool      // Whether the player is currently moving
}

// NewPlayer creates a new player instance
func NewPlayer(world ecs.World) *Player {
	sprite := assets.PlayerSprite

	p := &Player{
		sprite: sprite,
		world:  world,
	}

	return p
}

// Draw draws the player
func (p *Player) Draw(screen *ebiten.Image) {
	// Calculate sprite center
	bounds := p.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	op := &ebiten.DrawImageOptions{}

	// Scale to 1/3 of original size
	op.GeoM.Scale(1.0/3.0, 1.0/3.0)

	// The following sequence ensures rotation around the center of the sprite:
	// 1. Move to origin (center the sprite at 0,0)
	op.GeoM.Translate(-halfW, -halfH)

	// 2. Rotate around center point (which is now at 0,0)
	op.GeoM.Rotate(p.rotation)

	// 3. Move to world center
	centerX := float64(p.world.GetWidth()) / 2
	centerY := float64(p.world.GetHeight()) / 2
	op.GeoM.Translate(centerX, centerY)

	// 4. Apply player position offset
	op.GeoM.Translate(p.position.X, p.position.Y)

	// Draw sprite
	screen.DrawImage(p.sprite, op)
}

// HandleRotation updates the player's rotation based on input
func (p *Player) HandleRotation(keyboard input.KeyboardHandler) {
	speed := rotationPerSecond / float64(ebiten.TPS())

	if keyboard.IsKeyPressed(input.KeyLeft) {
		p.rotation -= speed
	}

	if keyboard.IsKeyPressed(input.KeyRight) {
		p.rotation += speed
	}
}

// HandleAcceleration handles player acceleration and updates position
func (p *Player) HandleAcceleration(keyboard input.KeyboardHandler) {
	if keyboard.IsKeyPressed(input.KeyUp) {
		if p.curAcceleration < maxAcceleration {
			p.curAcceleration = p.playerVelocity + 4
		}

		if p.curAcceleration >= 8 {
			p.curAcceleration = 8
		}

		p.playerVelocity = p.curAcceleration

		// Move in the direction we are pointing
		dx := stdmath.Sin(p.rotation) * p.curAcceleration
		dy := stdmath.Cos(p.rotation) * -p.curAcceleration

		// Move the player on screen
		p.position.X += dx
		p.position.Y += dy
	}
}

// HandleTouchInput processes touch input for player movement
func (p *Player) HandleTouchInput(touch input.TouchHandler) {
	// Check if there's an active swipe or hold
	if touch.IsHolding() {
		// Get the swipe information
		swipeInfo := touch.GetSwipeInfo()

		// Convert the swipe angle to a rotation angle for the player
		// Adjust by -π/2 because the player sprite points up at rotation 0,
		// but Atan2 returns 0 for the positive x-axis
		// Add π to reverse the direction (fix for reversed touch controls)
		newTargetRotation := swipeInfo.Angle - stdmath.Pi/2 + stdmath.Pi

		// Normalize the rotation to be between 0 and 2π for consistent handling
		for newTargetRotation < 0 {
			newTargetRotation += 2 * stdmath.Pi
		}
		for newTargetRotation >= 2*stdmath.Pi {
			newTargetRotation -= 2 * stdmath.Pi
		}

		// Record the time of this swipe
		now := time.Now()

		// Calculate the angle difference between the new target rotation and the last swipe angle
		angleDiff := stdmath.Abs(newTargetRotation - p.lastSwipeAngle)

		// Normalize the angle difference to be between 0 and π
		if angleDiff > stdmath.Pi {
			angleDiff = 2*stdmath.Pi - angleDiff
		}

		// We no longer need to calculate time since last swipe as we always update rotation

		// Always update the target rotation, but with different weights based on conditions
		// This ensures the player always responds to direction changes, even small ones
		var weight float64

		if !p.isMoving {
			// If player is not moving yet, use the new target directly
			weight = 1.0
		} else if angleDiff > 0.5 {
			// For large angle changes, use a moderate weight to smooth transition
			// but ensure it's responsive enough
			weight = stdmath.Max(0.5, 1.0-angleDiff/stdmath.Pi)
		} else if angleDiff > 0.1 {
			// For medium angle changes, use a higher weight for more responsiveness
			weight = stdmath.Max(0.7, 1.0-angleDiff/stdmath.Pi)
		} else {
			// For small angle changes, use an even higher weight for immediate response
			weight = 0.9
		}

		// Apply the weighted average with the calculated weight
		p.targetRotation = p.targetRotation*(1-weight) + newTargetRotation*weight

		// Update the last swipe time and angle
		p.lastSwipeTime = now
		p.lastSwipeAngle = newTargetRotation

		// Calculate acceleration based on swipe distance
		// Map the swipe distance to an acceleration value between 0 and maxAcceleration
		// The longer the swipe, the higher the acceleration
		normalizedDistance := stdmath.Min(swipeInfo.Distance/150.0, 1.0) // Normalize to 0-1 range with shorter distance requirement

		// For very quick swipes, maintain a minimum velocity to ensure responsiveness
		if swipeInfo.Speed > 800 && normalizedDistance < 0.3 {
			normalizedDistance = 0.3 // Ensure a minimum velocity for quick swipes
		}

		newTargetVelocity := normalizedDistance * maxAcceleration

		// Always update target velocity, but with different weights based on the velocity difference
		velocityDiff := stdmath.Abs(newTargetVelocity - p.targetVelocity)

		if velocityDiff > maxAcceleration*0.7 {
			// For large velocity changes, use a moderate weight
			weight = 0.5
		} else if velocityDiff > maxAcceleration*0.3 {
			// For medium velocity changes, use a higher weight
			weight = 0.7
		} else {
			// For small velocity changes, use an even higher weight
			weight = 0.9
		}

		// Apply the weighted average with the calculated weight
		p.targetVelocity = p.targetVelocity*(1-weight) + newTargetVelocity*weight

		// Mark that the player is moving
		p.isMoving = true
	} else {
		// No active swipe or hold, gradually slow down the player
		p.targetVelocity = 0

		// If the player has stopped moving, reset the moving flag
		if p.playerVelocity < 0.1 {
			p.isMoving = false
		}
	}
}

// Update updates the player state
func (p *Player) Update() error {
	// Get input handlers from input manager
	keyboard := input.GetKeyboard()
	touch := input.GetTouch()

	// Handle keyboard input
	p.HandleRotation(keyboard)
	p.HandleAcceleration(keyboard)

	// Handle touch input if available
	if touch != nil {
		p.HandleTouchInput(touch)
	}

	// Apply smooth rotation - always update rotation regardless of movement state
	// Smoothly interpolate current rotation towards target rotation

	// Normalize both rotations to be between 0 and 2π for consistent comparison
	normalizedTarget := p.targetRotation
	for normalizedTarget < 0 {
		normalizedTarget += 2 * stdmath.Pi
	}
	for normalizedTarget >= 2*stdmath.Pi {
		normalizedTarget -= 2 * stdmath.Pi
	}

	normalizedCurrent := p.rotation
	for normalizedCurrent < 0 {
		normalizedCurrent += 2 * stdmath.Pi
	}
	for normalizedCurrent >= 2*stdmath.Pi {
		normalizedCurrent -= 2 * stdmath.Pi
	}

	// Calculate the rotation difference
	rotationDiff := normalizedTarget - normalizedCurrent

	// Normalize the rotation difference to be between -π and π for shortest path rotation
	if rotationDiff > stdmath.Pi {
		rotationDiff -= 2 * stdmath.Pi
	}
	if rotationDiff < -stdmath.Pi {
		rotationDiff += 2 * stdmath.Pi
	}

	// Ensure rotationDiff is properly normalized for consistent handling of left and right turns

	// Calculate rotation step based on velocity and rotation difference
	// When player is moving faster, rotation should be slower
	// When player is standing still, rotation should be immediate
	var rotationFactor float64
	if p.playerVelocity > 5.0 {
		// At high speeds, rotate slower
		rotationFactor = rotationSmoothingFactor * 0.5
	} else if p.playerVelocity > 2.0 {
		// At medium speeds, rotate moderately
		rotationFactor = rotationSmoothingFactor * 0.7
	} else {
		// At low speeds or standing still, rotate quickly
		rotationFactor = rotationSmoothingFactor
	}

	// Apply rotation with the calculated factor
	rotationStep := rotationDiff * rotationFactor

	// Ensure a minimum step size for small differences to prevent getting stuck
	// but make it larger than before to ensure more responsive rotation
	if rotationDiff > 0.01 && rotationStep < 0.01 {
		rotationStep = 0.01
	} else if rotationDiff < -0.01 && rotationStep > -0.01 {
		rotationStep = -0.01
	}

	p.rotation += rotationStep

	// Apply velocity changes and position updates only if moving
	if p.isMoving {

		// Smoothly interpolate current velocity towards target velocity
		velocityDiff := p.targetVelocity - p.playerVelocity

		// Apply smoothing with a more responsive factor
		velocityStep := velocityDiff * velocitySmoothingFactor

		// Ensure a minimum step size for small differences to prevent getting stuck
		// but make it larger than before to ensure more responsive velocity changes
		if velocityDiff > 0.1 && velocityStep < 0.1 {
			velocityStep = 0.1
		} else if velocityDiff < -0.1 && velocityStep > -0.1 {
			velocityStep = -0.1
		}

		// Apply velocity change
		p.playerVelocity += velocityStep

		// Ensure velocity doesn't exceed maximum acceleration
		if p.playerVelocity > maxAcceleration {
			p.playerVelocity = maxAcceleration
		} else if p.playerVelocity < 0 {
			p.playerVelocity = 0
		}

		p.curAcceleration = p.playerVelocity

		// Only update position if player has significant velocity
		// This prevents the player from jumping when rotating while standing
		if p.playerVelocity > 0.2 {
			// Normalize rotation for consistent handling
			normalizedRotation := p.rotation
			for normalizedRotation < 0 {
				normalizedRotation += 2 * stdmath.Pi
			}
			for normalizedRotation >= 2*stdmath.Pi {
				normalizedRotation -= 2 * stdmath.Pi
			}

			// Calculate the angle halfway between the previous rotation and current rotation
			// This creates a smoother curved path during turns
			// Adjust the turning radius based on the player's velocity
			// Higher velocity = wider turns (larger turning radius)
			turnRadiusFactor := stdmath.Min(1.0, p.playerVelocity/maxAcceleration)

			// Calculate a weighted rotation that creates a curved path
			// The weight depends on the player's velocity - higher velocity means more weight on the previous rotation
			// This creates wider turns at higher speeds
			turnWeight := 0.2 + 0.5*turnRadiusFactor // Range from 0.2 (tight turns) to 0.7 (wide turns)

			// Calculate the effective rotation for movement, weighted between previous and current rotation
			// Use the sign of rotationStep to determine whether to add or subtract the turning effect
			var effectiveRotation float64
			if rotationStep > 0 {
				// Right turn - subtract the turning effect
				effectiveRotation = normalizedRotation - stdmath.Abs(rotationStep)*turnWeight
			} else if rotationStep < 0 {
				// Left turn - add the turning effect
				effectiveRotation = normalizedRotation + stdmath.Abs(rotationStep)*turnWeight
			} else {
				// No rotation - move straight
				effectiveRotation = normalizedRotation
			}

			// Normalize the effective rotation
			for effectiveRotation < 0 {
				effectiveRotation += 2 * stdmath.Pi
			}
			for effectiveRotation >= 2*stdmath.Pi {
				effectiveRotation -= 2 * stdmath.Pi
			}

			// Calculate movement using the effective rotation for a curved path
			dx := stdmath.Sin(effectiveRotation) * p.curAcceleration
			dy := stdmath.Cos(effectiveRotation) * -p.curAcceleration

			// Move the player on screen
			p.position.X += dx
			p.position.Y += dy
		}
	} else if p.playerVelocity > 0 {
		// Gradually slow down if not moving but still has velocity
		// Use a smoother deceleration curve that starts slow and then speeds up
		// This prevents abrupt stops at high speeds
		if p.playerVelocity > 5.0 {
			// For high speeds, decelerate more gradually
			p.playerVelocity *= 0.95
		} else if p.playerVelocity > 2.0 {
			// For medium speeds, decelerate moderately
			p.playerVelocity *= 0.9
		} else {
			// For low speeds, decelerate more quickly
			p.playerVelocity *= 0.8
		}

		// If velocity is very small, just stop completely to avoid tiny movements
		if p.playerVelocity < 0.2 {
			p.playerVelocity = 0
			// Return early to prevent position updates when velocity is too small
			// This prevents the player from jumping when rotating while standing
			return nil
		}
		p.curAcceleration = p.playerVelocity

		// Normalize rotation for consistent handling
		normalizedRotation := p.rotation
		for normalizedRotation < 0 {
			normalizedRotation += 2 * stdmath.Pi
		}
		for normalizedRotation >= 2*stdmath.Pi {
			normalizedRotation -= 2 * stdmath.Pi
		}

		// Calculate the angle halfway between the previous rotation and current rotation
		// This creates a smoother curved path during turns
		// Adjust the turning radius based on the player's velocity
		// Higher velocity = wider turns (larger turning radius)
		turnRadiusFactor := stdmath.Min(1.0, p.playerVelocity/maxAcceleration)

		// Calculate a weighted rotation that creates a curved path
		// The weight depends on the player's velocity - higher velocity means more weight on the previous rotation
		// This creates wider turns at higher speeds
		turnWeight := 0.2 + 0.5*turnRadiusFactor // Range from 0.2 (tight turns) to 0.7 (wide turns)

		// Calculate the effective rotation for movement, weighted between previous and current rotation
		// Use the sign of rotationStep to determine whether to add or subtract the turning effect
		var effectiveRotation float64
		if rotationStep > 0 {
			// Right turn - subtract the turning effect
			effectiveRotation = normalizedRotation - stdmath.Abs(rotationStep)*turnWeight
		} else if rotationStep < 0 {
			// Left turn - add the turning effect
			effectiveRotation = normalizedRotation + stdmath.Abs(rotationStep)*turnWeight
		} else {
			// No rotation - move straight
			effectiveRotation = normalizedRotation
		}

		// Normalize the effective rotation
		for effectiveRotation < 0 {
			effectiveRotation += 2 * stdmath.Pi
		}
		for effectiveRotation >= 2*stdmath.Pi {
			effectiveRotation -= 2 * stdmath.Pi
		}

		// Calculate movement using the effective rotation for a curved path
		dx := stdmath.Sin(effectiveRotation) * p.curAcceleration
		dy := stdmath.Cos(effectiveRotation) * -p.curAcceleration

		// Move the player on screen
		p.position.X += dx
		p.position.Y += dy
	}

	return nil
}
