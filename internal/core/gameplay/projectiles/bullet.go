package projectiles

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/ecs"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	stdmath "math"
	"time"
)

// Bullet represents a bullet entity
type Bullet struct {
	sprite    *ebiten.Image
	world     ecs.World
	position  math.Vector
	velocity  float64
	direction float64
	active    bool
	createdAt time.Time
	lifetime  time.Duration
}

// NewBullet creates a new bullet instance
func NewBullet(world ecs.World, position math.Vector, direction float64, velocity float64) *Bullet {
	// Check if the bullet sprite is nil
	if assets.PlayerBullet == nil {
		panic("PlayerBullet sprite is nil! Asset not loaded correctly.")
	}

	println("Creating new bullet at position:", position.X, position.Y)
	println("Bullet direction:", direction, "velocity:", velocity)

	// Increase velocity for testing
	testVelocity := velocity * 2.0
	println("Using increased test velocity:", testVelocity)

	bullet := &Bullet{
		sprite:    assets.PlayerBullet,
		world:     world,
		position:  position,
		direction: direction,
		velocity:  testVelocity, // Use increased velocity for testing
		active:    true,
		createdAt: time.Now(),
		lifetime:  time.Second * 10, // Increased from 2 to 10 seconds for testing
	}

	println("Created bullet with active:", bullet.active)
	println("Bullet lifetime set to", bullet.lifetime.Seconds(), "seconds")

	return bullet
}

// IsActive returns whether the bullet is active
func (b *Bullet) IsActive() bool {
	return b.active
}

// Deactivate deactivates the bullet
func (b *Bullet) Deactivate() {
	b.active = false
}

// GetPosition returns the bullet's position
func (b *Bullet) GetPosition() math.Vector {
	return b.position
}

// Update updates the bullet state
func (b *Bullet) Update(deltaTime float64) {
	println("Bullet.Update called. Active:", b.active, "DeltaTime:", deltaTime)
	println("Bullet position:", b.position.X, b.position.Y)

	if !b.active {
		println("Bullet is inactive, skipping update")
		return
	}

	// Check if the bullet has expired
	timeSinceCreation := time.Since(b.createdAt)
	println("Bullet lifetime:", timeSinceCreation.Seconds(), "seconds of", b.lifetime.Seconds(), "seconds")

	if timeSinceCreation > b.lifetime {
		println("Bullet has expired, deactivating")
		// For testing, extend lifetime instead of deactivating
		b.lifetime = time.Second * 10 // Extend to 10 seconds for testing
		println("Extended bullet lifetime to 10 seconds for testing")
		return
	}

	// Move the bullet in its direction
	dx := stdmath.Sin(b.direction) * b.velocity * deltaTime * 60.0
	dy := stdmath.Cos(b.direction) * -b.velocity * deltaTime * 60.0
	b.position.X += dx
	b.position.Y += dy
	println("Bullet moved to:", b.position.X, b.position.Y, "with dx:", dx, "dy:", dy)

	// Check if the bullet is out of bounds
	worldHalfWidth := float64(b.world.GetWidth()) / 2
	worldHalfHeight := float64(b.world.GetHeight()) / 2
	println("World bounds:", -worldHalfWidth, worldHalfWidth, -worldHalfHeight, worldHalfHeight)

	if b.position.X < -worldHalfWidth ||
		b.position.X > worldHalfWidth ||
		b.position.Y < -worldHalfHeight ||
		b.position.Y > worldHalfHeight {
		println("Bullet is out of bounds, deactivating")
		// For testing, teleport back to center instead of deactivating
		b.position.X = 0
		b.position.Y = 0
		println("Teleported bullet back to center for testing")
	}
}

// Draw draws the bullet
func (b *Bullet) Draw(screen *ebiten.Image, cameraOffsetX, cameraOffsetY float64) {
	println("Bullet.Draw called. Active:", b.active)

	if !b.active {
		println("Bullet is inactive, skipping draw")
		return
	}

	// Calculate sprite center
	bounds := b.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	println("Drawing bullet at position:", b.position.X, b.position.Y)
	println("Bullet sprite size:", bounds.Dx(), "x", bounds.Dy())
	println("Camera offset:", cameraOffsetX, cameraOffsetY)

	// Calculate world center
	centerX := float64(b.world.GetWidth()) / 2
	centerY := float64(b.world.GetHeight()) / 2
	println("World center:", centerX, centerY)

	// Calculate final screen position
	screenX := centerX + b.position.X + cameraOffsetX
	screenY := centerY + b.position.Y + cameraOffsetY
	println("Final screen position:", screenX, screenY)

	op := &ebiten.DrawImageOptions{}

	// Use a larger scale to make the bullet more visible
	const scale = 3.0 // Increased from 1.0 to 3.0 for much better visibility
	op.GeoM.Scale(scale, scale)
	println("Using scale factor:", scale)

	// The following sequence ensures rotation around the center of the sprite:
	// 1. Move to origin (center the sprite at 0,0)
	op.GeoM.Translate(-halfW, -halfH)

	// 2. Rotate around center point (which is now at 0,0)
	op.GeoM.Rotate(b.direction)

	// 3. Move to world center
	op.GeoM.Translate(centerX, centerY)

	// 4. Apply bullet position offset
	op.GeoM.Translate(b.position.X, b.position.Y)

	// 5. Apply camera offset
	op.GeoM.Translate(cameraOffsetX, cameraOffsetY)

	// Make bullet bright red for better visibility
	op.ColorM.Scale(2.0, 0.5, 0.5, 1.0) // Bright red tint

	// Draw sprite
	screen.DrawImage(b.sprite, op)
	println("Drew bullet sprite")

	// Draw a large colored rectangle at the same position for extra visibility
	rectSize := 20.0 // 20x20 pixel rectangle
	rect := ebiten.NewImage(int(rectSize), int(rectSize))
	rect.Fill(color.RGBA{255, 0, 0, 128}) // Semi-transparent red

	rectOp := &ebiten.DrawImageOptions{}
	// Position rectangle at the same position as the bullet, but offset to center it
	rectOp.GeoM.Translate(screenX - rectSize/2, screenY - rectSize/2)

	// Draw the rectangle
	screen.DrawImage(rect, rectOp)
	println("Drew additional rectangle for visibility")
}
