package input

import "log"

// Manager provides centralized access to all input handlers
type Manager struct {
	keyboard KeyboardHandler
	touch    TouchHandler
	width    int
	height   int
}

// NewManager creates a new input manager with default handlers
func NewManager() *Manager {
	return &Manager{
		keyboard: NewKeyboardHandler(),
		touch:    NewTouchHandler(),
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
	m.width = width
	m.height = height
	m.touch.SetScreenDimensions(width, height)
	log.Printf("Input manager screen dimensions set to: %dx%d", width, height)
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
