// Package ecs provides entity-component-system functionality.
package ecs

// BasicWorld is a simple implementation of the World interface.
// It provides basic width and height properties.
type BasicWorld struct {
	width        int
	height       int
	matchScreen  bool
}

// NewBasicWorld creates a new BasicWorld with the specified dimensions.
// By default, the world dimensions will match the screen dimensions.
func NewBasicWorld(width, height int) *BasicWorld {
	return &BasicWorld{
		width:        width,
		height:       height,
		matchScreen:  true, // By default, match screen dimensions
	}
}

// GetWidth returns the width of the world.
func (w *BasicWorld) GetWidth() int {
	return w.width
}

// GetHeight returns the height of the world.
func (w *BasicWorld) GetHeight() int {
	return w.height
}

// SetWidth sets the width of the world.
func (w *BasicWorld) SetWidth(width int) {
	w.width = width
}

// SetHeight sets the height of the world.
func (w *BasicWorld) SetHeight(height int) {
	w.height = height
}

// ShouldMatchScreen returns true if the world dimensions should
// automatically match the screen dimensions.
func (w *BasicWorld) ShouldMatchScreen() bool {
	return w.matchScreen
}

// SetMatchScreen sets whether the world dimensions should
// automatically match the screen dimensions.
func (w *BasicWorld) SetMatchScreen(match bool) {
	w.matchScreen = match
}
