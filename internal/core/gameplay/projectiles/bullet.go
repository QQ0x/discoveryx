// Package projectiles implements various projectile types and effects used in combat.
// It provides functionality for creating, updating, and rendering different types
// of projectiles such as bullets, lasers, and explosions. The package handles
// projectile movement, collision detection, and visual effects.
//
// Projectiles are a core gameplay element that enables combat interactions
// between the player and enemies. Each projectile type has unique behavior,
// appearance, and effects.
package projectiles

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
)

// Bullet behavior constants control the movement and lifetime of bullets.
// These values are carefully tuned to create a satisfying projectile feel
// with bullets that start relatively slow but quickly accelerate.
const (
	bulletInitialSpeed = 3.0  // Starting speed of bullets in units per frame
	bulletAcceleration = 1.05 // Multiplicative acceleration factor per frame (5% increase)
	bulletMaxLifetime  = 2.0  // Maximum lifetime in seconds before automatic despawn
)

// Bullet represents a projectile fired by the player.
// It accelerates exponentially until its lifetime expires, creating a
// dynamic feel where bullets start relatively slow but quickly gain speed.
// This acceleration pattern creates a visually interesting trail effect
// and gives players a sense of power as their shots rapidly accelerate.
//
// Bullets are automatically despawned after their lifetime expires or when
// they collide with enemies or obstacles (handled by the collision system).
type Bullet struct {
	Position math.Vector   // Current position in world coordinates relative to center
	Rotation float64       // Current rotation in radians (0 = up, increases clockwise)
	speed    float64       // Current speed in units per frame (increases over time)
	lifetime float64       // Current lifetime in seconds (increases until max)
	Image    *ebiten.Image // Sprite used to render the bullet
}

// NewBullet creates a new bullet at the given position and rotation.
// This factory function initializes a Bullet instance with the provided
// position and rotation, setting its initial speed from the bulletInitialSpeed
// constant and starting its lifetime at zero.
//
// Parameters:
//   - pos: The starting position vector for the bullet (typically the player's position
//     or a position offset from the player to represent the weapon muzzle)
//   - rotation: The direction in which the bullet will travel, in radians
//
// The created bullet is not automatically added to the game world;
// the caller is responsible for storing and managing the returned bullet.
func NewBullet(pos math.Vector, rotation float64, img *ebiten.Image) *Bullet {
	return &Bullet{
		Position: pos,
		Rotation: rotation,
		speed:    bulletInitialSpeed, // Start with the base speed
		lifetime: 0,                  // Initialize lifetime to zero
		Image:    img,
	}
}

// Update moves the bullet and updates its state for the current frame.
// This method is called once per frame for each active bullet and handles:
// 1. Applying exponential acceleration to increase the bullet's speed
// 2. Calculating and applying movement based on the bullet's rotation and speed
// 3. Updating the bullet's lifetime and checking for expiration
//
// The exponential acceleration creates a distinctive visual effect where
// bullets start relatively slow but quickly gain speed, creating a sense
// of power and momentum.
//
// Parameters:
// - deltaTime: The time elapsed since the last frame in seconds
//
// Returns:
// - true if the bullet's lifetime has expired and it should be removed
// - false if the bullet is still active and should continue to exist
func (b *Bullet) Update(deltaTime float64) bool {
	// Apply exponential acceleration to increase speed over time
	// The bulletAcceleration value is raised to the power of deltaTime*60.0
	// to ensure consistent acceleration regardless of frame rate
	b.speed *= stdmath.Pow(bulletAcceleration, deltaTime*60.0)

	// Calculate movement vector based on rotation and speed
	// sin(rotation) gives X component, cos(rotation) gives Y component
	// Note: Y is negated because in screen coordinates, Y increases downward
	dx := stdmath.Sin(b.Rotation) * b.speed * deltaTime * 60.0
	dy := stdmath.Cos(b.Rotation) * -b.speed * deltaTime * 60.0

	// Update position by adding the movement vector
	b.Position.X += dx
	b.Position.Y += dy

	// Increment lifetime and check if it has expired
	b.lifetime += deltaTime
	return b.lifetime >= bulletMaxLifetime
}

// Draw renders the bullet on the screen with proper transformations.
// This method handles all aspects of bullet visualization, including:
// - Loading the bullet sprite from assets
// - Scaling the sprite to the appropriate size
// - Rotating the sprite to match the bullet's direction
// - Positioning the sprite at the bullet's location in the world
// - Applying camera offsets for scrolling
//
// Parameters:
// - screen: The target image where the bullet should be drawn
// - offsetX, offsetY: Camera offset values for scrolling
// - worldWidth, worldHeight: Current dimensions of the game world
//
// The drawing process follows these steps:
// 1. Get the bullet sprite from assets
// 2. Set up transformation options for proper rendering
// 3. Apply scaling to size the bullet appropriately
// 4. Center the sprite on its origin point for accurate rotation
// 5. Apply rotation based on the bullet's direction
// 6. Calculate the final screen position considering:
//   - World center
//   - Bullet's position relative to center
//   - Camera offset for scrolling
//
// 7. Render the sprite to the screen
func (b *Bullet) Draw(screen *ebiten.Image, offsetX, offsetY float64, worldWidth, worldHeight int) {
	// Use the bullet's image, defaulting to the player bullet if nil
	img := b.Image
	if img == nil {
		img = assets.PlayerBullet
	}

	// Create transformation options for rendering
	op := &ebiten.DrawImageOptions{}

	// Apply scaling to size the bullet appropriately
	// A scale of 0.5 makes the bullet half its original size
	scale := 0.5
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	op.GeoM.Scale(scale, scale)

	// Center the sprite on its origin point for accurate rotation
	// This ensures the bullet rotates around its center
	op.GeoM.Translate(-float64(w)*scale/2, -float64(h)*scale/2)

	// Apply rotation to match the bullet's direction
	op.GeoM.Rotate(b.Rotation)

	// Calculate the screen center using the provided world dimensions
	// This is the reference point for all world-space coordinates
	centerX := float64(worldWidth) / 2
	centerY := float64(worldHeight) / 2

	// Calculate and apply the final screen position
	// This combines the world center, bullet position, and camera offset
	op.GeoM.Translate(centerX+b.Position.X+offsetX, centerY+b.Position.Y+offsetY)

	// Draw the bullet sprite with all transformations applied
	screen.DrawImage(img, op)
}
