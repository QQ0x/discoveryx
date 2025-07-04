// Package screen provides functionality for managing screen dimensions and layout.
// It handles the complexities of different screen sizes, resolutions, and orientations
// across various platforms (desktop, mobile, etc.) and provides a consistent interface
// for the rest of the game to interact with the display.
//
// This package is particularly important for:
// - Supporting different device resolutions and aspect ratios
// - Handling window resizing on desktop platforms
// - Managing orientation changes on mobile devices
// - Providing coordinate transformations for input and rendering
// - Ensuring UI elements are positioned correctly regardless of screen size
package screen

import (
	"discoveryx/internal/constants"
	"log"
)

// Manager handles all screen-related functionality, including:
// - Dynamic resizing for adapting to different window/screen sizes
// - Screen dimensions tracking and reporting
// - Coordinate transformations between physical and logical coordinates
// - Layout calculations for proper rendering and UI positioning
//
// Manager is the single source of truth for screen dimensions in the application.
// All components should use this manager to get screen dimensions instead of
// calling ebiten.WindowSize() or screen.Size() directly to ensure consistency
// across the entire application.
//
// The Manager supports two modes of operation:
// 1. Dynamic resizing: The game adapts to fill the available screen space
// 2. Fixed dimensions: The game maintains a constant logical size regardless of screen size
type Manager struct {
	width           int  // Current logical width of the screen/window
	height          int  // Current logical height of the screen/window
	dynamicResizing bool // Whether the game should adapt to different screen sizes
}

// New creates a new screen manager with default settings.
// This factory function initializes a Manager with:
// - Width and height from constants.ScreenWidth/ScreenHeight
// - Dynamic resizing enabled by default
//
// The default settings are suitable for most game scenarios, but can be
// adjusted after creation using the SetDimensions and SetDynamicResizing methods.
func New() *Manager {
	return &Manager{
		width:           constants.ScreenWidth,  // Default width from constants
		height:          constants.ScreenHeight, // Default height from constants
		dynamicResizing: true,                   // Enable dynamic resizing by default
	}
}

// GetWidth returns the current logical screen width.
// This is used throughout the game for positioning elements and
// determining boundaries. The value may change during gameplay if
// dynamic resizing is enabled and the window is resized.
func (m *Manager) GetWidth() int {
	return m.width
}

// GetHeight returns the current logical screen height.
// This is used throughout the game for positioning elements and
// determining boundaries. The value may change during gameplay if
// dynamic resizing is enabled and the window is resized.
func (m *Manager) GetHeight() int {
	return m.height
}

// SetDimensions updates the screen dimensions.
// This method is typically called when the window is resized or when
// the device orientation changes. It updates the internal dimensions
// and returns true if the dimensions actually changed.
//
// The return value can be used to determine if dependent systems
// (like rendering targets or UI layouts) need to be updated.
func (m *Manager) SetDimensions(width, height int) bool {
	if m.width != width || m.height != height {
		m.width = width
		m.height = height

		// Only log if debug logging is enabled
		if constants.DebugLogging {
			log.Printf("Screen dimensions set to: %dx%d", width, height)
		}

		return true // Dimensions changed
	}
	return false // No change
}

// IsDynamicResizingEnabled returns whether dynamic resizing is enabled.
// When dynamic resizing is enabled, the game will adapt to fill the available
// screen space. When disabled, the game will maintain a fixed logical size
// regardless of the physical screen dimensions.
//
// This setting affects how the game appears when the window is resized or
// when running on devices with different screen sizes.
func (m *Manager) IsDynamicResizingEnabled() bool {
	return m.dynamicResizing
}

// SetDynamicResizing enables or disables dynamic screen resizing.
// This method allows switching between adaptive and fixed-size modes:
// - When enabled: The game content scales to fill the available space
// - When disabled: The game maintains a fixed logical resolution
//
// Different game scenes may require different resizing behaviors.
// For example, UI scenes might work better with dynamic resizing,
// while gameplay scenes might need fixed dimensions for consistent mechanics.
func (m *Manager) SetDynamicResizing(enabled bool) {
	m.dynamicResizing = enabled
}

// CalculateLayout determines the logical screen size based on the actual window size.
// This method is called by the game's Layout method (part of ebiten.Game interface)
// to determine the logical resolution for rendering.
//
// The behavior depends on the dynamic resizing setting:
// - When enabled: Returns the actual window dimensions, allowing the game to
//   fill the available space while maintaining aspect ratio
// - When disabled: Returns the fixed dimensions from constants, resulting in
//   letterboxing or pillarboxing on screens with different aspect ratios
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

// GetHalfWidth returns half of the screen width, useful for UI layout.
// This is a convenience method commonly used for:
// - Centering elements horizontally
// - Dividing the screen into left/right sections
// - Positioning elements relative to the center
func (m *Manager) GetHalfWidth() int {
	return m.width / 2
}

// GetHalfHeight returns half of the screen height, useful for UI layout.
// This is a convenience method commonly used for:
// - Centering elements vertically
// - Dividing the screen into top/bottom sections
// - Positioning elements relative to the center
func (m *Manager) GetHalfHeight() int {
	return m.height / 2
}

// DefaultManager is a singleton instance of the screen manager.
// This provides a convenient global access point to screen dimensions
// for components that don't have direct access to the game instance.
// Using the singleton pattern simplifies screen dimension access throughout the codebase.
var DefaultManager = New()

// GetWidth provides easy access to the default screen manager's width.
// This is a convenience function for accessing the screen width without
// needing a reference to a Manager instance. It's particularly useful
// for utility functions and components that need screen dimensions but
// aren't directly connected to the main game loop.
func GetWidth() int {
	return DefaultManager.GetWidth()
}

// GetHeight provides easy access to the default screen manager's height.
// This is a convenience function for accessing the screen height without
// needing a reference to a Manager instance. It's particularly useful
// for utility functions and components that need screen dimensions but
// aren't directly connected to the main game loop.
func GetHeight() int {
	return DefaultManager.GetHeight()
}

// SetDimensions provides easy access to the default screen manager's SetDimensions method.
// This is a convenience function for updating screen dimensions without
// needing a reference to a Manager instance. It returns true if the
// dimensions were actually changed.
func SetDimensions(width, height int) bool {
	return DefaultManager.SetDimensions(width, height)
}
