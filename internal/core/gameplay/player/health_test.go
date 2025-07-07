package player

import (
	"discoveryx/internal/core/physics"
	"discoveryx/internal/utils/math"
	"testing"
)

// MockWorld implements the World interface for testing
type MockWorld struct{
	width  int
	height int
	match  bool
}

func NewMockWorld() *MockWorld {
	return &MockWorld{
		width:  800,
		height: 600,
		match:  false,
	}
}

func (m *MockWorld) GetWidth() int  { return m.width }
func (m *MockWorld) GetHeight() int { return m.height }
func (m *MockWorld) SetWidth(width int) { m.width = width }
func (m *MockWorld) SetHeight(height int) { m.height = height }
func (m *MockWorld) ShouldMatchScreen() bool { return m.match }
func (m *MockWorld) SetMatchScreen(match bool) { m.match = match }

// TestPlayerHealth tests the player health system
func TestPlayerHealth(t *testing.T) {
	// Create a player for testing
	player := NewPlayer(NewMockWorld())

	// Test initial health
	if player.GetHealth() != MaxPlayerHealth {
		t.Errorf("Initial health should be %v, got %v", MaxPlayerHealth, player.GetHealth())
	}

	// Test taking damage
	player.TakeDamage(20)
	if player.GetHealth() != MaxPlayerHealth-20 {
		t.Errorf("Health after damage should be %v, got %v", MaxPlayerHealth-20, player.GetHealth())
	}

	// Test invincibility after taking damage
	if !player.isInvincible {
		t.Errorf("Player should be invincible after taking damage")
	}

	// Test that player doesn't take damage while invincible
	player.TakeDamage(20)
	if player.GetHealth() != MaxPlayerHealth-20 {
		t.Errorf("Health should not change while invincible, expected %v, got %v", 
			MaxPlayerHealth-20, player.GetHealth())
	}

	// Test healing
	player.Heal(10)
	if player.GetHealth() != MaxPlayerHealth-10 {
		t.Errorf("Health after healing should be %v, got %v", MaxPlayerHealth-10, player.GetHealth())
	}

	// Test healing beyond max health
	player.Heal(MaxPlayerHealth)
	if player.GetHealth() != MaxPlayerHealth {
		t.Errorf("Health should be capped at %v, got %v", MaxPlayerHealth, player.GetHealth())
	}

	// Test invincibility expiration
	player.TakeDamage(10) // Make player invincible
	player.UpdateHealthSystem(InvincibilityDuration + 0.1) // Update with time > invincibility duration
	if player.isInvincible {
		t.Errorf("Player should not be invincible after invincibility duration")
	}
}

// TestWallCollision tests the wall collision detection and damage
func TestWallCollision(t *testing.T) {
	// Create a player for testing
	player := NewPlayer(NewMockWorld())

	// Create a wall
	wall := physics.RectCollider{
		Position: math.Vector{X: 100, Y: 100},
		Width:    50,
		Height:   50,
	}

	// Position player to collide with wall
	player.SetPosition(math.Vector{X: 100, Y: 75})

	// Check for collision
	collision := player.CheckWallCollision([]physics.RectCollider{wall})

	// Test collision detection
	if !collision {
		t.Errorf("Player should collide with wall")
	}

	// Test damage from wall collision
	expectedHealth := MaxPlayerHealth - WallCollisionDamage
	if player.GetHealth() != expectedHealth {
		t.Errorf("Health after wall collision should be %v, got %v", 
			expectedHealth, player.GetHealth())
	}
}
