package player

import (
	"discoveryx/internal/core/ecs"
	"discoveryx/internal/core/gameplay/projectiles"
	"discoveryx/internal/input"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	stdmath "math"
	"time"
)

// WeaponSystem manages the player's weapons and bullets
type WeaponSystem struct {
	player        *Player
	world         ecs.World
	bullets       []*projectiles.Bullet
	lastFireTime  time.Time
	fireRate      time.Duration
	rightJoystick *RightJoystick
	isFiring      bool
	bulletSpeed   float64
}

// RightJoystick represents the invisible joystick on the right half of the screen
type RightJoystick struct {
	active   bool
	touchID  ebiten.TouchID
	centerX  float64
	centerY  float64
	currentX float64
	currentY float64
	angle    float64
	distance float64
}

// NewWeaponSystem creates a new weapon system
func NewWeaponSystem(player *Player, world ecs.World) *WeaponSystem {
	return &WeaponSystem{
		player:        player,
		world:         world,
		bullets:       make([]*projectiles.Bullet, 0, 100),
		lastFireTime:  time.Now(),
		fireRate:      time.Millisecond * 200, // Fire a bullet every 200ms
		rightJoystick: &RightJoystick{},
		bulletSpeed:   20.0, // Increased from 10.0 to 20.0 for faster bullet movement
	}
}

// Update updates the weapon system
func (ws *WeaponSystem) Update(inputManager *input.Manager, deltaTime float64) {
	println("WeaponSystem.Update called. Current bullets:", len(ws.bullets))

	// Update right joystick
	ws.updateRightJoystick(inputManager.Touch())

	// Check if it's time to fire based on fire rate
	canFire := time.Since(ws.lastFireTime) > ws.fireRate
	println("Can fire:", canFire, "Time since last fire:", time.Since(ws.lastFireTime).Milliseconds(), "ms")

	// Check joystick status
	println("Joystick active:", ws.rightJoystick.active)

	// Check keyboard status
	spacePressed := inputManager.Keyboard().IsKeyPressed(input.KeySpace)
	println("Space pressed:", spacePressed)

	// Fire bullets if joystick is active
	if ws.rightJoystick.active && canFire {
		println("Firing bullet due to active joystick")
		ws.fireBullet()
		ws.lastFireTime = time.Now()
	}

	// Fire bullets if spacebar is pressed (for PC)
	if spacePressed && canFire {
		println("Firing bullet due to spacebar press")
		ws.fireBullet()
		ws.lastFireTime = time.Now()
	}

	// Update bullets
	println("Updating", len(ws.bullets), "bullets")
	for i := len(ws.bullets) - 1; i >= 0; i-- {
		println("Updating bullet", i)
		ws.bullets[i].Update(deltaTime)

		// Remove inactive bullets
		if !ws.bullets[i].IsActive() {
			println("Removing inactive bullet", i)
			// Remove bullet by swapping with the last element and truncating the slice
			ws.bullets[i] = ws.bullets[len(ws.bullets)-1]
			ws.bullets = ws.bullets[:len(ws.bullets)-1]
			println("Bullets remaining after removal:", len(ws.bullets))
		}
	}

	// Force create a bullet for testing if no bullets exist
	if len(ws.bullets) == 0 {
		println("No bullets exist, creating a test bullet")
		ws.fireBullet()
	}
}

// Draw draws all active bullets
func (ws *WeaponSystem) Draw(screen *ebiten.Image, cameraOffsetX, cameraOffsetY float64) {
	println("WeaponSystem.Draw called with", len(ws.bullets), "bullets")

	for i, bullet := range ws.bullets {
		println("Drawing bullet", i)
		bullet.Draw(screen, cameraOffsetX, cameraOffsetY)
	}
}

// fireBullet creates a new bullet and adds it to the bullets slice
func (ws *WeaponSystem) fireBullet() {
	println("fireBullet method called")

	playerPos := ws.player.GetPosition()
	println("Player position:", playerPos.X, playerPos.Y)

	// If joystick is active, use its angle for bullet direction
	var direction float64
	if ws.rightJoystick.active {
		direction = ws.rightJoystick.angle
		println("Using joystick angle for bullet direction:", direction)
	} else {
		// Otherwise use player's rotation
		direction = ws.player.GetRotation()
		println("Using player rotation for bullet direction:", direction)
	}

	// Calculate bullet spawn position at the front of the player
	// Player sprite size is scaled to 1/3 in the Draw method, so we need to account for that
	// Assuming the player sprite is roughly 30 pixels in size after scaling
	const playerRadius = 15.0
	// Add some extra distance to ensure the bullet appears clearly in front of the player
	// Increased from 5.0 to 15.0 to position the bullet further in front of the player
	const extraDistance = 15.0

	// Calculate offset position using trigonometry
	offsetX := stdmath.Sin(direction) * (playerRadius + extraDistance)
	offsetY := -stdmath.Cos(direction) * (playerRadius + extraDistance)
	println("Bullet offset:", offsetX, offsetY)

	// Create bullet position by adding offset to player position
	bulletPos := math.Vector{
		X: playerPos.X + offsetX,
		Y: playerPos.Y + offsetY,
	}
	println("Bullet position:", bulletPos.X, bulletPos.Y)

	// Create a new bullet with higher velocity for faster movement
	bullet := projectiles.NewBullet(ws.world, bulletPos, direction, ws.bulletSpeed)
	println("Created new bullet with velocity:", ws.bulletSpeed)

	// Add the bullet to the bullets slice
	ws.bullets = append(ws.bullets, bullet)
	println("Added bullet to bullets slice. Total bullets:", len(ws.bullets))
}

// updateRightJoystick updates the right joystick state based on touch input
func (ws *WeaponSystem) updateRightJoystick(touch input.TouchHandler) {
	if touch == nil {
		ws.rightJoystick.active = false
		return
	}

	// Get screen dimensions
	halfWidth := ws.world.GetWidth() / 2

	// Process just pressed touches
	justPressedIDs := inpututil.AppendJustPressedTouchIDs(nil)
	for _, id := range justPressedIDs {
		x, y := ebiten.TouchPosition(id)

		// Only process touches on the right half of the screen
		if x < halfWidth {
			continue
		}

		// Set the joystick center for new touches
		ws.rightJoystick.active = true
		ws.rightJoystick.touchID = id
		ws.rightJoystick.centerX = float64(x)
		ws.rightJoystick.centerY = float64(y)
		ws.rightJoystick.currentX = float64(x)
		ws.rightJoystick.currentY = float64(y)
		ws.rightJoystick.distance = 0
		ws.rightJoystick.angle = 0
	}

	// Process all active touches
	touchIDs := ebiten.AppendTouchIDs(nil)
	for _, id := range touchIDs {
		x, y := ebiten.TouchPosition(id)

		// Only process touches on the right half of the screen
		if x < halfWidth {
			continue
		}

		// If this is the active touch, update joystick position
		if ws.rightJoystick.active && ws.rightJoystick.touchID == id {
			if inpututil.IsTouchJustReleased(id) {
				// Reset joystick when touch is released
				ws.rightJoystick.active = false
			} else {
				// Update joystick position
				ws.rightJoystick.currentX = float64(x)
				ws.rightJoystick.currentY = float64(y)

				// Calculate distance and angle
				dx := ws.rightJoystick.currentX - ws.rightJoystick.centerX
				dy := ws.rightJoystick.currentY - ws.rightJoystick.centerY
				ws.rightJoystick.distance = stdmath.Sqrt(dx*dx + dy*dy)
				ws.rightJoystick.angle = stdmath.Atan2(dy, dx)
			}
		}
	}
}
