package ecs

// World represents the game world with dimensions that entities exist within,
// not just screen dimensions. It provides methods to get and set the world's
// width and height.
//
// The World dimensions may be different from the screen dimensions, depending on
// the game's design. For example, a 2D platformer might have a world that is much
// larger than the screen, while a puzzle game might have a world that exactly
// matches the screen.
//
// All game components should use the World interface to get world dimensions
// instead of directly accessing screen dimensions.
type World interface {
	// GetWidth returns the width of the game world
	GetWidth() int

	// GetHeight returns the height of the game world
	GetHeight() int

	// SetWidth sets the width of the game world
	SetWidth(width int)

	// SetHeight sets the height of the game world
	SetHeight(height int)

	// ShouldMatchScreen returns true if the world dimensions should
	// automatically match the screen dimensions
	ShouldMatchScreen() bool

	// SetMatchScreen sets whether the world dimensions should
	// automatically match the screen dimensions
	SetMatchScreen(match bool)
}
