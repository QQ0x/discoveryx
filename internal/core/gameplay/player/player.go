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
	rotationPerSecond = -4.5 // Rotation speed in radians per second
	maxAcceleration   = 40.0 // Maximum acceleration value (significantly increased for much faster movement)

	// Constants for smooth movement
	rotationSmoothingMin    = 0.06                   // smoothing factor at full speed (closer to original for better performance)
	rotationSmoothingMax    = 0.4                    // smoothing factor when standing still
	velocitySmoothingFactor = 0.25                   // faster velocity changes (significantly increased for much more responsive acceleration)
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
	const scale = 1.0 / 3.0
	op.GeoM.Scale(scale, scale)

	// The following sequence ensures rotation around the center of the sprite:
	// 1. Move to origin (center the sprite at 0,0)
	op.GeoM.Translate(-halfW*scale, -halfH*scale)

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
	speed := stdmath.Abs(rotationPerSecond) / float64(ebiten.TPS())

	if keyboard.IsKeyPressed(input.KeyLeft) {
		// For left key, we still want to rotate clockwise, so we add to rotation
		p.rotation += speed
	}

	if keyboard.IsKeyPressed(input.KeyRight) {
		// For right key, continue to rotate clockwise
		p.rotation += speed * 2 // Faster rotation for right key to maintain responsiveness
	}

	// Normalize rotation to keep it within 0 to 2π
	for p.rotation >= 2*stdmath.Pi {
		p.rotation -= 2 * stdmath.Pi
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
	if !touch.IsHolding() {
		p.isMoving = false
		p.targetVelocity = 0
		return
	}

	swipeInfo := touch.GetSwipeInfo()

	// Convert swipe angle so that 0 rad points up like the player sprite
	newRotation := swipeInfo.Angle + stdmath.Pi/2
	for newRotation < 0 {
		newRotation += 2 * stdmath.Pi
	}
	for newRotation >= 2*stdmath.Pi {
		newRotation -= 2 * stdmath.Pi
	}

	p.targetRotation = newRotation

	// Map swipe distance to target velocity with a simplified dynamic scaling
	// that maintains the improved behavior for larger circles but with better performance
	var newVel float64
	if swipeInfo.Distance <= 225.0 {
		// Enhanced scaling for smaller circles (divisor further reduced from 10.0 to 8.0 for even faster movement)
		newVel = swipeInfo.Distance / 8.0
	} else {
		// Enhanced scaling for larger circles: base velocity + more significant linear scaling
		baseVel := 225.0 / 8.0 // Base velocity significantly increased by reducing divisor
		additionalDistance := swipeInfo.Distance - 225.0
		// Increased factor for additional distance to improve forward speed
		additionalVel := additionalDistance * 0.20 // 20% of additional distance (up from 15%)
		newVel = baseVel + additionalVel
	}

	// Cap at maximum acceleration
	newVel = stdmath.Min(newVel, maxAcceleration)

	// Maintain momentum when already moving (increased from 90% to 95% for better speed preservation)
	if p.isMoving && newVel < p.playerVelocity*0.95 {
		newVel = p.playerVelocity * 0.95
	}
	p.targetVelocity = newVel
	p.isMoving = true
}

// Update updates the player state
func (p *Player) Update() error {
	keyboard := input.GetKeyboard()
	touch := input.GetTouch()

	// keyboard input for desktop testing
	p.HandleRotation(keyboard)
	p.HandleAcceleration(keyboard)

	if touch != nil {
		p.HandleTouchInput(touch)
	}

	// Smooth rotation towards the target rotation with dynamic factor based
	// on current velocity. Higher speeds rotate slower to create a curve.
	rotationDiff := p.targetRotation - p.rotation

	// Normalize the difference to be between -π and π for shortest path rotation
	for rotationDiff > stdmath.Pi {
		rotationDiff -= 2 * stdmath.Pi
	}
	for rotationDiff < -stdmath.Pi {
		rotationDiff += 2 * stdmath.Pi
	}

	// Simple adjustment for left turns to address perceptual bias
	// Apply a very small fixed adjustment to all left turns for better performance
	if rotationDiff < 0 {
		rotationDiff *= 1.03 // Minimal 3% adjustment for left turns
	}

	// Always use the shortest path for rotation
	speedRatio := p.playerVelocity / maxAcceleration
	factor := rotationSmoothingMax - (rotationSmoothingMax-rotationSmoothingMin)*speedRatio
	factor = stdmath.Max(rotationSmoothingMin, stdmath.Min(rotationSmoothingMax, factor))
	p.rotation += rotationDiff * factor

	// Normalize rotation to keep it within 0 to 2π
	for p.rotation >= 2*stdmath.Pi {
		p.rotation -= 2 * stdmath.Pi
	}

	// Smooth velocity towards the target velocity
	velocityDiff := p.targetVelocity - p.playerVelocity

	// Apply stronger smoothing when changing direction drastically
	// This helps the player respond more quickly to sharp direction changes
	if p.isMoving && stdmath.Abs(rotationDiff) > stdmath.Pi/2 {
		// When turning more than 90 degrees, apply stronger smoothing
		p.playerVelocity += velocityDiff * (velocitySmoothingFactor * 1.5)
	} else {
		p.playerVelocity += velocityDiff * velocitySmoothingFactor
	}

	if p.playerVelocity > maxAcceleration {
		p.playerVelocity = maxAcceleration
	} else if p.playerVelocity < 0 {
		p.playerVelocity = 0
	}

	if p.playerVelocity > 0.05 {
		dx := stdmath.Sin(p.rotation) * p.playerVelocity
		dy := stdmath.Cos(p.rotation) * -p.playerVelocity
		p.position.X += dx
		p.position.Y += dy
	} else if !p.isMoving {
		// apply reduced friction when no input (reduced from 0.9 to 0.95 for longer momentum preservation)
		p.playerVelocity *= 0.95
		if p.playerVelocity < 0.01 {
			p.playerVelocity = 0
		}
	}

	return nil
}
