package player

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/ecs"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
)

var curAcceleration float64

const (
	rotationPerSecond = stdmath.Pi
	maxAcceleration   = 8.0
)

// Player represents the player entity
type Player struct {
	sprite         *ebiten.Image
	world          ecs.World
	rotation       float64
	position       math.Vector
	playerVelocity float64 // Player speed
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

	// Move to origin
	op.GeoM.Translate(-halfW, -halfH)
	// Rotate around center
	op.GeoM.Rotate(p.rotation)
	// Move to world center
	centerX := float64(p.world.GetWidth()) / 2
	centerY := float64(p.world.GetHeight()) / 2
	op.GeoM.Translate(centerX, centerY)

	op.GeoM.Translate(p.position.X, p.position.Y)

	// Draw sprite
	screen.DrawImage(p.sprite, op)
}

// Update updates the player state
func (p *Player) Update() error {
	// Handle player movement

	speed := rotationPerSecond / float64(ebiten.TPS())

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.rotation -= speed
	}

	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.rotation += speed
	}

	p.accelerate()

	return nil
}

// accelerate handles player acceleration
func (p *Player) accelerate() {
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		if curAcceleration < maxAcceleration {
			curAcceleration = p.playerVelocity + 4
		}

		if curAcceleration >= 8 {
			curAcceleration = 8
		}

		p.playerVelocity = curAcceleration

		// Move in the direction we are pointing
		dx := stdmath.Sin(p.rotation) * curAcceleration
		dy := stdmath.Cos(p.rotation) * -curAcceleration

		// Move the player on screen
		p.position.X += dx
		p.position.Y += dy
	}
}
