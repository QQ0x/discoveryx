package ecs

// BasicWorld is a simple implementation of the World interface.
// It provides a lightweight, minimal implementation that stores only
// the essential properties needed for world dimensions without any
// additional ECS functionality like entity or system management.
//
// This implementation is suitable for:
// - Simple games with minimal entity management needs
// - Prototyping and testing
// - Scenes that don't require complex entity interactions
// - UI screens where only dimensions matter
//
// For more complex games, this can be extended or replaced with a more
// feature-rich implementation that includes entity and system management.
type BasicWorld struct {
	width        int  // The width of the game world in logical units
	height       int  // The height of the game world in logical units
	matchScreen  bool // Whether world dimensions should match screen dimensions
}

// NewBasicWorld creates a new BasicWorld with the specified dimensions.
// This factory function initializes a BasicWorld with:
// - The provided width and height
// - Screen matching enabled by default
//
// The initial dimensions are typically set to match the current screen size,
// but can be any values depending on the game's requirements.
//
// Example usage:
//
//	// Create a world that matches the screen size
//	world := NewBasicWorld(screenWidth, screenHeight)
//
//	// Create a fixed-size world for a level
//	world := NewBasicWorld(2000, 1000)
//	world.SetMatchScreen(false) // Disable screen matching for fixed size
func NewBasicWorld(width, height int) *BasicWorld {
	return &BasicWorld{
		width:        width,        // Initial world width
		height:       height,       // Initial world height
		matchScreen:  true,         // By default, match screen dimensions
	}
}

// GetWidth returns the width of the world in logical units.
// This implementation simply returns the stored width value.
// Game objects use this to determine horizontal boundaries.
func (w *BasicWorld) GetWidth() int {
	return w.width
}

// GetHeight returns the height of the world in logical units.
// This implementation simply returns the stored height value.
// Game objects use this to determine vertical boundaries.
func (w *BasicWorld) GetHeight() int {
	return w.height
}

// SetWidth sets the width of the world in logical units.
// This implementation directly updates the stored width value.
// This method is typically called during initialization, when
// changing levels, or when the screen size changes (if matchScreen is true).
func (w *BasicWorld) SetWidth(width int) {
	w.width = width
}

// SetHeight sets the height of the world in logical units.
// This implementation directly updates the stored height value.
// This method is typically called during initialization, when
// changing levels, or when the screen size changes (if matchScreen is true).
func (w *BasicWorld) SetHeight(height int) {
	w.height = height
}

// ShouldMatchScreen returns true if the world dimensions should
// automatically match the screen dimensions.
// This implementation simply returns the stored matchScreen flag.
// The game engine uses this to determine whether to update world
// dimensions when the screen size changes.
func (w *BasicWorld) ShouldMatchScreen() bool {
	return w.matchScreen
}

// SetMatchScreen sets whether the world dimensions should
// automatically match the screen dimensions.
// This implementation directly updates the stored matchScreen flag.
// Different game scenes may require different settings:
// - UI scenes typically match the screen (true)
// - Game levels typically have fixed dimensions (false)
func (w *BasicWorld) SetMatchScreen(match bool) {
	w.matchScreen = match
}
