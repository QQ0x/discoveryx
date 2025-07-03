package input

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// KeyboardHandler provides an abstraction for keyboard input
type KeyboardHandler interface {
	IsKeyPressed(key ebiten.Key) bool
}

// DefaultKeyboardHandler is the default implementation of KeyboardHandler
type DefaultKeyboardHandler struct{}

// NewKeyboardHandler creates a new default keyboard handler
func NewKeyboardHandler() KeyboardHandler {
	return &DefaultKeyboardHandler{}
}

// IsKeyPressed checks if a specific key is currently pressed
func (h *DefaultKeyboardHandler) IsKeyPressed(key ebiten.Key) bool {
	return ebiten.IsKeyPressed(key)
}

// Common key constants for easier access
const (
	KeyLeft  = ebiten.KeyLeft
	KeyRight = ebiten.KeyRight
	KeyUp    = ebiten.KeyUp
	KeyDown  = ebiten.KeyDown
	KeySpace = ebiten.KeySpace
)
