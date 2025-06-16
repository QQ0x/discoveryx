package input

// Manager provides centralized access to all input handlers
type Manager struct {
	keyboard KeyboardHandler
}

// NewManager creates a new input manager with default handlers
func NewManager() *Manager {
	return &Manager{
		keyboard: NewKeyboardHandler(),
	}
}

// Keyboard returns the keyboard handler
func (m *Manager) Keyboard() KeyboardHandler {
	return m.keyboard
}

// SetKeyboardHandler allows setting a custom keyboard handler (useful for testing)
func (m *Manager) SetKeyboardHandler(handler KeyboardHandler) {
	m.keyboard = handler
}

// DefaultManager is a singleton instance of the input manager
var DefaultManager = NewManager()

// GetKeyboard provides easy access to the default keyboard handler
func GetKeyboard() KeyboardHandler {
	return DefaultManager.Keyboard()
}