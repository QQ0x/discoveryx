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
	"discoveryx/internal/constants"
	"discoveryx/internal/core/physics"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
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
	Health            int           // Current health of the enemy
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
func NewEnemy(enemyType string, x, y float64, rotation float64, imagePath string) *Enemy {
	return &Enemy{
		Type:              enemyType,
		Position:          math.Vector{X: x, Y: y},
		Rotation:          rotation,
		ImagePath:         imagePath,
		TimeSinceLastShot: 0,
		Health:            constants.EnemyMaxHealth,
	}
}

// Update updates the enemy's state for the current frame.
// This method is called once per frame for each active enemy and handles:
// 1. Lazy loading of the enemy's sprite image (only when first needed)
// 2. Physics interactions with the environment
// 3. Any state changes or animations (in future implementations)
//
// The current implementation is intentionally simple:
//   - The image is loaded on first update if not already loaded
//   - A high velocity value is passed to ApplyGravity to prevent movement
//     (since the physics system ignores gravity for high-velocity objects)
//
// This method returns an error if the update fails, which can be used
// to signal that the enemy should be removed or that the game should
// handle an error condition.
func (e *Enemy) Update() error {
	// Load image if not already loaded - lazy initialization
	// This defers image loading until actually needed and visible
	if e.Image == nil {
		// Load the image from assets using the cached path
		e.Image = assets.GetImage(e.ImagePath)
	}

	// Apply gravity to the enemy's position with a high velocity value (100.0)
	// Since this is above the LowVelocityThreshold (10.0), gravity is NOT applied
	// This ensures enemies stay at their original spawn positions
	// The 1.0/60.0 parameter represents a time step of one frame at 60 FPS
	e.Position = physics.ApplyGravity(e.Position, 100.0, 1.0/60.0)

	return nil
}

// TakeDamage reduces the enemy's health by the given amount.
// Health will not drop below zero.
func (e *Enemy) TakeDamage(amount int) {
	e.Health -= amount
	if e.Health < 0 {
		e.Health = 0
	}
}

// IsDead returns true if the enemy's health has reached zero.
func (e *Enemy) IsDead() bool {
	return e.Health <= 0
}

// Draw renders the enemy on the screen with proper transformations.
// This method handles all aspects of enemy visualization, including:
// - Positioning relative to the world center and camera
// - Rotation to match the enemy's orientation
// - Scaling to the appropriate size
// - Applying any visual effects (in future implementations)
//
// Parameters:
// - screen: The target image where the enemy should be drawn
// - offsetX, offsetY: Camera offset values for scrolling
// - worldWidth, worldHeight: Current dimensions of the game world
//
// The drawing process follows these steps:
// 1. Skip rendering if the image hasn't been loaded yet
// 2. Set up transformation options for proper rendering
// 3. Center the sprite on its origin point for accurate rotation
// 4. Apply rotation based on the enemy's orientation
// 5. Scale the sprite to the appropriate size
// 6. Calculate the final screen position considering:
//   - World center
//   - Enemy's position relative to center
//   - Camera offset for scrolling
//
// 7. Apply the final position transformation
// 8. Render the sprite to the screen
func (e *Enemy) Draw(screen *ebiten.Image, offsetX, offsetY float64, worldWidth, worldHeight int) {
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
