package input

import "testing"

// mockMovement records StartMoving and StopMoving calls.
type mockMovement struct {
	started []SwipeDirection
	stopped int
}

func (m *mockMovement) StartMoving(dir SwipeDirection) {
	m.started = append(m.started, dir)
}

func (m *mockMovement) StopMoving() { m.stopped++ }

// fakeSource allows predefined touch sequences.
type fakeSource struct{ touches []TouchData }

func (f fakeSource) Touches() []TouchData { return f.touches }

func TestLeftSwipe(t *testing.T) {
	move := &mockMovement{}
	src := &fakeSource{}
	handler := NewLeftTouchHandler(src, move)

	// start touch at (10,10)
	src.touches = []TouchData{{ID: 1, X: 10, Y: 10}}
	handler.Update(100, 100)

	// move right beyond threshold
	src.touches = []TouchData{{ID: 1, X: 50, Y: 10}}
	handler.Update(100, 100)

	if len(move.started) == 0 || move.started[0] != SwipeRight {
		t.Fatalf("expected swipe right, got %v", move.started)
	}

	// finger lifted
	src.touches = nil
	handler.Update(100, 100)
	if move.stopped != 1 {
		t.Fatalf("expected stop on touch end")
	}
}
