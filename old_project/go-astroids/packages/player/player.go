package player

import (
	"example.com/go_astroids/assets"
	"example.com/go_astroids/internal/engine"
	// "example.com/go_astroids/internal/game"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
)

var curAcceleration float64

const (
	rotationPerSecond = math.Pi
	maxAcceleration   = 8.0
)

type Player struct {
	sprite         *ebiten.Image
	world          engine.World
	rotation       float64
	position       Vector
	playerVelocity float64 // Player Geschwindigkeit
}

// Create player instance
func NewPlayer(world engine.World) *Player {
	sprite := assets.PlayerSprite

	p := &Player{
		sprite: sprite,
		world:  world,
	}

	return p
}

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

func (p *Player) accelerate() { // Beschleunigung
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		if curAcceleration < maxAcceleration {
			curAcceleration = p.playerVelocity + 4
		}

		if curAcceleration >= 8 {
			curAcceleration = 8
		}

		p.playerVelocity = curAcceleration

		// Move in the direction we are pointing
		dx := math.Sin(p.rotation) * curAcceleration
		dy := math.Cos(p.rotation) * -curAcceleration

		// Move the player on screen
		p.position.X += dx
		p.position.Y += dy
	}
}
