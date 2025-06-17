package input

import "log"

// RightTouchHandler captures events on the right half for future use.
type RightTouchHandler struct {
	source TouchSource
	width  int
}

// NewRightTouchHandler creates a new passive handler.
func NewRightTouchHandler(src TouchSource) *RightTouchHandler {
	if src == nil {
		src = EbitenTouchSource{}
	}
	return &RightTouchHandler{source: src}
}

// Update logs all touches on the right half of the screen.
func (h *RightTouchHandler) Update(width, height int) {
	h.width = width
	for _, t := range h.source.Touches() {
		if t.X >= width/2 {
			log.Printf("right touch id=%d x=%d y=%d", t.ID, t.X, t.Y)
		}
	}
}
