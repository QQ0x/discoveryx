package input

// Manager provides centralized access to all input handlers
type Manager struct {
	keyboard KeyboardHandler
	touch    TouchHandler
}

// NewManager creates a new input manager with default handlers
func NewManager() *Manager {
	return &Manager{
		keyboard: NewKeyboardHandler(),
		touch:    NewDefaultTouchHandler(nil),
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
