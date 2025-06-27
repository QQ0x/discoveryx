package input

import (
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
}

// NewTouchHandler creates a new default touch handler
func NewTouchHandler() TouchHandler {
	return &DefaultTouchHandler{
		initialTouchPos:    make(map[ebiten.TouchID]struct{ x, y float64 }),
		currentTouchPos:    make(map[ebiten.TouchID]struct{ x, y float64 }),
		lastSignificantPos: make(map[ebiten.TouchID]struct{ x, y float64 }),
		touchStartTime:     make(map[ebiten.TouchID]time.Time),
		detectedSwipes:     make(map[Direction]bool),
		holdingDirection:   make(map[Direction]bool),
		lastDirection:      make(map[ebiten.TouchID]Direction),
		swipeThreshold:     30.0,                   // Minimum distance for swipe detection
		swipeDuration:      200 * time.Millisecond, // Maximum duration for swipe detection
		currentSwipeAngle:  0,
		isHolding:          false,
		swipeDistance:      0,
		currentSwipeSpeed:  0,
		activeTouchID:      0,
	}
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

// SetScreenDimensions sets the screen dimensions
func (h *DefaultTouchHandler) SetScreenDimensions(width, height int) {
	h.screenWidth = width
	h.screenHeight = height
	log.Printf("Touch handler screen dimensions set to: %dx%d", width, height)
}

// Update updates the touch handler state
func (h *DefaultTouchHandler) Update() {
	// Reset swipe detection for this frame
	for dir := range h.detectedSwipes {
		h.detectedSwipes[dir] = false
	}

	// Use the stored screen dimensions
	halfWidth := h.screenWidth / 2

	// Get all active touch IDs
	h.touchIDs = inpututil.AppendJustPressedTouchIDs(h.touchIDs[:0])

	// Process new touches
	for _, id := range h.touchIDs {
		x, y := ebiten.TouchPosition(id)
		h.initialTouchPos[id] = struct{ x, y float64 }{float64(x), float64(y)}
		h.currentTouchPos[id] = struct{ x, y float64 }{float64(x), float64(y)}
		h.lastSignificantPos[id] = struct{ x, y float64 }{float64(x), float64(y)}
		h.touchStartTime[id] = time.Now()
		h.lastDirection[id] = DirectionNone

		// Log touch events for right half of screen
		if x >= halfWidth {
			log.Printf("Touch event detected on right half: ID=%d, Position=(%d, %d)", id, x, y)
		}
	}

	// Process ongoing touches
	for id := range h.initialTouchPos {
		if inpututil.IsTouchJustReleased(id) {
			// Clean up when touch is released
			delete(h.initialTouchPos, id)
			delete(h.currentTouchPos, id)
			delete(h.lastSignificantPos, id)
			delete(h.touchStartTime, id)
			delete(h.lastDirection, id)

			// Reset holding state when touch is released
			for dir := range h.holdingDirection {
				h.holdingDirection[dir] = false
			}

			// Reset the holding flag
			h.isHolding = false

			// Reset the active touch ID if this is the active touch
			if id == h.activeTouchID {
				h.activeTouchID = 0
				h.currentSwipeAngle = 0
				h.swipeDistance = 0
				h.currentSwipeSpeed = 0
			}

			continue
		}

		// Update current position
		x, y := ebiten.TouchPosition(id)
		h.currentTouchPos[id] = struct{ x, y float64 }{float64(x), float64(y)}

		// Skip processing for touches on the right half
		if h.initialTouchPos[id].x >= float64(halfWidth) {
			// Just log the movement for right half
			log.Printf("Touch movement on right half: ID=%d, Position=(%d, %d)", id, x, y)
			continue
		}

		// Process touches on the left half
		currentX := h.currentTouchPos[id].x
		currentY := h.currentTouchPos[id].y

		// Use lastSignificantPos for distance and angle calculations
		referenceX := h.lastSignificantPos[id].x
		referenceY := h.lastSignificantPos[id].y

		// Calculate distance and direction
		dx := currentX - referenceX
		dy := currentY - referenceY
		distance := math.Sqrt(dx*dx + dy*dy)

		// Calculate angle (in radians)
		angle := math.Atan2(dy, dx)

		// Calculate speed (distance / time)
		elapsed := time.Since(h.touchStartTime[id])
		speed := distance / elapsed.Seconds()

		// Store the previous swipe info for comparison
		prevSwipeInfo := h.GetSwipeInfo()

		// Detect swipe
		if distance >= h.swipeThreshold {
			// Store the active touch ID
			h.activeTouchID = id

			// Store the swipe distance
			h.swipeDistance = distance

			// Store the swipe angle
			h.currentSwipeAngle = angle

			// Store the swipe speed
			h.currentSwipeSpeed = speed

			// Set the holding flag
			h.isHolding = true

			// Determine cardinal direction for backward compatibility
			var direction Direction
			// Convert angle to degrees and normalize to 0-360
			degrees := math.Mod(angle*180/math.Pi+360, 360)

			if degrees >= 315 || degrees < 45 {
				direction = DirectionRight
			} else if degrees >= 45 && degrees < 135 {
				direction = DirectionDown
			} else if degrees >= 135 && degrees < 225 {
				direction = DirectionLeft
			} else { // degrees >= 225 && degrees < 315
				direction = DirectionUp
			}

			// Check if this is a new swipe or continuing hold
			if elapsed <= h.swipeDuration {
				// This is a new swipe
				h.detectedSwipes[direction] = true
				log.Printf("Swipe detected: %v", direction)
			}

			// Set holding state for backward compatibility
			h.holdingDirection[direction] = true
			log.Printf("Holding direction: %v (angle=%f)", direction, angle)

			// Check if the swipe direction has changed
			if prevSwipeInfo.Angle != 0 {
				// Calculate the angle difference
				angleDiff := math.Abs(angle - prevSwipeInfo.Angle)

				// Update reference point to maintain connection between finger movement and in-game response,
				// but with optimized thresholds to improve performance:
				// 1. Significant direction change (> 0.2 radians)
				// 2. Substantial movement (> 60 pixels) regardless of angle change
				if angleDiff > 0.2 || distance > 60 {
					log.Printf("Swipe direction changed: %f -> %f (diff: %f, distance: %f)",
						prevSwipeInfo.Angle, angle, angleDiff, distance)

					// Update the lastSignificantPos to the current position
					h.lastSignificantPos[id] = h.currentTouchPos[id]
					log.Printf("Updated lastSignificantPos: (%f, %f)",
						h.currentTouchPos[id].x, h.currentTouchPos[id].y)
				}
			}

			// Update the last direction
			h.lastDirection[id] = direction

			// Check if the swipe is continuing in the same direction (increasing distance)
			if prevSwipeInfo.Distance > 0 && distance > prevSwipeInfo.Distance {
				log.Printf("Swipe continuing: distance increased from %f to %f", prevSwipeInfo.Distance, distance)
			}

			// Check if the swipe is returning toward the starting point (decreasing distance)
			if prevSwipeInfo.Distance > 0 && distance < prevSwipeInfo.Distance {
				log.Printf("Swipe returning: distance decreased from %f to %f", prevSwipeInfo.Distance, distance)
			}
		}
	}
}
