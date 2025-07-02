package enemies

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/physics"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
)

// Enemy represents an enemy entity in the game
type Enemy struct {
	Type      string      // Type of enemy (e.g., "Pilz", "Kristall", etc.)
	Position  math.Vector // Position in world coordinates
	Rotation  float64     // Rotation angle in degrees
	Image     *ebiten.Image // Enemy image
	ImagePath string      // Path to the enemy image
}

// NewEnemy creates a new enemy with the specified parameters
func NewEnemy(enemyType string, x, y float64, rotation float64, imagePath string) *Enemy {
	return &Enemy{
		Type:      enemyType,
		Position:  math.Vector{X: x, Y: y},
		Rotation:  rotation,
		ImagePath: imagePath,
	}
}

// Update updates the enemy's state
func (e *Enemy) Update() error {
	// Load image if not already loaded
	if e.Image == nil {
		// Load the image from assets
		e.Image = assets.GetImage(e.ImagePath)
	}

	// Apply gravity to the enemy's position with a high velocity value (100.0)
	// Since this is above the LowVelocityThreshold (10.0), gravity is NOT applied
	// This ensures enemies stay at their original spawn positions
	e.Position = physics.ApplyGravity(e.Position, 100.0, 1.0/60.0)

	return nil
}

// Draw draws the enemy on the screen
func (e *Enemy) Draw(screen *ebiten.Image, offsetX, offsetY float64, worldWidth, worldHeight int) {
	if e.Image == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}

	// Set the origin to the center of the image for rotation
	width, height := e.Image.Bounds().Dx(), e.Image.Bounds().Dy()
	op.GeoM.Translate(-float64(width)/2, -float64(height)/2)

	// Apply rotation
	op.GeoM.Rotate(e.Rotation * (stdmath.Pi / 180.0))

	// Apply scaling to make enemies slightly larger than the player
	// Player is at 1/3 scale, so we use 0.5 scale for enemies
	op.GeoM.Scale(0.5, 0.5)

	// Calculate the screen center using the provided world dimensions
	centerX := float64(worldWidth) / 2
	centerY := float64(worldHeight) / 2

	// Calculate the screen position for this enemy
	// 1. Start at the center of the screen
	// 2. Add the enemy's world position
	// 3. Apply the camera offset
	screenX := centerX + e.Position.X + offsetX
	screenY := centerY + e.Position.Y + offsetY

	// Move to the calculated position, adjusting for the scaling factor
	// Since we're scaling by 2.0, we need to multiply the screen position by 2.0
	op.GeoM.Translate(screenX*2.0, screenY*2.0)

	screen.DrawImage(e.Image, op)
}
