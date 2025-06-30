package screen

import (
	"discoveryx/internal/constants"
	"log"
)

// Manager handles all screen-related functionality, including:
// - Dynamic resizing
// - Screen dimensions
// - Coordinate transformations
// - Layout calculations
//
// Manager is the single source of truth for screen dimensions in the application.
// All components should use this manager to get screen dimensions instead of
// calling ebiten.WindowSize() or screen.Size() directly.
type Manager struct {
	width           int
	height          int
	dynamicResizing bool
}

// New creates a new screen manager with default settings
func New() *Manager {
	return &Manager{
		width:           constants.ScreenWidth,
		height:          constants.ScreenHeight,
		dynamicResizing: true, // Enable dynamic resizing by default
	}
}

// GetWidth returns the current screen width
func (m *Manager) GetWidth() int {
	return m.width
}

// GetHeight returns the current screen height
func (m *Manager) GetHeight() int {
	return m.height
}

// SetDimensions updates the screen dimensions
// Returns true if dimensions were changed
func (m *Manager) SetDimensions(width, height int) bool {
	if m.width != width || m.height != height {
		m.width = width
		m.height = height

		// Only log if debug logging is enabled
		if constants.DebugLogging {
			log.Printf("Screen dimensions set to: %dx%d", width, height)
		}

		return true
	}
	return false
}

// IsDynamicResizingEnabled returns whether dynamic resizing is enabled
func (m *Manager) IsDynamicResizingEnabled() bool {
	return m.dynamicResizing
}

// SetDynamicResizing enables or disables dynamic screen resizing
func (m *Manager) SetDynamicResizing(enabled bool) {
	m.dynamicResizing = enabled
}

// CalculateLayout determines the logical screen size based on the actual window size
// This is used by the game's Layout method
func (m *Manager) CalculateLayout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if m.dynamicResizing {
		// Update dimensions if they've changed
		m.SetDimensions(outsideWidth, outsideHeight)
		return outsideWidth, outsideHeight
	} else {
		// When dynamic resizing is disabled, maintain fixed dimensions
		return constants.ScreenWidth, constants.ScreenHeight
	}
}

// GetHalfWidth returns half of the screen width, useful for UI layout
func (m *Manager) GetHalfWidth() int {
	return m.width / 2
}

// GetHalfHeight returns half of the screen height, useful for UI layout
func (m *Manager) GetHalfHeight() int {
	return m.height / 2
}

// DefaultManager is a singleton instance of the screen manager
var DefaultManager = New()

// GetWidth provides easy access to the default screen manager's width
func GetWidth() int {
	return DefaultManager.GetWidth()
}

// GetHeight provides easy access to the default screen manager's height
func GetHeight() int {
	return DefaultManager.GetHeight()
}

// SetDimensions provides easy access to the default screen manager's SetDimensions method
func SetDimensions(width, height int) bool {
	return DefaultManager.SetDimensions(width, height)
}
