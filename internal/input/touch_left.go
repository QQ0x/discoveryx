package input

import "log"

// LeftTouchHandler processes touches on the left half of the screen.
type LeftTouchHandler struct {
	source     TouchSource
	controller MovementController
	width      int
	active     map[ebiten.TouchID]*touchState
}

type touchState struct {
	startX int
	startY int
	lastX  int
	lastY  int
	dir    SwipeDirection
	moving bool
}

// NewLeftTouchHandler creates a new handler for the left screen.
func NewLeftTouchHandler(src TouchSource, ctrl MovementController) *LeftTouchHandler {
	if src == nil {
		src = EbitenTouchSource{}
	}
	return &LeftTouchHandler{
		source:     src,
		controller: ctrl,
		active:     map[ebiten.TouchID]*touchState{},
	}
}

// Update processes current touches and triggers movement events.
func (h *LeftTouchHandler) Update(width, height int) {
	h.width = width
	touches := h.source.Touches()
	seen := map[ebiten.TouchID]bool{}
	for _, t := range touches {
		if t.X >= width/2 {
			continue
		}
		seen[t.ID] = true
		state, ok := h.active[t.ID]
		if !ok {
			h.active[t.ID] = &touchState{startX: t.X, startY: t.Y, lastX: t.X, lastY: t.Y}
			continue
		}
		dx := t.X - state.startX
		dy := t.Y - state.startY
		if !state.moving {
			const threshold = 30
			if dx*dx+dy*dy > threshold*threshold {
				state.dir = calcDirection(dx, dy)
				state.moving = true
				log.Printf("swipe detected: %v", state.dir)
				if h.controller != nil {
					h.controller.StartMoving(state.dir)
				}
			}
		} else if h.controller != nil {
			h.controller.StartMoving(state.dir)
		}
		state.lastX = t.X
		state.lastY = t.Y
	}
	// handle ended touches
	for id, st := range h.active {
		if !seen[id] {
			if st.moving && h.controller != nil {
				h.controller.StopMoving()
			}
			delete(h.active, id)
		}
	}
}

func calcDirection(dx, dy int) SwipeDirection {
	if abs(dx) > abs(dy) {
		if dx < 0 {
			return SwipeLeft
		}
		return SwipeRight
	}
	if dy < 0 {
		return SwipeUp
	}
	return SwipeDown
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
