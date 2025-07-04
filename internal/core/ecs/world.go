// Package ecs implements an Entity Component System (ECS) architecture for the game.
// ECS is a software architectural pattern commonly used in game development that
// follows composition over inheritance, allowing for greater flexibility and
// performance in complex game systems.
//
// The ECS architecture consists of three main concepts:
// - Entities: Game objects that are composed of components (e.g., players, enemies)
// - Components: Data containers that define aspects of entities (e.g., position, health)
// - Systems: Logic that processes entities with specific components (e.g., rendering, physics)
//
// This package provides the foundation for organizing game objects and their behaviors
// in a modular, efficient, and maintainable way.
package ecs

// World represents the game world with dimensions that entities exist within,
// not just screen dimensions. It provides methods to get and set the world's
// width and height, as well as manage the relationship between world and screen dimensions.
//
// The World interface is a core part of the ECS architecture, serving as the container
// for all entities and systems. It defines the spatial boundaries within which
// game objects exist and interact.
//
// The World dimensions may be different from the screen dimensions, depending on
// the game's design. For example:
// - A 2D platformer might have a world that is much larger than the screen,
//   requiring camera movement to follow the player
// - A puzzle game might have a world that exactly matches the screen dimensions
// - A strategy game might have a world with a fixed size that can be zoomed in/out
//
// All game components should use the World interface to get world dimensions
// instead of directly accessing screen dimensions to ensure proper positioning
// and collision detection regardless of the view being displayed.
type World interface {
	// GetWidth returns the width of the game world in logical units.
	// This defines the horizontal boundary for entity positioning and movement.
	// Game objects typically cannot move beyond these boundaries (0 to Width).
	GetWidth() int

	// GetHeight returns the height of the game world in logical units.
	// This defines the vertical boundary for entity positioning and movement.
	// Game objects typically cannot move beyond these boundaries (0 to Height).
	GetHeight() int

	// SetWidth sets the width of the game world in logical units.
	// This is typically called during initialization or when changing levels.
	// Systems that depend on world boundaries (like physics or spawning)
	// should be notified when this changes.
	SetWidth(width int)

	// SetHeight sets the height of the game world in logical units.
	// This is typically called during initialization or when changing levels.
	// Systems that depend on world boundaries (like physics or spawning)
	// should be notified when this changes.
	SetHeight(height int)

	// ShouldMatchScreen returns true if the world dimensions should
	// automatically match the screen dimensions.
	// This is useful for UI-focused scenes or games where the playable
	// area should always fill the entire screen regardless of resolution.
	ShouldMatchScreen() bool

	// SetMatchScreen sets whether the world dimensions should
	// automatically match the screen dimensions.
	// When set to true, the world will resize whenever the screen size changes.
	// When set to false, the world maintains a fixed size regardless of screen size.
	SetMatchScreen(match bool)
}
