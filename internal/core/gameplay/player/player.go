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

// GetPosition returns the player's position
func (p *Player) GetPosition() math.Vector {
	return p.position
}

// GetVelocity returns the player's current velocity
func (p *Player) GetVelocity() float64 {
	return p.playerVelocity
}

// GetRotation returns the player's rotation in radians
func (p *Player) GetRotation() float64 {
	return p.rotation
}

// SetPosition sets the player's position
func (p *Player) SetPosition(position math.Vector) {
	p.position = position
}

// Draw draws the player
func (p *Player) Draw(screen *ebiten.Image, cameraOffsetX, cameraOffsetY float64) {
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

	// 5. Apply camera offset
	op.GeoM.Translate(cameraOffsetX, cameraOffsetY)

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

	// Calculate the difference between current rotation and new rotation
	rotationDiff := newRotation - p.rotation

	// Normalize the difference to be between -π and π for shortest path rotation
	for rotationDiff > stdmath.Pi {
		rotationDiff -= 2 * stdmath.Pi
	}
	for rotationDiff < -stdmath.Pi {
		rotationDiff += 2 * stdmath.Pi
	}

	// For small angle changes, make the adjustment more gradual
	// This allows for smaller curves during direction changes
	if stdmath.Abs(rotationDiff) < stdmath.Pi/4 { // Less than 45 degrees
		// Apply a sensitivity factor for small turns (smaller = more gradual)
		sensitivityFactor := 0.6 // Adjust this value to control small curve sensitivity

		// Calculate a more gradual target rotation
		p.targetRotation = p.rotation + rotationDiff*sensitivityFactor
	} else {
		// For larger turns, use the original behavior
		p.targetRotation = newRotation
	}

	// Map swipe distance to target velocity with a more gradual dynamic scaling
	// to prevent too fast acceleration with small swipes
	var newVel float64
	if swipeInfo.Distance <= 10.0 {
		// Special case for very small swipes to allow extremely slow movement
		newVel = swipeInfo.Distance / 25.0
	} else if swipeInfo.Distance <= 225.0 {
		// Enhanced scaling for smaller circles (divisor increased to 20.0 for much lower minimum speed)
		newVel = swipeInfo.Distance / 20.0
	} else {
		// Enhanced scaling for larger circles: base velocity + more significant linear scaling
		baseVel := 225.0 / 20.0 // Base velocity calculated with the new divisor
		additionalDistance := swipeInfo.Distance - 225.0
		// Increased factor for additional distance to improve forward speed
		additionalVel := additionalDistance * 0.2 // 20% of additional distance (reduced from 30%)
		newVel = baseVel + additionalVel
	}

	// Cap at maximum acceleration
	newVel = stdmath.Min(newVel, constants.MaxAcceleration)

	// Maintain momentum when already moving (reduced from 85% to 70% to allow for quicker deceleration with small inputs)
	if p.isMoving && newVel < p.playerVelocity*0.70 {
		newVel = p.playerVelocity * 0.70
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
	// Use a non-linear formula to make turns tighter at lower speeds
	// Raise the speedRatio to the power of CurvePower to create an adjustable curve
	// Higher CurvePower values make turns tighter at lower speeds
	adjustedSpeedRatio := stdmath.Pow(speedRatio, constants.CurvePower)
	factor := constants.RotationSmoothingMax - (constants.RotationSmoothingMax-constants.RotationSmoothingMin)*adjustedSpeedRatio
	factor = stdmath.Max(constants.RotationSmoothingMin, stdmath.Min(constants.RotationSmoothingMax, factor))

	// Ensure small rotation differences still have a noticeable effect
	// by applying a minimum rotation amount for very small inputs
	minRotationFactor := 0.35 // Increased minimum factor for very small rotation differences
	// Expanded threshold for small rotation differences to make more subtle curves possible
	if stdmath.Abs(rotationDiff) < 0.1 && rotationDiff != 0 {
		// For very small rotation differences, use a higher factor to make them more noticeable
		factor = stdmath.Max(factor, minRotationFactor)
	}

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

	if p.playerVelocity > 0.02 {
		// Apply delta time to movement (threshold reduced from 0.05 to 0.02 to allow slower movement)
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
