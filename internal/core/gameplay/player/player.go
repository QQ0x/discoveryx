// Package player implements the player entity and its related functionality.
// It handles player movement, input processing, rendering, and state management.
// The player is the central entity that the user controls, and this package
// provides all the necessary components for a responsive and smooth player experience.
//
// The player's movement system supports both keyboard and touch input, with
// sophisticated smoothing algorithms to ensure natural-feeling motion and rotation.
// Physics effects like momentum, friction, and environmental forces are also applied
// to create realistic movement.
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

// Player represents the player entity in the game world.
// It implements the ecs.Entity interface and manages all aspects of the player:
// - Visual representation (sprite)
// - Position and movement in the game world
// - Rotation and orientation
// - Velocity and acceleration
// - Input handling (keyboard and touch)
// - Physics interactions
//
// The Player struct uses a combination of immediate and target values for
// movement properties to enable smooth transitions and natural-feeling controls.
type Player struct {
	sprite          *ebiten.Image // The player's visual representation
	world           ecs.World     // Reference to the game world for boundaries and positioning
	rotation        float64       // Current rotation angle in radians (0 = up, increases clockwise)
	position        math.Vector   // Current position in the game world (relative to center)
	playerVelocity  float64       // Current movement speed in units per frame
	curAcceleration float64       // Current acceleration rate
	health          Health        // Player health information

	// Fields for smooth movement and input handling
	targetRotation float64   // Desired rotation angle the player is turning toward
	targetVelocity float64   // Desired velocity the player is accelerating toward
	lastSwipeTime  time.Time // Timestamp of the most recent touch swipe for timing calculations
	lastSwipeAngle float64   // Direction of the most recent touch swipe for movement calculations
	isMoving       bool      // Whether the player is actively being controlled (affects friction)
}

// NewPlayer creates a new player instance with default settings.
// This factory function initializes a Player with:
// - The player sprite loaded from assets
// - A reference to the game world for boundaries and positioning
// - Default values for movement properties
//
// The world parameter provides access to world dimensions and other
// game state that the player needs to interact with the environment.
func NewPlayer(world ecs.World) *Player {
	// Load the player sprite from the assets package
	sprite := assets.PlayerSprite

	// Create and initialize the player with default values
	p := &Player{
		sprite: sprite,
		world:  world,
		health: NewHealth(constants.PlayerMaxHealth),
		// Other fields will initialize to their zero values:
		// - position: (0,0) vector
		// - rotation: 0 radians (facing up)
		// - velocity: 0 (stationary)
		// - isMoving: false (not actively controlled)
	}

	return p
}

// GetPosition returns the player's current position vector.
// This is used by other systems for:
// - Camera following
// - Collision detection
// - Projectile spawning
// - UI positioning
func (p *Player) GetPosition() math.Vector {
	return p.position
}

// GetVelocity returns the player's current velocity magnitude.
// This is used by other systems for:
// - Physics calculations
// - Visual effects that depend on speed
// - Sound effects with doppler effect
// - AI that reacts to player speed
func (p *Player) GetVelocity() float64 {
	return p.playerVelocity
}

// GetRotation returns the player's current rotation in radians.
// This is used by other systems for:
// - Determining the direction of projectiles
// - Visual effects that depend on orientation
// - Collision detection with directional hitboxes
func (p *Player) GetRotation() float64 {
	return p.rotation
}

// SetPosition sets the player's position to the specified vector.
// This is typically used for:
// - Initial positioning
// - Teleportation effects
// - Respawning after death
// - Cutscene positioning
func (p *Player) SetPosition(position math.Vector) {
	p.position = position
}

// Draw renders the player sprite to the screen with proper transformation.
// This method handles:
// - Scaling the sprite to the appropriate size
// - Rotating the sprite to match the player's orientation
// - Positioning the sprite at the player's location in the world
// - Applying camera offsets for scrolling
//
// The cameraOffsetX and cameraOffsetY parameters adjust the rendering position
// to account for camera movement in a scrolling world.
func (p *Player) Draw(screen *ebiten.Image, cameraOffsetX, cameraOffsetY float64) {
	// Get the dimensions of the sprite for centering
	bounds := p.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	// Create transformation options for rendering
	op := &ebiten.DrawImageOptions{}

	// Scale the sprite down to an appropriate size
	const scale = 1.0 / 3.0
	op.GeoM.Scale(scale, scale)

	// Apply transformations in the correct order:
	// 1. Center the sprite on its origin point
	op.GeoM.Translate(-halfW*scale, -halfH*scale)
	// 2. Rotate around the origin
	op.GeoM.Rotate(p.rotation)
	// 3. Position at the world center
	centerX := float64(p.world.GetWidth()) / 2
	centerY := float64(p.world.GetHeight()) / 2
	op.GeoM.Translate(centerX, centerY)
	// 4. Apply the player's position offset from center
	op.GeoM.Translate(p.position.X, p.position.Y)
	// 5. Apply camera offset for scrolling
	op.GeoM.Translate(cameraOffsetX, cameraOffsetY)

	// Draw the sprite with all transformations applied
	screen.DrawImage(p.sprite, op)
}

// ColliderRadius returns the radius used for collision calculations.
func (p *Player) ColliderRadius() float64 {
	return 12.0
}

// Health returns the player's current health struct.
func (p *Player) Health() Health {
	return p.health
}

// TakeDamage reduces the player's health by the given amount.
func (p *Player) TakeDamage(amount int) {
	p.health.Damage(amount)
}

// IsDead reports whether the player has no health remaining.
func (p *Player) IsDead() bool {
	return p.health.IsDead()
}

// HandleTouchInput converts touch swipes to player rotation and velocity.
// This method processes touch input from the touch handler and translates it
// into player movement and rotation. It implements a sophisticated control scheme
// that maps swipe direction to player rotation and swipe distance to velocity.
//
// The touch control system features:
// - Direction-based rotation that follows the swipe angle
// - Distance-based velocity that scales with swipe length
// - Momentum preservation for smooth movement
// - Gradual rotation for small adjustments
//
// This method is called by the Update method when touch input is detected.
func (p *Player) HandleTouchInput(touch input.TouchHandler) {
	// If the player isn't holding the touch, stop movement
	if !touch.IsHolding() {
		p.isMoving = false
		p.targetVelocity = 0
		return
	}

	// Get detailed information about the current swipe
	swipeInfo := touch.GetSwipeInfo()

	// Adjust angle to match player orientation (convert from touch coordinate system)
	// Add PI/2 to rotate from touch space to game space (in touch space, 0 is right)
	newRotation := swipeInfo.Angle + stdmath.Pi/2

	// Normalize rotation to the range [0, 2π)
	for newRotation < 0 {
		newRotation += 2 * stdmath.Pi
	}
	for newRotation >= 2*stdmath.Pi {
		newRotation -= 2 * stdmath.Pi
	}

	// Calculate the difference between the new rotation and current rotation
	rotationDiff := newRotation - p.rotation

	// Normalize the difference to the range [-π, π] to ensure we rotate
	// in the shortest direction (never more than 180 degrees)
	for rotationDiff > stdmath.Pi {
		rotationDiff -= 2 * stdmath.Pi
	}
	for rotationDiff < -stdmath.Pi {
		rotationDiff += 2 * stdmath.Pi
	}

	// Apply gradual rotation for small turns to prevent jerky movement
	// For small adjustments (less than 45 degrees), we apply partial rotation
	// to make fine-tuning more precise
	if stdmath.Abs(rotationDiff) < stdmath.Pi/4 {
		sensitivityFactor := 0.6 // Controls how responsive small adjustments are
		p.targetRotation = p.rotation + rotationDiff*sensitivityFactor
	} else {
		// For larger turns, rotate directly to the target angle
		p.targetRotation = newRotation
	}

	// Calculate velocity based on swipe distance using a tiered approach
	// This creates a non-linear response curve that gives better control
	var newVel float64
	if swipeInfo.Distance <= 10.0 {
		// Very slow movement for small swipes (precision control)
		newVel = swipeInfo.Distance / 25.0
	} else if swipeInfo.Distance <= 225.0 {
		// Medium speed for moderate swipes (normal movement)
		newVel = swipeInfo.Distance / 20.0
	} else {
		// Higher speed for large swipes (boost/sprint)
		// Base velocity from the moderate tier
		baseVel := 225.0 / 20.0
		// Additional velocity from the extra distance
		additionalDistance := swipeInfo.Distance - 225.0
		additionalVel := additionalDistance * 0.2 // Higher scaling factor for boost
		newVel = baseVel + additionalVel
	}

	// Cap velocity at the maximum allowed acceleration
	newVel = stdmath.Min(newVel, constants.MaxAcceleration)

	// Maintain momentum: if already moving fast, don't slow down too abruptly
	// This prevents jerky movement when adjusting direction
	if p.isMoving && newVel < p.playerVelocity*0.70 {
		newVel = p.playerVelocity * 0.70 // Preserve 70% of current velocity
	}

	// Set the target velocity and mark the player as moving
	p.targetVelocity = newVel
	p.isMoving = true
}

// HandleKeyboardInput processes arrow key input for player movement and rotation.
// This method translates keyboard input into player movement and rotation commands.
// It supports a traditional control scheme using arrow keys:
// - Left/Right arrows: Rotate the player
// - Up arrow: Move forward in the current direction
//
// The keyboard control system features:
// - Constant rotation speed when turning
// - Fixed maximum velocity when moving forward
// - Small velocity when rotating in place for better feedback
//
// This method is called by the Update method every frame to process keyboard input.
func (p *Player) HandleKeyboardInput(keyboard input.KeyboardHandler) {
	// Check which arrow keys are currently pressed
	leftPressed := keyboard.IsKeyPressed(input.KeyLeft)
	rightPressed := keyboard.IsKeyPressed(input.KeyRight)
	upPressed := keyboard.IsKeyPressed(input.KeyUp)

	// If no keys are pressed, stop movement
	if !leftPressed && !rightPressed && !upPressed {
		p.isMoving = false
		p.targetVelocity = 0
		return
	}

	// Mark the player as actively moving
	p.isMoving = true

	// Process rotation from left/right keys
	// Left key rotates counterclockwise (positive in radians)
	if leftPressed {
		p.targetRotation += constants.RotationPerSecond / 60.0
	}

	// Right key rotates clockwise (negative in radians)
	if rightPressed {
		p.targetRotation -= constants.RotationPerSecond / 60.0
	}

	// Keep rotation in the valid range [0, 2π)
	for p.targetRotation >= 2*stdmath.Pi {
		p.targetRotation -= 2 * stdmath.Pi
	}
	for p.targetRotation < 0 {
		p.targetRotation += 2 * stdmath.Pi
	}

	// Process acceleration from up key
	if upPressed {
		// Move at maximum speed when pressing up
		p.targetVelocity = constants.MaxAcceleration
	} else if leftPressed || rightPressed {
		// Apply a small velocity when only rotating
		// This helps provide visual feedback that the controls are working
		p.targetVelocity = 0.1
	} else {
		// No movement keys pressed
		p.targetVelocity = 0
	}
}

// Update processes player input and updates movement, rotation and physics.
// This is the main method called each frame to update the player's state.
// It handles:
// 1. Processing input from keyboard and touch
// 2. Smoothly interpolating rotation and velocity toward target values
// 3. Applying movement based on current rotation and velocity
// 4. Applying physics effects like friction and environmental forces
// 5. Ensuring values stay within valid ranges
//
// The deltaTime parameter ensures frame-rate independent movement,
// making the game behave consistently regardless of the device's performance.
func (p *Player) Update(inputManager *input.Manager, deltaTime float64) error {
	// Get input handlers from the input manager
	keyboard := inputManager.Keyboard()
	touch := inputManager.Touch()

	// Process keyboard input first (base controls)
	p.HandleKeyboardInput(keyboard)

	// If touch is active, it overrides keyboard input
	if touch != nil && touch.IsHolding() {
		p.HandleTouchInput(touch)
	}

	// ---- ROTATION HANDLING ----

	// Calculate rotation difference and normalize to shortest path
	// This ensures we always rotate the shortest way to the target angle
	rotationDiff := p.targetRotation - p.rotation
	for rotationDiff > stdmath.Pi {
		rotationDiff -= 2 * stdmath.Pi
	}
	for rotationDiff < -stdmath.Pi {
		rotationDiff += 2 * stdmath.Pi
	}

	// Small adjustment for left turns to compensate for turning bias
	// This makes left and right turns feel equally responsive
	if rotationDiff < 0 {
		rotationDiff *= 1.03 // 3% boost to left turn responsiveness
	}

	// Calculate rotation smoothing factor based on speed
	// Faster movement = slower rotation (more realistic turning)
	speedRatio := p.playerVelocity / constants.MaxAcceleration
	adjustedSpeedRatio := stdmath.Pow(speedRatio, constants.CurvePower)

	// Interpolate between min and max smoothing based on speed
	factor := constants.RotationSmoothingMax - (constants.RotationSmoothingMax-constants.RotationSmoothingMin)*adjustedSpeedRatio

	// Clamp the factor to valid range
	factor = stdmath.Max(constants.RotationSmoothingMin, stdmath.Min(constants.RotationSmoothingMax, factor))

	// Ensure small rotations are still noticeable
	// This prevents very small adjustments from being ignored
	minRotationFactor := 0.35
	if stdmath.Abs(rotationDiff) < 0.1 && rotationDiff != 0 {
		factor = stdmath.Max(factor, minRotationFactor)
	}

	// Apply rotation with smoothing
	// The factor determines how quickly we rotate toward the target
	// Multiply by deltaTime*60 to make it frame-rate independent
	p.rotation += rotationDiff * factor * deltaTime * 60.0

	// Keep rotation in the valid range [0, 2π)
	for p.rotation >= 2*stdmath.Pi {
		p.rotation -= 2 * stdmath.Pi
	}

	// ---- VELOCITY HANDLING ----

	// Calculate the difference between target and current velocity
	velocityDiff := p.targetVelocity - p.playerVelocity

	// Apply velocity smoothing with special handling for sharp turns
	if p.isMoving && stdmath.Abs(rotationDiff) > stdmath.Pi/2 {
		// Apply stronger smoothing during sharp turns (>90 degrees)
		// This simulates slowing down to turn, then speeding up again
		p.playerVelocity += velocityDiff * (constants.VelocitySmoothingFactor * 1.5) * deltaTime * 60.0
	} else {
		// Normal velocity smoothing for straight movement or gentle turns
		p.playerVelocity += velocityDiff * constants.VelocitySmoothingFactor * deltaTime * 60.0
	}

	// Clamp velocity to valid range
	if p.playerVelocity > constants.MaxAcceleration {
		p.playerVelocity = constants.MaxAcceleration
	} else if p.playerVelocity < 0 {
		p.playerVelocity = 0
	}

	// ---- POSITION UPDATING ----

	// Apply movement if velocity is above minimum threshold
	if p.playerVelocity > 0.02 {
		// Calculate movement vector based on rotation and velocity
		// sin(rotation) gives X component, cos(rotation) gives Y component
		// Note: Y is negated because in screen coordinates, Y increases downward
		dx := stdmath.Sin(p.rotation) * p.playerVelocity * deltaTime * 60.0
		dy := stdmath.Cos(p.rotation) * -p.playerVelocity * deltaTime * 60.0

		// Update position
		p.position.X += dx
		p.position.Y += dy
	} else if !p.isMoving {
		// Apply friction when not actively moving
		// This creates a natural deceleration effect
		frictionFactor := stdmath.Pow(0.95, deltaTime*60.0) // 5% reduction per frame at 60fps
		p.playerVelocity *= frictionFactor

		// Stop completely if velocity becomes negligible
		if p.playerVelocity < 0.01 {
			p.playerVelocity = 0
		}
	}

	// Apply environmental physics effects
	// This handles interactions with the game world like gravity wells
	p.position = physics.ApplyGravity(p.position, p.playerVelocity, deltaTime)

	// Prevent moving outside the world bounds
	var collided bool
	p.position, collided = physics.ClampToWorld(p.position, p.world)
	if collided {
		p.TakeDamage(constants.PlayerWallDamage)
	}

	return nil
}
