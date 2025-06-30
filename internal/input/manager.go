package input

import (
	"discoveryx/internal/screen"
)

// Manager provides centralized access to all input handlers
type Manager struct {
	keyboard KeyboardHandler
	touch    TouchHandler
	screenManager *screen.Manager
}

// NewManager creates a new input manager with default handlers
func NewManager() *Manager {
	return &Manager{
		keyboard:      NewKeyboardHandler(),
		touch:         NewTouchHandler(),
		screenManager: screen.New(),
	}
}

// Keyboard returns the keyboard handler
func (m *Manager) Keyboard() KeyboardHandler {
	return m.keyboard
}

// Touch returns the touch handler
func (m *Manager) Touch() TouchHandler {
	return m.touch
}

// SetKeyboardHandler allows setting a custom keyboard handler (useful for testing)
func (m *Manager) SetKeyboardHandler(handler KeyboardHandler) {
	m.keyboard = handler
}

// SetTouchHandler allows setting a custom touch handler (useful for testing)
func (m *Manager) SetTouchHandler(handler TouchHandler) {
	m.touch = handler
}

// SetScreenDimensions sets the screen dimensions
func (m *Manager) SetScreenDimensions(width, height int) {
	// Update screen manager dimensions
	if m.screenManager.SetDimensions(width, height) {
		// If dimensions changed, update touch handler
		m.touch.SetScreenDimensions(width, height)

		// Logging is handled by the screen manager
	}
}

// Update updates all input handlers
func (m *Manager) Update() {
	// Update touch handler
	m.touch.Update()
}

// DefaultManager is a singleton instance of the input manager
var DefaultManager = NewManager()

// GetKeyboard provides easy access to the default keyboard handler
func GetKeyboard() KeyboardHandler {
	return DefaultManager.Keyboard()
}

// GetTouch provides easy access to the default touch handler
func GetTouch() TouchHandler {
	return DefaultManager.Touch()
}

// UpdateInput updates all input handlers in the default manager
func UpdateInput() {
	DefaultManager.Update()
}

// SetScreenDimensions sets the screen dimensions for the default manager
func SetScreenDimensions(width, height int) {
	DefaultManager.SetScreenDimensions(width, height)
}
