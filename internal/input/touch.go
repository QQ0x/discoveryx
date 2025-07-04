package input

import (
	"discoveryx/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"log"
	"math"
	"time"
)

// Direction represents a swipe direction on a touch screen.
// This enum is used to categorize touch gestures into cardinal directions
// (up, down, left, right) for simplified game control input.
// 
// The game uses these directions to map touch gestures to player actions:
// - Up: Move player up
// - Down: Move player down
// - Left: Move player left
// - Right: Move player right
// - None: No movement or invalid gesture
type Direction int

const (
	DirectionNone  Direction = iota // No direction or invalid gesture
	DirectionUp                     // Upward swipe (player moves up)
	DirectionDown                   // Downward swipe (player moves down)
	DirectionLeft                   // Leftward swipe (player moves left)
	DirectionRight                  // Rightward swipe (player moves right)
)

// SwipeInfo contains detailed information about a touch swipe gesture.
// This struct provides comprehensive data about the swipe, allowing the game
// to respond appropriately to different types of gestures with varying
// intensities and directions.
//
// The information can be used to implement velocity-based movement,
// directional attacks, or other gesture-controlled game mechanics.
type SwipeInfo struct {
	// Angle is the angle of the swipe in radians
	// - 0: Right
	// - PI/2 (1.57): Down
	// - PI (3.14): Left
	// - 3PI/2 (4.71): Up
	// This provides precise directional information beyond just cardinal directions.
	Angle float64

	// Direction is the simplified cardinal direction of the swipe
	// This is derived from the angle and provides an easy way to determine
	// the general direction without dealing with angle calculations.
	Direction Direction

	// Distance is the length of the swipe in pixels
	// This can be used to determine the intensity of the action,
	// such as movement speed or attack power.
	Distance float64

	// Speed is the velocity of the swipe in pixels per second
	// This can be used to implement momentum-based movement or
	// to distinguish between quick flicks and slow drags.
	Speed float64
}

// TouchHandler provides an abstraction for touch input across different platforms.
// This interface defines the contract for handling touch gestures and virtual joysticks,
// allowing the game to respond to touch input in a platform-independent way.
//
// The touch control scheme divides the screen into two halves:
// - Left half: Movement controls using swipe gestures
// - Right half: Firing controls using a virtual joystick
//
// This separation allows for simultaneous movement and firing actions,
// mimicking the dual-stick control scheme common in action games.
type TouchHandler interface {
	// IsSwipeDetected returns true if a swipe in the given direction is detected on the left half of the screen.
	// This is used to detect player movement commands through swipe gestures.
	// The method returns true only for the frame when the swipe is first detected,
	// making it suitable for triggering one-time actions.
	IsSwipeDetected(direction Direction) bool

	// IsHolding returns true if the user is continuing to hold after a swipe on the left half.
	// This can be used to implement continuous movement as long as the player
	// keeps their finger on the screen after swiping.
	IsHolding() bool

	// GetSwipeInfo returns detailed information about the current swipe.
	// This provides access to the angle, direction, distance, and speed of the swipe,
	// allowing for more nuanced control based on how the player swiped.
	GetSwipeInfo() SwipeInfo

	// Update processes touch events and updates the internal state.
	// This should be called once per frame to ensure touch input is properly detected.
	// It handles new touches, ongoing touches, and touch releases for both
	// the movement and firing controls.
	Update()

	// SetScreenDimensions updates the handler with the current screen dimensions.
	// This ensures touch coordinates are correctly mapped regardless of screen size
	// or orientation changes, which is particularly important for mobile devices.
	SetScreenDimensions(width, height int)

	// Right-half joystick methods for shooting control

	// IsFireJustSwiped returns true if a fire action was just initiated on the right half.
	// This is used to detect when the player wants to start firing, and returns true
	// only for the frame when the fire action is first detected.
	IsFireJustSwiped() bool

	// IsFireHolding returns true if the player is continuing to hold after initiating a fire action.
	// This can be used to implement continuous firing as long as the player
	// keeps their finger on the screen after the initial fire action.
	IsFireHolding() bool

	// GetFireJoystickPosition returns the current position of the virtual joystick for firing.
	// This provides the x,y coordinates that can be used to determine the firing direction.
	// The coordinates are relative to the initial touch position, allowing for
	// directional aiming similar to a physical joystick.
	GetFireJoystickPosition() (float64, float64)
}

// DefaultTouchHandler is the default implementation of TouchHandler.
// It provides a complete touch input system with swipe detection and virtual joystick
// functionality, designed specifically for the game's dual-control scheme:
// - Left half of screen: Movement controls via swipe gestures
// - Right half of screen: Firing controls via virtual joystick
//
// The handler tracks multiple simultaneous touches, distinguishes between
// different types of gestures, and provides detailed information about
// swipe direction, distance, and speed.
type DefaultTouchHandler struct {
	// Screen dimensions
	screenWidth        int                                    // Current screen width in pixels
	screenHeight       int                                    // Current screen height in pixels

	// Touch tracking for movement (left half)
	touchIDs           []ebiten.TouchID                       // Reusable slice for getting new touch IDs
	initialTouchPos    map[ebiten.TouchID]struct{ x, y float64 } // Starting position of each touch
	currentTouchPos    map[ebiten.TouchID]struct{ x, y float64 } // Current position of each touch
	lastSignificantPos map[ebiten.TouchID]struct{ x, y float64 } // Last position used for angle calculation
	touchStartTime     map[ebiten.TouchID]time.Time           // When each touch began
	detectedSwipes     map[Direction]bool                     // Whether a swipe was detected in each direction
	holdingDirection   map[Direction]bool                     // Whether the user is holding in each direction
	swipeThreshold     float64                                // Minimum distance to register a swipe
	swipeDuration      time.Duration                          // Maximum time for a valid swipe
	currentSwipeAngle  float64                                // Angle of the current swipe in radians
	isHolding          bool                                   // Whether any touch is being held
	swipeDistance      float64                                // Distance of the current swipe in pixels
	currentSwipeSpeed  float64                                // Speed of the current swipe in pixels/second
	activeTouchID      ebiten.TouchID                         // ID of the touch currently being tracked
	lastDirection      map[ebiten.TouchID]Direction           // Last detected direction for each touch

	// Right-side joystick for firing
	fireTouchID     ebiten.TouchID                         // ID of the touch for firing control
	fireInitialPos  struct{ x, y float64 }                 // Starting position of the fire touch
	fireCurrentPos  struct{ x, y float64 }                 // Current position of the fire touch
	fireStartTime   time.Time                              // When the fire touch began
	fireHolding     bool                                   // Whether the fire touch is being held
	fireJustSwiped  bool                                   // Whether a fire swipe was just detected
	fireJoystickPos struct{ x, y float64 }                 // Current position of the fire joystick
	fireDefaultPos  struct{ x, y float64 }                 // Default position of the fire joystick when inactive
}

// NewTouchHandler creates a new default touch handler.
// This factory function initializes all the necessary data structures and
// configuration values for the touch input system. It returns an implementation
// of the TouchHandler interface ready to use in the game.
//
// The handler is configured with default values from the constants package:
// - SwipeThreshold: Minimum distance in pixels to register a swipe
// - SwipeDuration: Maximum time window for a valid swipe
// - ScreenWidth/Height: Initial screen dimensions
func NewTouchHandler() TouchHandler {
	// Create and initialize the handler with all necessary maps and default values
	h := &DefaultTouchHandler{
		// Initialize maps for tracking multiple simultaneous touches
		initialTouchPos:    make(map[ebiten.TouchID]struct{ x, y float64 }), // Starting positions
		currentTouchPos:    make(map[ebiten.TouchID]struct{ x, y float64 }), // Current positions
		lastSignificantPos: make(map[ebiten.TouchID]struct{ x, y float64 }), // Reference positions for angle calculation
		touchStartTime:     make(map[ebiten.TouchID]time.Time),              // Touch start timestamps

		// Maps for tracking swipe state
		detectedSwipes:     make(map[Direction]bool),                        // Swipe detection flags
		holdingDirection:   make(map[Direction]bool),                        // Holding state for each direction
		lastDirection:      make(map[ebiten.TouchID]Direction),              // Last direction for each touch

		// Configuration values from constants
		swipeThreshold:     constants.SwipeThreshold, // Minimum distance for swipe detection
		swipeDuration:      constants.SwipeDuration,  // Maximum duration for swipe detection

		// Initialize tracking variables
		currentSwipeAngle:  0,                        // No initial swipe
		isHolding:          false,                    // Not holding initially
		swipeDistance:      0,                        // No initial distance
		currentSwipeSpeed:  0,                        // No initial speed
		activeTouchID:      0,                        // No active touch initially
	}

	// Set up the fire joystick position in the bottom-right corner
	// This is where the virtual joystick appears when inactive
	h.fireDefaultPos = struct{ x, y float64 }{
		constants.ScreenWidth - 50,  // 50 pixels from the right edge
		constants.ScreenHeight - 50, // 50 pixels from the bottom edge
	}

	// Initialize the joystick at its default position
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

// IsFireJustSwiped returns true if a fire swipe was just detected on the right half of the screen
func (h *DefaultTouchHandler) IsFireJustSwiped() bool {
	return h.fireJustSwiped
}

// IsFireHolding returns true if the user is holding after a fire swipe on the right half
func (h *DefaultTouchHandler) IsFireHolding() bool {
	return h.fireHolding
}

// GetFireJoystickPosition returns the current position of the fire joystick
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

// Update processes touch input and updates swipe detection state.
// This method is the core of the touch input system and should be called once per frame.
// It handles all aspects of touch input processing:
// 1. Detecting new touches and categorizing them (movement vs. firing)
// 2. Tracking ongoing touches and calculating swipe parameters
// 3. Detecting touch releases and cleaning up
// 4. Updating the virtual joystick position for firing
//
// The screen is divided into two halves:
// - Left half: Movement controls via swipe gestures
// - Right half: Firing controls via virtual joystick
func (h *DefaultTouchHandler) Update() {
	// Reset one-time detection flags at the beginning of each frame
	// This ensures IsSwipeDetected() and IsFireJustSwiped() only return true
	// for a single frame when the action is first detected
	for dir := range h.detectedSwipes {
		h.detectedSwipes[dir] = false
	}
	h.fireJustSwiped = false

	// Calculate the dividing line between left and right control areas
	halfWidth := h.screenWidth / 2

	// ---- STEP 1: Process newly detected touches ----

	// Get all touches that just started this frame
	h.touchIDs = inpututil.AppendJustPressedTouchIDs(h.touchIDs[:0])

	// Process each new touch
	for _, id := range h.touchIDs {
		x, y := ebiten.TouchPosition(id)

		// Determine if this touch is on the right half (firing controls)
		// or left half (movement controls) of the screen
		if x >= halfWidth {
			// ---- RIGHT HALF: FIRING CONTROLS ----

			// Only process if we don't already have an active fire touch
			if h.fireTouchID == 0 {
				// Initialize fire touch tracking
				h.fireTouchID = id
				h.fireInitialPos = struct{ x, y float64 }{float64(x), float64(y)}
				h.fireCurrentPos = h.fireInitialPos
				h.fireStartTime = time.Now()
				h.fireJoystickPos = h.fireInitialPos
				h.fireHolding = false
			}

			// Log touch events if debug logging is enabled
			if constants.DebugLogging {
				log.Printf("Touch event detected on right half: ID=%d, Position=(%d, %d)", id, x, y)
			}
			continue
		}

		// ---- LEFT HALF: MOVEMENT CONTROLS ----

		// Initialize movement touch tracking
		h.initialTouchPos[id] = struct{ x, y float64 }{float64(x), float64(y)}
		h.currentTouchPos[id] = struct{ x, y float64 }{float64(x), float64(y)}
		h.lastSignificantPos[id] = struct{ x, y float64 }{float64(x), float64(y)}
		h.touchStartTime[id] = time.Now()
		h.lastDirection[id] = DirectionNone
	}

	// ---- STEP 2: Process ongoing movement touches (left half) ----
	for id := range h.initialTouchPos {
		// Check if this touch has been released
		if inpututil.IsTouchJustReleased(id) {
			// ---- TOUCH RELEASE HANDLING ----

			// Clean up all data associated with this touch
			delete(h.initialTouchPos, id)
			delete(h.currentTouchPos, id)
			delete(h.lastSignificantPos, id)
			delete(h.touchStartTime, id)
			delete(h.lastDirection, id)

			// Reset all holding states since the touch is released
			for dir := range h.holdingDirection {
				h.holdingDirection[dir] = false
			}
			h.isHolding = false

			// If this was the active touch being tracked for swipe info,
			// reset all the swipe tracking data
			if id == h.activeTouchID {
				h.activeTouchID = 0
				h.currentSwipeAngle = 0
				h.swipeDistance = 0
				h.currentSwipeSpeed = 0
			}
			continue
		}

		// ---- ONGOING TOUCH PROCESSING ----

		// Update the current position of this touch
		x, y := ebiten.TouchPosition(id)
		h.currentTouchPos[id] = struct{ x, y float64 }{float64(x), float64(y)}

		// Calculate movement vector from the last significant position
		// This is used to determine swipe direction and distance
		currentX := h.currentTouchPos[id].x
		currentY := h.currentTouchPos[id].y
		referenceX := h.lastSignificantPos[id].x
		referenceY := h.lastSignificantPos[id].y
		dx := currentX - referenceX
		dy := currentY - referenceY
		distance := math.Sqrt(dx*dx + dy*dy)

		// Calculate swipe parameters
		angle := math.Atan2(dy, dx)  // Angle in radians
		elapsed := time.Since(h.touchStartTime[id])
		speed := distance / elapsed.Seconds()  // Pixels per second

		// Store previous swipe info for comparison
		prevSwipeInfo := h.GetSwipeInfo()

		// ---- SWIPE DETECTION ----
		// Only consider it a swipe if the distance threshold is met
		if distance >= h.swipeThreshold {
			// Update the active swipe data
			h.activeTouchID = id
			h.swipeDistance = distance
			h.currentSwipeAngle = angle
			h.currentSwipeSpeed = speed
			h.isHolding = true

			// Convert angle to degrees and determine cardinal direction
			// This maps the continuous angle to discrete directions:
			// - Right: 315° to 45°
			// - Down: 45° to 135°
			// - Left: 135° to 225°
			// - Up: 225° to 315°
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

			// Register as a new swipe if it's within the time threshold
			// This distinguishes quick swipes from slow drags
			if elapsed <= h.swipeDuration {
				h.detectedSwipes[direction] = true
			}

			// Update the holding direction regardless of time
			h.holdingDirection[direction] = true

			// Update the reference position if there's a significant change
			// in angle or distance. This prevents small jitter from changing
			// the swipe direction and allows for curved swipe paths.
			if prevSwipeInfo.Angle != 0 {
				// Consider updating if distance is significant
				needsUpdate := distance > 60
				if !needsUpdate {
					// Or if angle change is significant (about 11 degrees)
					angleDiff := math.Abs(angle - prevSwipeInfo.Angle)
					needsUpdate = angleDiff > 0.2
				}
				if needsUpdate {
					h.lastSignificantPos[id] = h.currentTouchPos[id]
				}
			}

			// Store the last detected direction for this touch
			h.lastDirection[id] = direction
		}
	}

	// ---- STEP 3: Process firing touch (right half) ----
	if h.fireTouchID != 0 {
		// Check if the fire touch has been released
		if inpututil.IsTouchJustReleased(h.fireTouchID) {
			// Reset fire touch state and return joystick to default position
			h.fireTouchID = 0
			h.fireHolding = false
			h.fireJoystickPos = h.fireDefaultPos
		} else {
			// Update the current position of the fire touch
			x, y := ebiten.TouchPosition(h.fireTouchID)
			h.fireCurrentPos = struct{ x, y float64 }{float64(x), float64(y)}

			// Calculate distance from initial touch position
			// This is used for both swipe detection and joystick position
			dx := h.fireCurrentPos.x - h.fireInitialPos.x
			dy := h.fireCurrentPos.y - h.fireInitialPos.y
			distance := math.Sqrt(dx*dx + dy*dy)

			// Detect fire swipe if threshold is met and not already in holding state
			if distance >= h.swipeThreshold && !h.fireHolding {
				// Register as a new fire swipe if it's within the time threshold
				if time.Since(h.fireStartTime) <= h.swipeDuration {
					h.fireJustSwiped = true
				}
				h.fireHolding = true

				// Note: The actual joystick position is updated here,
				// but the game would typically use GetFireJoystickPosition()
				// to determine firing direction
			}
		}
	}
}
