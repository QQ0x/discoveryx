package player

import (
	"discoveryx/internal/constants"
	"discoveryx/internal/core/physics"
	"discoveryx/internal/utils/math"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"time"
)

// Health-related constants
const (
	MaxPlayerHealth        = 100.0 // Maximum health points for the player
	InvincibilityDuration  = 1.5   // Duration of invincibility in seconds after taking damage
	InvincibilityFlashRate = 0.1   // Rate at which the player flashes during invincibility (in seconds)
	WallCollisionDamage    = 5.0   // Damage taken when colliding with walls
	EnemyCollisionDamage   = 10.0  // Damage taken when colliding with enemies
	ProjectileDamage       = 15.0  // Damage taken when hit by enemy projectiles
)

// AddHealthToPlayer adds health-related fields to the Player struct.
// This function should be called in the NewPlayer function.
func (p *Player) AddHealthSystem() {
	p.health = MaxPlayerHealth
	p.isInvincible = false
	p.invincibilityTimer = 0
	p.shouldRender = true
}

// GetHealth returns the player's current health.
// This is used by other systems for:
// - UI health display
// - Game over detection
// - Achievement tracking
func (p *Player) GetHealth() float64 {
	return p.health
}

// TakeDamage reduces the player's health by the specified amount.
// If the player is currently invincible, no damage is taken.
// After taking damage, the player becomes invincible for a short duration.
//
// Parameters:
// - amount: The amount of damage to take
//
// Returns:
// - bool: True if the player took damage, false if invincible
func (p *Player) TakeDamage(amount float64) bool {
	// If the player is invincible, don't take damage
	if p.isInvincible {
		return false
	}

	// Reduce health by the damage amount
	p.health -= amount
	if p.health < 0 {
		p.health = 0
	}

	// Activate invincibility frames
	p.isInvincible = true
	p.invincibilityTimer = 0

	return true
}

// Heal increases the player's health by the specified amount, up to the maximum.
//
// Parameters:
// - amount: The amount of health to restore
func (p *Player) Heal(amount float64) {
	p.health += amount
	if p.health > MaxPlayerHealth {
		p.health = MaxPlayerHealth
	}
}

// UpdateHealthSystem updates the player's health-related state.
// This method should be called in the Update method.
// It handles:
// 1. Invincibility timer and visual feedback
// 2. Collision detection with walls, enemies, and projectiles
//
// Parameters:
// - deltaTime: The time elapsed since the last frame in seconds
func (p *Player) UpdateHealthSystem(deltaTime float64) {
	// Update invincibility state
	if p.isInvincible {
		// Increment the invincibility timer
		p.invincibilityTimer += deltaTime

		// Flash the player by toggling visibility
		p.shouldRender = (int(p.invincibilityTimer/InvincibilityFlashRate) % 2) == 0

		// Check if invincibility has expired
		if p.invincibilityTimer >= InvincibilityDuration {
			p.isInvincible = false
			p.shouldRender = true
		}
	}
}

// GetCollider returns a circular collider for the player.
// This is used for collision detection with enemies and projectiles.
//
// Returns:
// - physics.CircleCollider: The player's collision area
func (p *Player) GetCollider() physics.CircleCollider {
	collider := physics.GetEntityCollider(p.position, p.sprite, 1.0/3.0)

	if constants.DebugPlayerWallCollision {
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] PLAYER COLLIDER: Position=%v, Radius=%v, Velocity=%v, Rotation=%v\n", 
			timestamp, collider.Position, collider.Radius, p.playerVelocity, p.rotation)
	}

	return collider
}

// GetAABBCollider returns an AABB collider for the player.
// This is used for precise collision detection with walls.
//
// Returns:
// - physics.AABBCollider: The player's AABB collision area
func (p *Player) GetAABBCollider() physics.AABBCollider {
	collider := physics.GetAABBColliderFromSprite(p.position, p.sprite, 1.0/3.0)

	if constants.DebugPlayerWallCollision {
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] PLAYER AABB COLLIDER: Position=%v, Size=%vx%v, Velocity=%v, Rotation=%v\n", 
			timestamp, collider.Position, collider.Width, collider.Height, p.playerVelocity, p.rotation)
	}

	return collider
}

// CheckWallCollision checks if the player is colliding with a wall and handles the collision.
// This method should be called in the Update method after updating the player's position.
//
// Parameters:
// - walls: A slice of wall colliders to check against
//
// Returns:
// - bool: True if a collision occurred, false otherwise
func (p *Player) CheckWallCollision(walls []physics.RectCollider) bool {
	playerCollider := p.GetCollider()

	for _, wall := range walls {
		// Check for collision with the wall
		if collision, normal := physics.CheckCircleRectCollision(playerCollider, wall); collision {
			// Calculate the overlap depth
			depth := playerCollider.Radius - math.Distance(playerCollider.Position, math.Vector{
				X: playerCollider.Position.X + normal.X*playerCollider.Radius,
				Y: playerCollider.Position.Y + normal.Y*playerCollider.Radius,
			})

			// Resolve the collision by moving the player away from the wall
			p.position = physics.ResolveCollision(p.position, normal, depth)

			// Apply damage from wall collision
			p.TakeDamage(WallCollisionDamage)

			return true
		}
	}

	return false
}

// IsInvincible returns whether the player is currently invincible.
// This is used to check if the player can take damage from enemies or projectiles.
//
// Returns:
// - bool: True if the player is invincible, false otherwise
func (p *Player) IsInvincible() bool {
	return p.isInvincible
}

// ShouldRender returns whether the player should be rendered.
// This is used to implement the flashing effect during invincibility.
//
// Returns:
// - bool: True if the player should be rendered, false otherwise
func (p *Player) ShouldRender() bool {
	return p.shouldRender
}

// DrawWithInvincibility renders the player with invincibility visual effects.
// This method should be called instead of the regular Draw method when using the health system.
//
// Parameters:
// - screen: The target image where the player should be drawn
// - cameraOffsetX, cameraOffsetY: Camera offset values for scrolling
func (p *Player) DrawWithInvincibility(screen *ebiten.Image, cameraOffsetX, cameraOffsetY float64) {
	// Skip rendering if the player shouldn't be rendered during invincibility flashing
	if !p.shouldRender {
		return
	}

	// Apply invincibility visual effect (slightly transparent)
	op := &ebiten.DrawImageOptions{}

	if p.isInvincible {
		// Make the player slightly transparent during invincibility
		op.ColorM.Scale(1.0, 1.0, 1.0, 0.7)
	}

	// Get the dimensions of the sprite for centering
	bounds := p.sprite.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

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
