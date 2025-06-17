package input

import "github.com/hajimehoshi/ebiten/v2"

// SwipeDirection represents a directional swipe on screen.
type SwipeDirection int

const (
	SwipeNone SwipeDirection = iota
	SwipeLeft
	SwipeRight
	SwipeUp
	SwipeDown
)

// MovementController handles movement events triggered by touch input.
type MovementController interface {
	// StartMoving begins movement in a direction.
	StartMoving(dir SwipeDirection)
	// StopMoving stops any ongoing movement.
	StopMoving()
}

// TouchData represents the coordinates for a touch event.
type TouchData struct {
	ID ebiten.TouchID
	X  int
	Y  int
}

// TouchSource defines an object that can provide current touch data.
type TouchSource interface {
	Touches() []TouchData
}

// EbitenTouchSource implements TouchSource using ebiten's input API.
type EbitenTouchSource struct{}

// Touches returns all current touches from ebiten.
func (EbitenTouchSource) Touches() []TouchData {
	ids := ebiten.AppendTouchIDs(nil)
	touches := make([]TouchData, 0, len(ids))
	for _, id := range ids {
		x, y := ebiten.TouchPosition(id)
		touches = append(touches, TouchData{ID: id, X: x, Y: y})
	}
	return touches
}
