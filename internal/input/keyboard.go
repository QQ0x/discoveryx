package input

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// KeyboardHandler provides an abstraction for keyboard input across different platforms.
// This interface defines the contract for handling keyboard events, allowing the game
// to respond to keyboard input in a platform-independent way.
//
// The keyboard input system is primarily used on desktop platforms but can also
// support external keyboards on mobile devices. It provides a simple way to check
// if specific keys are currently pressed, which is used for player movement,
// actions, and menu navigation.
type KeyboardHandler interface {
	// IsKeyPressed returns true if the specified key is currently being pressed.
	// This method can be called each frame to implement continuous actions
	// (like movement) or can be combined with edge detection in the game logic
	// to implement one-time actions (like firing or jumping).
	IsKeyPressed(key ebiten.Key) bool
}

// DefaultKeyboardHandler is the default implementation of KeyboardHandler.
// It provides a thin wrapper around Ebiten's built-in keyboard functions,
// adapting them to the KeyboardHandler interface used by the game.
//
// This implementation is stateless and simply forwards calls to the underlying
// Ebiten engine, making it lightweight and efficient.
type DefaultKeyboardHandler struct{}

// NewKeyboardHandler creates a new default keyboard handler.
// This factory function returns an implementation of the KeyboardHandler
// interface that uses Ebiten's built-in keyboard input functions.
//
// Since the DefaultKeyboardHandler is stateless, this function always
// returns the same type of handler without any configuration options.
func NewKeyboardHandler() KeyboardHandler {
	return &DefaultKeyboardHandler{}
}

// IsKeyPressed checks if a specific key is currently pressed.
// This implementation directly delegates to Ebiten's IsKeyPressed function,
// providing a thin abstraction layer that could be replaced with a mock
// implementation for testing or a more complex implementation if needed.
func (h *DefaultKeyboardHandler) IsKeyPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key)
}

// Common key constants for easier access.
// These constants are aliases for Ebiten's key constants, providing
// a more convenient way to reference commonly used keys without
// having to import the Ebiten package directly in game logic code.
const (
	KeyLeft  = ebiten.KeyLeft   // Left arrow key (player movement left)
	KeyRight = ebiten.KeyRight  // Right arrow key (player movement right)
	KeyUp    = ebiten.KeyUp     // Up arrow key (player movement up)
	KeyDown  = ebiten.KeyDown   // Down arrow key (player movement down)
	KeySpace = ebiten.KeySpace  // Space key (often used for firing or jumping)
)
