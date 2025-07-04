package ecs

import "github.com/hajimehoshi/ebiten/v2"

// Entity represents a game object with update and draw capabilities.
// In this implementation, entities are active objects that encapsulate both
// data (traditionally components) and behavior (traditionally systems).
//
// This approach is a simplified version of the classic ECS pattern, where:
// - Entities in a pure ECS would be just identifiers without methods
// - Components would store all data
// - Systems would process entities with specific components
//
// The simplified approach used here combines these concepts into a single
// interface, making it easier to implement and understand for smaller games,
// while still providing the benefits of modular game object design.
//
// Common examples of entities in the game include:
// - Player ships
// - Enemy ships
// - Projectiles
// - Power-ups
// - UI elements
type Entity interface {
	// Update handles the entity's game logic for a single frame.
	// This method is called once per frame and should update the entity's
	// state based on input, time, and game rules. It should handle:
	// - Movement and physics
	// - AI and behavior
	// - Collision detection
	// - State changes
	//
	// Returns an error if the update fails, which can be used to signal
	// that the entity should be removed or that the game should handle an error.
	Update() error

	// Draw renders the entity to the provided screen.
	// This method is responsible for all visual representation of the entity,
	// including:
	// - Sprite rendering
	// - Animation
	// - Visual effects
	// - Debug visualization
	//
	// The screen parameter is the target image where the entity should be drawn.
	Draw(screen *ebiten.Image)
}
