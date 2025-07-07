// Package enemies implements enemy entities and their spawning mechanics.
// It provides functionality for creating, updating, and rendering different types
// of enemies in the game world, as well as sophisticated spawning algorithms
// that place enemies at appropriate locations based on the world geometry.
//
// The enemies package works closely with the world generation system to ensure
// enemies are placed in challenging but fair positions, typically along walls
// and other environmental features.
package enemies

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/physics"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	stdmath "math"
)

// Enemy represents an enemy entity in the game world.
// Enemies are obstacles or adversaries that the player must avoid or defeat.
// Each enemy has a specific type, position, rotation, and visual representation.
//
// Enemies are typically spawned by the Spawner along walls and other environmental
// features. They remain stationary in the current implementation but could be
// extended with movement patterns and attack behaviors in future versions.
//
// The Enemy struct implements the basic functionality needed for rendering
// and updating enemy state, while specific behaviors would be implemented
// in the AI system.
type Enemy struct {
	Type              string        // Type of enemy (e.g., "Pilz", "Kristall", etc.)
	Position          math.Vector   // Position in world coordinates relative to center
	Rotation          float64       // Rotation angle in degrees (0-360)
	Image             *ebiten.Image // Cached enemy sprite for rendering
	ImagePath         string        // Path to the enemy image in the assets system
	TimeSinceLastShot float64       // Time elapsed since the enemy last fired

	// Health and collision related fields
	Health            float64       // Current health points
	MaxHealth         float64       // Maximum health points
	IsDying           bool          // Whether the enemy is in the death animation
	DeathTimer        float64       // Timer for tracking death animation
	ExplosionFrame    int           // Current frame of the explosion animation
	ExplosionImage    *ebiten.Image // Explosion sprite sheet
}

// NewEnemy creates a new enemy with the specified parameters.
// This factory function initializes an Enemy instance with the provided
// type, position, rotation, and image path. The actual image loading is
// deferred until the first Update call to improve performance during
// mass enemy creation.
//
// Parameters:
// - enemyType: The type of enemy to create (e.g., "Pilz", "Kristall")
// - x, y: The position coordinates relative to the world center
// - rotation: The rotation angle in degrees (0-360)
// - imagePath: The path to the enemy's sprite in the assets system
//
// The created enemy is not automatically added to the game world;
// the caller is responsible for storing and managing the returned enemy.
// EnemyHealthByType maps enemy types to their maximum health values
var EnemyHealthByType = map[string]float64{
	"Pilz":     50.0,
	"Kristall": 75.0,
	"Default":  30.0,
}

func NewEnemy(enemyType string, x, y float64, rotation float64, imagePath string) *Enemy {
	// Determine the maximum health based on enemy type
	maxHealth := EnemyHealthByType["Default"]
	if health, exists := EnemyHealthByType[enemyType]; exists {
		maxHealth = health
	}

	return &Enemy{
		Type:              enemyType,
		Position:          math.Vector{X: x, Y: y},
		Rotation:          rotation,
		ImagePath:         imagePath,
		TimeSinceLastShot: 0,

		// Initialize health-related fields
		Health:            maxHealth,
		MaxHealth:         maxHealth,
		IsDying:           false,
		DeathTimer:        0,
		ExplosionFrame:    0,
		ExplosionImage:    nil, // Will be loaded when needed
	}
}

// Constants for enemy behavior
const (
	ExplosionFrameCount = 8    // Number of frames in the explosion animation
	ExplosionFrameTime  = 0.1  // Time per frame in seconds
	ExplosionScale      = 1.0  // Scale factor for the explosion animation
)

// Update updates the enemy's state for the current frame.
// This method is called once per frame for each active enemy and handles:
// 1. Lazy loading of the enemy's sprite image (only when first needed)
// 2. Physics interactions with the environment
// 3. Death animation if the enemy is dying
// 4. Any state changes or animations
//
// This method returns an error if the update fails, which can be used
// to signal that the enemy should be removed or that the game should
// handle an error condition.
//
// Returns:
// - bool: True if the enemy should be removed from the game, false otherwise
func (e *Enemy) Update(deltaTime float64) bool {
	// If the enemy is in the death animation
	if e.IsDying {
		// Load explosion image if not already loaded
		if e.ExplosionImage == nil {
			e.ExplosionImage = assets.GetImage("images/gameScene/Explosion/Explosion.png")
		}

		// Update death timer
		e.DeathTimer += deltaTime

		// Calculate current explosion frame based on timer
		frameTime := e.DeathTimer / ExplosionFrameTime
		e.ExplosionFrame = int(frameTime)

		// If the animation is complete, signal that the enemy should be removed
		if e.ExplosionFrame >= ExplosionFrameCount {
			return true // Remove the enemy
		}

		return false // Keep the enemy until animation completes
	}

	// Load image if not already loaded - lazy initialization
	// This defers image loading until actually needed and visible
	if e.Image == nil {
		// Load the image from assets using the cached path
		e.Image = assets.GetImage(e.ImagePath)
	}

	// Apply gravity to the enemy's position with a high velocity value (100.0)
	// Since this is above the LowVelocityThreshold (10.0), gravity is NOT applied
	// This ensures enemies stay at their original spawn positions
	// The deltaTime parameter represents a time step of one frame
	e.Position = physics.ApplyGravity(e.Position, 100.0, deltaTime)

	return false // Don't remove the enemy
}

// TakeDamage reduces the enemy's health by the specified amount.
// If health reaches zero or below, the enemy starts its death animation.
//
// Parameters:
// - amount: The amount of damage to take
//
// Returns:
// - bool: True if the enemy died from this damage, false otherwise
func (e *Enemy) TakeDamage(amount float64) bool {
	// If already dying, ignore damage
	if e.IsDying {
		return false
	}

	// Reduce health by the damage amount
	e.Health -= amount

	// Check if the enemy has died
	if e.Health <= 0 {
		e.Health = 0
		e.IsDying = true
		e.DeathTimer = 0
		e.ExplosionFrame = 0
		return true
	}

	return false
}

// GetCollider returns a circular collider for the enemy.
// This is used for collision detection with the player and projectiles.
//
// Returns:
// - physics.CircleCollider: The enemy's collision area
func (e *Enemy) GetCollider() physics.CircleCollider {
	// Skip collision if the enemy is dying
	if e.IsDying {
		return physics.CircleCollider{Position: e.Position, Radius: 0}
	}

	// Ensure the image is loaded before using it for collision detection
	if e.Image == nil {
		e.Image = assets.GetImage(e.ImagePath)
	}

	// If the image is still nil (could happen if the asset doesn't exist),
	// return a default collider with a reasonable radius
	if e.Image == nil {
		return physics.CircleCollider{
			Position: e.Position,
			Radius:   15.0, // Default radius for enemies
		}
	}

	return physics.GetEntityCollider(e.Position, e.Image, 0.5)
}

// Draw renders the enemy on the screen with proper transformations.
// This method handles all aspects of enemy visualization, including:
// - Positioning relative to the world center and camera
// - Rotation to match the enemy's orientation
// - Scaling to the appropriate size
// - Death animation with explosion sprites
//
// Parameters:
// - screen: The target image where the enemy should be drawn
// - offsetX, offsetY: Camera offset values for scrolling
// - worldWidth, worldHeight: Current dimensions of the game world
//
// The drawing process follows these steps:
// 1. Check if the enemy is dying and render explosion animation if so
// 2. Skip rendering if the image hasn't been loaded yet
// 3. Set up transformation options for proper rendering
// 4. Center the sprite on its origin point for accurate rotation
// 5. Apply rotation based on the enemy's orientation
// 6. Scale the sprite to the appropriate size
// 7. Calculate the final screen position considering:
//   - World center
//   - Enemy's position relative to center
//   - Camera offset for scrolling
//
// 8. Apply the final position transformation
// 9. Render the sprite to the screen
func (e *Enemy) Draw(screen *ebiten.Image, offsetX, offsetY float64, worldWidth, worldHeight int) {
	// If the enemy is dying, render the explosion animation
	if e.IsDying {
		// Skip rendering if the explosion image hasn't been loaded yet
		if e.ExplosionImage == nil {
			return
		}

		// Create transformation options for rendering
		op := &ebiten.DrawImageOptions{}

		// Calculate the dimensions of a single explosion frame
		// The explosion sprite sheet is a horizontal strip of frames
		explosionWidth := e.ExplosionImage.Bounds().Dx() / ExplosionFrameCount
		explosionHeight := e.ExplosionImage.Bounds().Dy()

		// Set the origin to the center of the explosion frame
		op.GeoM.Translate(-float64(explosionWidth)/2, -float64(explosionHeight)/2)

		// Apply scaling for the explosion
		op.GeoM.Scale(ExplosionScale, ExplosionScale)

		// Calculate the screen center
		centerX := float64(worldWidth) / 2
		centerY := float64(worldHeight) / 2

		// Calculate the screen position for the explosion
		screenX := centerX + e.Position.X + offsetX
		screenY := centerY + e.Position.Y + offsetY

		// Move to the calculated position, adjusting for the scaling factor
		op.GeoM.Translate(screenX*(1/ExplosionScale), screenY*(1/ExplosionScale))

		// Calculate the source rectangle for the current explosion frame
		frameX := e.ExplosionFrame * explosionWidth
		frameY := 0

		// Draw the current explosion frame
		screen.DrawImage(
			e.ExplosionImage.SubImage(image.Rect(
				frameX, frameY,
				frameX+explosionWidth, frameY+explosionHeight,
			)).(*ebiten.Image),
			op,
		)

		return
	}

	// Skip rendering if the image hasn't been loaded yet
	if e.Image == nil {
		return
	}

	// Create transformation options for rendering
	op := &ebiten.DrawImageOptions{}

	// Set the origin to the center of the image for rotation
	// This ensures the sprite rotates around its center point
	width, height := e.Image.Bounds().Dx(), e.Image.Bounds().Dy()
	op.GeoM.Translate(-float64(width)/2, -float64(height)/2)

	// Apply rotation, converting from degrees to radians
	op.GeoM.Rotate(e.Rotation * (stdmath.Pi / 180.0))

	// Apply scaling to make enemies slightly larger than the player
	// Player is at 1/3 scale, so we use 0.5 scale for enemies
	// This makes enemies approximately 1.5x the size of the player
	op.GeoM.Scale(0.5, 0.5)

	// Calculate the screen center using the provided world dimensions
	// This is the reference point for all world-space coordinates
	centerX := float64(worldWidth) / 2
	centerY := float64(worldHeight) / 2

	// Calculate the screen position for this enemy
	// 1. Start at the center of the screen
	// 2. Add the enemy's world position (which is relative to center)
	// 3. Apply the camera offset for scrolling
	screenX := centerX + e.Position.X + offsetX
	screenY := centerY + e.Position.Y + offsetY

	// Move to the calculated position, adjusting for the scaling factor
	// Since we're scaling by 0.5, we need to multiply the screen position by 2.0
	// to compensate for the scaling effect on the translation
	op.GeoM.Translate(screenX*2.0, screenY*2.0)

	// Draw the enemy sprite with all transformations applied
	screen.DrawImage(e.Image, op)
}
