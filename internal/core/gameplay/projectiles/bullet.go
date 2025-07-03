package projectiles

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
)

const (
	bulletInitialSpeed = 3.0
	bulletAcceleration = 1.05
	bulletMaxLifetime  = 2.0
)

// Bullet represents a projectile fired by the player.
// It accelerates exponentially until its lifetime expires.
type Bullet struct {
	Position math.Vector
	Rotation float64
	speed    float64
	lifetime float64
}

// NewBullet creates a new bullet at the given position and rotation.
func NewBullet(pos math.Vector, rotation float64) *Bullet {
	return &Bullet{
		Position: pos,
		Rotation: rotation,
		speed:    bulletInitialSpeed,
	}
}

// Update moves the bullet. It returns true if the bullet's lifetime has expired.
func (b *Bullet) Update(deltaTime float64) bool {
	// Exponential acceleration
	b.speed *= stdmath.Pow(bulletAcceleration, deltaTime*60.0)

	dx := stdmath.Sin(b.Rotation) * b.speed * deltaTime * 60.0
	dy := stdmath.Cos(b.Rotation) * -b.speed * deltaTime * 60.0
	b.Position.X += dx
	b.Position.Y += dy

	b.lifetime += deltaTime
	return b.lifetime >= bulletMaxLifetime
}

// Draw renders the bullet on the screen.
func (b *Bullet) Draw(screen *ebiten.Image, offsetX, offsetY float64, worldWidth, worldHeight int) {
	img := assets.PlayerBullet
	op := &ebiten.DrawImageOptions{}

	scale := 0.5
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(-float64(w)*scale/2, -float64(h)*scale/2)
	op.GeoM.Rotate(b.Rotation)

	centerX := float64(worldWidth) / 2
	centerY := float64(worldHeight) / 2
	op.GeoM.Translate(centerX+b.Position.X+offsetX, centerY+b.Position.Y+offsetY)

	screen.DrawImage(img, op)
}
