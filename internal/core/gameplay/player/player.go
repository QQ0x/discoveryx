package player

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/constants"
	"discoveryx/internal/core/ecs"
	"discoveryx/internal/core/physics"
	"discoveryx/internal/input"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
	"time"
)

// Player movement constants are now defined in the constants package

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
		// Enhanced scaling for smaller circles (divisor reduced to 6.4 for 25% more velocity)
		newVel = swipeInfo.Distance / 6.4
	} else {
		// Enhanced scaling for larger circles: base velocity + more significant linear scaling
		baseVel := 225.0 / 6.4 // Base velocity increased by 25% by reducing divisor from 8.0 to 6.4
		additionalDistance := swipeInfo.Distance - 225.0
		// Increased factor for additional distance to improve forward speed
		additionalVel := additionalDistance * 0.25 // 25% of additional distance (increased by 25% from 20%)
		newVel = baseVel + additionalVel
	}

	// Cap at maximum acceleration
	newVel = stdmath.Min(newVel, constants.MaxAcceleration)

	// Maintain momentum when already moving (increased from 90% to 95% for better speed preservation)
	if p.isMoving && newVel < p.playerVelocity*0.95 {
		newVel = p.playerVelocity * 0.95
	}
	p.targetVelocity = newVel
	p.isMoving = true
}

// HandleKeyboardInput processes keyboard input for player movement
func (p *Player) HandleKeyboardInput(keyboard input.KeyboardHandler) {
	// Check which keys are pressed
	leftPressed := keyboard.IsKeyPressed(input.KeyLeft)
	rightPressed := keyboard.IsKeyPressed(input.KeyRight)
	upPressed := keyboard.IsKeyPressed(input.KeyUp)

	// If no keys are pressed, stop moving
	if !leftPressed && !rightPressed && !upPressed {
		p.isMoving = false
		p.targetVelocity = 0
		return
	}

	// Set the player as moving
	p.isMoving = true

	// Handle rotation
	if leftPressed {
		// For left key, rotate clockwise
		p.targetRotation += constants.RotationPerSecond / 60.0
	}

	if rightPressed {
		// For right key, rotate counter-clockwise
		p.targetRotation -= constants.RotationPerSecond / 60.0
	}

	// Normalize target rotation
	for p.targetRotation >= 2*stdmath.Pi {
		p.targetRotation -= 2 * stdmath.Pi
	}
	for p.targetRotation < 0 {
		p.targetRotation += 2 * stdmath.Pi
	}

	// Handle acceleration
	if upPressed {
		// When moving forward, set target velocity to max
		p.targetVelocity = constants.MaxAcceleration

		// If also turning, adjust rotation speed based on velocity
		// This creates the curved movement effect
		if leftPressed || rightPressed {
			// No additional adjustment needed here as the rotation is already set above
			// and the smooth movement system will create the curved effect
		}
	} else if leftPressed || rightPressed {
		// When only rotating (without forward movement), set a small velocity
		// This allows the player to rotate in place
		p.targetVelocity = 0.1 // Small non-zero value for rotation in place
	} else {
		// No movement keys pressed
		p.targetVelocity = 0
	}
}

// Update updates the player state
func (p *Player) Update(inputManager *input.Manager, deltaTime float64) error {
	keyboard := inputManager.Keyboard()
	touch := inputManager.Touch()

	// Process keyboard input using the smooth movement system
	p.HandleKeyboardInput(keyboard)

	// Only process touch input if it's active and available
	if touch != nil && touch.IsHolding() {
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
	speedRatio := p.playerVelocity / constants.MaxAcceleration
	factor := constants.RotationSmoothingMax - (constants.RotationSmoothingMax-constants.RotationSmoothingMin)*speedRatio
	factor = stdmath.Max(constants.RotationSmoothingMin, stdmath.Min(constants.RotationSmoothingMax, factor))
	// Apply delta time to rotation smoothing
	p.rotation += rotationDiff * factor * deltaTime * 60.0

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
		// Apply delta time to velocity smoothing
		p.playerVelocity += velocityDiff * (constants.VelocitySmoothingFactor * 1.5) * deltaTime * 60.0
	} else {
		// Apply delta time to velocity smoothing
		p.playerVelocity += velocityDiff * constants.VelocitySmoothingFactor * deltaTime * 60.0
	}

	if p.playerVelocity > constants.MaxAcceleration {
		p.playerVelocity = constants.MaxAcceleration
	} else if p.playerVelocity < 0 {
		p.playerVelocity = 0
	}

	if p.playerVelocity > 0.05 {
		// Apply delta time to movement
		dx := stdmath.Sin(p.rotation) * p.playerVelocity * deltaTime * 60.0
		dy := stdmath.Cos(p.rotation) * -p.playerVelocity * deltaTime * 60.0
		p.position.X += dx
		p.position.Y += dy
	} else if !p.isMoving {
		// Apply delta time to friction
		frictionFactor := stdmath.Pow(0.95, deltaTime*60.0)
		p.playerVelocity *= frictionFactor
		if p.playerVelocity < 0.01 {
			p.playerVelocity = 0
		}
	}

	// Apply gravity to the player's position, but only when velocity is low
	p.position = physics.ApplyGravity(p.position, p.playerVelocity, deltaTime)

	return nil
}
