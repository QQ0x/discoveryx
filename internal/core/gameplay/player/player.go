package player

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/ecs"
	"discoveryx/internal/input"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
)

// Constants for player movement
const (
	rotationPerSecond = 3.0  // Rotation speed in radians per second
	maxAcceleration   = 10.0 // Maximum acceleration value
)

// Player represents the player entity
type Player struct {
	sprite         *ebiten.Image
	world          ecs.World
	rotation       float64
	position       math.Vector
	playerVelocity float64 // Player speed
	curAcceleration float64 // Current acceleration
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

// HandleRotation updates the player's rotation based on input
func (p *Player) HandleRotation(keyboard input.KeyboardHandler) {
	speed := rotationPerSecond / float64(ebiten.TPS())

	if keyboard.IsKeyPressed(input.KeyLeft) {
		p.rotation -= speed
	}

	if keyboard.IsKeyPressed(input.KeyRight) {
		p.rotation += speed
	}
}

// HandleAcceleration handles player acceleration and updates position
func (p *Player) HandleAcceleration(keyboard input.KeyboardHandler) {
	if keyboard.IsKeyPressed(input.KeyUp) {
		if p.curAcceleration < maxAcceleration {
			p.curAcceleration = p.playerVelocity + 4
		}

		if p.curAcceleration >= 8 {
			p.curAcceleration = 8
		}

		p.playerVelocity = p.curAcceleration

		// Move in the direction we are pointing
		dx := stdmath.Sin(p.rotation) * p.curAcceleration
		dy := stdmath.Cos(p.rotation) * -p.curAcceleration

		// Move the player on screen
		p.position.X += dx
		p.position.Y += dy
	}
}

// Update updates the player state
func (p *Player) Update() error {
	// Get keyboard handler from input manager
	keyboard := input.GetKeyboard()

	// Handle player movement
	p.HandleRotation(keyboard)
	p.HandleAcceleration(keyboard)

	return nil
}
