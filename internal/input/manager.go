// Package input provides a unified system for handling user input across different platforms.
// It supports multiple input methods including keyboard, touch, and gamepad,
// with a flexible architecture that allows for easy addition of new input types.
//
// The input system is designed to:
// - Abstract platform-specific input details
// - Provide consistent access to input state
// - Support different control schemes
// - Handle input coordinate mapping for different screen sizes and resolutions
// - Enable gesture recognition for touch input
package input

import (
	"discoveryx/internal/screen"
)

// Manager provides centralized access to all input handlers.
// It coordinates between different input methods and ensures they work together
// consistently across different platforms (desktop, mobile, etc.).
//
// The Manager maintains references to individual input handlers and provides
// methods to access them. It also handles screen dimension changes to ensure
// input coordinates are correctly mapped to game coordinates.
type Manager struct {
	keyboard      KeyboardHandler   // Handles keyboard input (primarily for desktop)
	touch         TouchHandler      // Handles touch input (primarily for mobile)
	screenManager *screen.Manager   // Manages screen dimensions for input coordinate mapping
}

// NewManager creates a new input manager with default handlers.
// This initializes all input subsystems with their default implementations,
// creating a fully functional input system ready for use.
// The default handlers are suitable for most game scenarios, but can be
// replaced with custom implementations if needed.
func NewManager() *Manager {
	return &Manager{
		keyboard:      NewKeyboardHandler(), // Initialize keyboard input
		touch:         NewTouchHandler(),    // Initialize touch input
		screenManager: screen.New(),         // Initialize screen manager for coordinate mapping
	}
}

// Keyboard returns the keyboard handler.
// This provides access to keyboard state, including key presses, releases,
// and held keys. It's primarily used on desktop platforms but can also
// support external keyboards on mobile devices.
func (m *Manager) Keyboard() KeyboardHandler {
	return m.keyboard
}

// Touch returns the touch handler.
// This provides access to touch state, including touch positions, gestures,
// and multi-touch support. It's primarily used on mobile platforms but can
// also support touchscreens on desktop devices.
func (m *Manager) Touch() TouchHandler {
	return m.touch
}

// SetKeyboardHandler allows setting a custom keyboard handler.
// This is particularly useful for:
// - Testing with mock input
// - Implementing custom keyboard control schemes
// - Supporting specialized keyboard hardware
func (m *Manager) SetKeyboardHandler(handler KeyboardHandler) {
	m.keyboard = handler
}

// SetTouchHandler allows setting a custom touch handler.
// This is particularly useful for:
// - Testing with mock input
// - Implementing custom touch control schemes
// - Supporting specialized touch hardware or gestures
func (m *Manager) SetTouchHandler(handler TouchHandler) {
	m.touch = handler
}

// SetScreenDimensions updates screen dimensions and propagates changes to handlers.
// This is called whenever the game window is resized or when the device
// orientation changes. It ensures that input coordinates (especially touch)
// are correctly mapped to game coordinates regardless of screen size or resolution.
//
// The method returns immediately if the dimensions haven't changed significantly,
// avoiding unnecessary updates to the input handlers.
func (m *Manager) SetScreenDimensions(width, height int) {
	// Only update handlers if dimensions have actually changed
	if m.screenManager.SetDimensions(width, height) {
		// Update touch handler with new dimensions for proper coordinate mapping
		m.touch.SetScreenDimensions(width, height)
	}
}

// Update processes all input handlers for the current frame.
// This method should be called once per frame, typically at the beginning
// of the game's Update method. It ensures that all input state is fresh
// and consistent for the current frame.
//
// The update process includes:
// - Polling for new input events
// - Updating internal state of each handler
// - Processing gestures for touch input
func (m *Manager) Update() {
	// Update touch input state
	m.touch.Update()

	// Note: Keyboard doesn't need explicit updates as Ebiten handles keyboard state automatically
}

// DefaultManager is a singleton instance of the input manager.
// This provides a convenient global access point to the input system
// for components that don't have direct access to the game instance.
// Using the singleton pattern simplifies input access throughout the codebase.
var DefaultManager = NewManager()

// GetKeyboard returns the keyboard handler from the default manager.
// This is a convenience function for accessing keyboard input without
// needing a reference to the input manager. It's particularly useful
// for utility functions and components that need input but aren't
// directly connected to the main game loop.
func GetKeyboard() KeyboardHandler {
	return DefaultManager.Keyboard()
}

// GetTouch returns the touch handler from the default manager.
// This is a convenience function for accessing touch input without
// needing a reference to the input manager. It's particularly useful
// for UI components that need to handle touch events independently.
func GetTouch() TouchHandler {
	return DefaultManager.Touch()
}

// UpdateInput processes all input handlers in the default manager for the current frame.
// This is a convenience function that updates all input state in the default manager.
// It should be called once per frame, typically at the beginning of the game's update cycle.
//
// Note: The main game loop typically uses its own input manager instance rather than
// the default manager, but this function is useful for components that use the
// default manager.
func UpdateInput() {
	DefaultManager.Update()
}

// SetScreenDimensions updates screen dimensions in the default manager.
// This is a convenience function for updating screen dimensions without
// needing a reference to the input manager. It should be called whenever
// the game window is resized or when the device orientation changes.
//
// This ensures that input coordinates (especially touch) are correctly mapped
// to game coordinates for components using the default manager.
func SetScreenDimensions(width, height int) {
	DefaultManager.SetScreenDimensions(width, height)
}
