package input

// TouchHandler defines behaviour for processing touch input each frame.
type TouchHandler interface {
	Update(width, height int)
}

// DefaultTouchHandler coordinates left and right touch areas.
type DefaultTouchHandler struct {
	left  *LeftTouchHandler
	right *RightTouchHandler
}

// NewDefaultTouchHandler creates a touch handler with optional dependencies.
func NewDefaultTouchHandler(ctrl MovementController) *DefaultTouchHandler {
	src := EbitenTouchSource{}
	return &DefaultTouchHandler{
		left:  NewLeftTouchHandler(src, ctrl),
		right: NewRightTouchHandler(src),
	}
}

// Update processes both screen halves.
func (h *DefaultTouchHandler) Update(width, height int) {
	h.left.Update(width, height)
	h.right.Update(width, height)
}
