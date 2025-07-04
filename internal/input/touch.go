package input

import (
	"discoveryx/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"log"
	"math"
	"time"
)

// Direction represents a swipe direction
type Direction int

const (
	DirectionNone Direction = iota
	DirectionUp
	DirectionDown
	DirectionLeft
	DirectionRight
)

// SwipeInfo contains information about a swipe
type SwipeInfo struct {
	// Angle is the angle of the swipe in radians (0 is right, PI/2 is down, PI is left, 3PI/2 is up)
	Angle float64
	// Direction is the cardinal direction of the swipe
	Direction Direction
	// Distance is the distance of the swipe in pixels
	Distance float64
	// Speed is the speed of the swipe in pixels per second
	Speed float64
}

// TouchHandler provides an abstraction for touch input
type TouchHandler interface {
	// IsSwipeDetected returns true if a swipe in the given direction is detected on the left half of the screen
	IsSwipeDetected(direction Direction) bool

	// IsHolding returns true if the user is holding after a swipe on the left half
	IsHolding() bool

	// GetSwipeInfo returns information about the current swipe
	GetSwipeInfo() SwipeInfo

	// Update updates the touch handler state
	Update()

	// SetScreenDimensions sets the screen dimensions
	SetScreenDimensions(width, height int)

	// Right-half joystick methods for shooting
	IsFireJustSwiped() bool
	IsFireHolding() bool
	GetFireJoystickPosition() (float64, float64)
}

// DefaultTouchHandler is the default implementation of TouchHandler
type DefaultTouchHandler struct {
	screenWidth        int
	screenHeight       int
	touchIDs           []ebiten.TouchID
	initialTouchPos    map[ebiten.TouchID]struct{ x, y float64 }
	currentTouchPos    map[ebiten.TouchID]struct{ x, y float64 }
	lastSignificantPos map[ebiten.TouchID]struct{ x, y float64 }
	touchStartTime     map[ebiten.TouchID]time.Time
	detectedSwipes     map[Direction]bool
	holdingDirection   map[Direction]bool
	swipeThreshold     float64
	swipeDuration      time.Duration
	currentSwipeAngle  float64
	isHolding          bool
	swipeDistance      float64
	currentSwipeSpeed  float64
	activeTouchID      ebiten.TouchID
	lastDirection      map[ebiten.TouchID]Direction

	// Right-side joystick for firing
	fireTouchID     ebiten.TouchID
	fireInitialPos  struct{ x, y float64 }
	fireCurrentPos  struct{ x, y float64 }
	fireStartTime   time.Time
	fireHolding     bool
	fireJustSwiped  bool
	fireJoystickPos struct{ x, y float64 }
	fireDefaultPos  struct{ x, y float64 }
}

// NewTouchHandler creates a new default touch handler
func NewTouchHandler() TouchHandler {
	h := &DefaultTouchHandler{
		initialTouchPos:    make(map[ebiten.TouchID]struct{ x, y float64 }),
		currentTouchPos:    make(map[ebiten.TouchID]struct{ x, y float64 }),
		lastSignificantPos: make(map[ebiten.TouchID]struct{ x, y float64 }),
		touchStartTime:     make(map[ebiten.TouchID]time.Time),
		detectedSwipes:     make(map[Direction]bool),
		holdingDirection:   make(map[Direction]bool),
		lastDirection:      make(map[ebiten.TouchID]Direction),
		swipeThreshold:     constants.SwipeThreshold, // Minimum distance for swipe detection
		swipeDuration:      constants.SwipeDuration,  // Maximum duration for swipe detection
		currentSwipeAngle:  0,
		isHolding:          false,
		swipeDistance:      0,
		currentSwipeSpeed:  0,
		activeTouchID:      0,
	}

	h.fireDefaultPos = struct{ x, y float64 }{constants.ScreenWidth - 50, constants.ScreenHeight - 50}
	h.fireJoystickPos = h.fireDefaultPos
	return h
}

// IsSwipeDetected checks if a swipe in the given direction is detected
func (h *DefaultTouchHandler) IsSwipeDetected(direction Direction) bool {
	return h.detectedSwipes[direction]
}

// IsHolding checks if the user is holding after a swipe
func (h *DefaultTouchHandler) IsHolding() bool {
	return h.isHolding
}

// GetSwipeInfo returns information about the current swipe
func (h *DefaultTouchHandler) GetSwipeInfo() SwipeInfo {
	// Find the active direction
	var activeDirection Direction
	for dir, active := range h.holdingDirection {
		if active {
			activeDirection = dir
			break
		}
	}

	return SwipeInfo{
		Angle:     h.currentSwipeAngle,
		Direction: activeDirection,
		Distance:  h.swipeDistance,
		Speed:     h.currentSwipeSpeed,
	}
}

func (h *DefaultTouchHandler) IsFireJustSwiped() bool {
	return h.fireJustSwiped
}

func (h *DefaultTouchHandler) IsFireHolding() bool {
	return h.fireHolding
}

func (h *DefaultTouchHandler) GetFireJoystickPosition() (float64, float64) {
	return h.fireJoystickPos.x, h.fireJoystickPos.y
}

// SetScreenDimensions sets the screen dimensions
func (h *DefaultTouchHandler) SetScreenDimensions(width, height int) {
	h.screenWidth = width
	h.screenHeight = height
	h.fireDefaultPos = struct{ x, y float64 }{float64(width) - 50, float64(height) - 50}
	if h.fireTouchID == 0 {
		h.fireJoystickPos = h.fireDefaultPos
	}
}

// Update updates the touch handler state
func (h *DefaultTouchHandler) Update() {
	for dir := range h.detectedSwipes {
		h.detectedSwipes[dir] = false
	}
	h.fireJustSwiped = false

	halfWidth := h.screenWidth / 2

	h.touchIDs = inpututil.AppendJustPressedTouchIDs(h.touchIDs[:0])

	for _, id := range h.touchIDs {
		x, y := ebiten.TouchPosition(id)
		if x >= halfWidth {
			if h.fireTouchID == 0 {
				h.fireTouchID = id
				h.fireInitialPos = struct{ x, y float64 }{float64(x), float64(y)}
				h.fireCurrentPos = h.fireInitialPos
				h.fireStartTime = time.Now()
				h.fireJoystickPos = h.fireInitialPos
				h.fireHolding = false
			}

			if constants.DebugLogging {
				log.Printf("Touch event detected on right half: ID=%d, Position=(%d, %d)", id, x, y)
			}
			continue
		}

		h.initialTouchPos[id] = struct{ x, y float64 }{float64(x), float64(y)}
		h.currentTouchPos[id] = struct{ x, y float64 }{float64(x), float64(y)}
		h.lastSignificantPos[id] = struct{ x, y float64 }{float64(x), float64(y)}
		h.touchStartTime[id] = time.Now()
		h.lastDirection[id] = DirectionNone
	}

	for id := range h.initialTouchPos {
		if inpututil.IsTouchJustReleased(id) {
			delete(h.initialTouchPos, id)
			delete(h.currentTouchPos, id)
			delete(h.lastSignificantPos, id)
			delete(h.touchStartTime, id)
			delete(h.lastDirection, id)

			for dir := range h.holdingDirection {
				h.holdingDirection[dir] = false
			}

			h.isHolding = false

			if id == h.activeTouchID {
				h.activeTouchID = 0
				h.currentSwipeAngle = 0
				h.swipeDistance = 0
				h.currentSwipeSpeed = 0
			}

			continue
		}

		x, y := ebiten.TouchPosition(id)
		h.currentTouchPos[id] = struct{ x, y float64 }{float64(x), float64(y)}

		currentX := h.currentTouchPos[id].x
		currentY := h.currentTouchPos[id].y

		referenceX := h.lastSignificantPos[id].x
		referenceY := h.lastSignificantPos[id].y

		dx := currentX - referenceX
		dy := currentY - referenceY
		distance := math.Sqrt(dx*dx + dy*dy)

		angle := math.Atan2(dy, dx)

		elapsed := time.Since(h.touchStartTime[id])
		speed := distance / elapsed.Seconds()

		prevSwipeInfo := h.GetSwipeInfo()

		if distance >= h.swipeThreshold {
			h.activeTouchID = id
			h.swipeDistance = distance
			h.currentSwipeAngle = angle
			h.currentSwipeSpeed = speed
			h.isHolding = true

			degrees := math.Mod(angle*180/math.Pi+360, 360)
			var direction Direction
			switch {
			case degrees >= 315 || degrees < 45:
				direction = DirectionRight
			case degrees >= 45 && degrees < 135:
				direction = DirectionDown
			case degrees >= 135 && degrees < 225:
				direction = DirectionLeft
			default:
				direction = DirectionUp
			}

			if elapsed <= h.swipeDuration {
				h.detectedSwipes[direction] = true
			}

			h.holdingDirection[direction] = true

			if prevSwipeInfo.Angle != 0 {
				needsUpdate := distance > 60
				if !needsUpdate {
					angleDiff := math.Abs(angle - prevSwipeInfo.Angle)
					needsUpdate = angleDiff > 0.2
				}
				if needsUpdate {
					h.lastSignificantPos[id] = h.currentTouchPos[id]
				}
			}

			h.lastDirection[id] = direction
		}
	}

	if h.fireTouchID != 0 {
		if inpututil.IsTouchJustReleased(h.fireTouchID) {
			h.fireTouchID = 0
			h.fireHolding = false
			h.fireJoystickPos = h.fireDefaultPos
		} else {
			x, y := ebiten.TouchPosition(h.fireTouchID)
			h.fireCurrentPos = struct{ x, y float64 }{float64(x), float64(y)}
			dx := h.fireCurrentPos.x - h.fireInitialPos.x
			dy := h.fireCurrentPos.y - h.fireInitialPos.y
			distance := math.Sqrt(dx*dx + dy*dy)
			if distance >= h.swipeThreshold && !h.fireHolding {
				if time.Since(h.fireStartTime) <= h.swipeDuration {
					h.fireJustSwiped = true
				}
				h.fireHolding = true
			}
		}
	}
}
