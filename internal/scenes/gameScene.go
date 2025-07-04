package scenes

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/constants"
	"discoveryx/internal/core/gameplay/enemies"
	"discoveryx/internal/core/gameplay/player"
	"discoveryx/internal/core/gameplay/projectiles"
	"discoveryx/internal/core/worldgen"
	"discoveryx/internal/input"
	"discoveryx/internal/rendering/shaders"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	stdmath "math"
)

// GameScene represents the main gameplay scene with player, enemies, and world
type GameScene struct {
	player            *player.Player
	generatedWorld    *worldgen.GeneratedWorld
	cameraPosition    math.Vector
	enemies           []*enemies.Enemy
	brightnessShader  *shaders.BrightnessShader
	bullets           []*projectiles.Bullet
	timeSinceLastShot float64
}

// NewGameScene creates a new game scene with the provided player
func NewGameScene(player *player.Player) *GameScene {
	return &GameScene{
		player:            player,
		cameraPosition:    math.Vector{X: 0, Y: 0},
		timeSinceLastShot: 0,
	}
}

// Initialize sets up the game scene with world generation, shaders, and enemy placement
func (s *GameScene) Initialize(state *State) error {
	generator, err := worldgen.NewWorldGenerator()
	if err != nil {
		return err
	}

	config := worldgen.DefaultWorldGenConfig()

	s.generatedWorld, err = worldgen.NewGeneratedWorld(
		state.World.GetWidth(),
		state.World.GetHeight(),
		generator,
		config,
	)
	if err != nil {
		return err
	}

	s.brightnessShader, err = shaders.NewBrightnessShader()
	if err != nil {
		return err
	}

	objectTypes := []string{"enemy_1"}
	s.enemies = enemies.SpawnObjectsOnWalls(s.generatedWorld, objectTypes, 1.0, 32.0)

	if len(s.enemies) > 0 {
		firstEnemyPos := s.enemies[0].Position
		s.player.SetPosition(firstEnemyPos)
	}

	return nil
}

// Update handles the game logic and camera movement for the scene
func (s *GameScene) Update(state *State) error {
	if s.generatedWorld == nil {
		if err := s.Initialize(state); err != nil {
			return err
		}
	}

	if err := s.player.Update(state.Input, state.DeltaTime); err != nil {
		return err
	}

	for _, enemy := range s.enemies {
		if err := enemy.Update(); err != nil {
			return err
		}
	}

	s.handleShooting(state)
	s.handleEnemyShooting(state)
	var activeBullets []*projectiles.Bullet
	for _, b := range s.bullets {
		if !b.Update(state.DeltaTime) {
			activeBullets = append(activeBullets, b)
		}
	}
	s.bullets = activeBullets

	s.resolveBulletCollisions()
	s.resolvePlayerWallCollision()
	s.resolveEnemyDeaths()

	position := s.player.GetPosition()
	screenWidth := float64(state.World.GetWidth())
	screenHeight := float64(state.World.GetHeight())

	s.generatedWorld.SetPlayerPosition(position.X, position.Y)

	// Camera system implementation
	playerVelocity := s.player.GetVelocity()
	cameraTargetX := -s.cameraPosition.X
	cameraTargetY := -s.cameraPosition.Y
	offsetX := position.X - cameraTargetX
	offsetY := position.Y - cameraTargetY
	deadZoneWidth := screenWidth * constants.CameraDeadZoneX
	deadZoneHeight := screenHeight * constants.CameraDeadZoneY

	var newCameraTargetX, newCameraTargetY float64

	if playerVelocity >= constants.CameraVelocityThreshold {
		// Use deadzone-based camera following for normal movement speed
		newCameraTargetX = cameraTargetX
		newCameraTargetY = cameraTargetY

		if stdmath.Abs(offsetX) > deadZoneWidth/2 {
			if offsetX > 0 {
				newCameraTargetX = position.X - deadZoneWidth/2
			} else {
				newCameraTargetX = position.X + deadZoneWidth/2
			}
		}

		if stdmath.Abs(offsetY) > deadZoneHeight/2 {
			if offsetY > 0 {
				newCameraTargetY = position.Y - deadZoneHeight/2
			} else {
				newCameraTargetY = position.Y + deadZoneHeight/2
			}
		}
	} else {
		// Gradually center on player when moving slowly
		centeringFactor := (constants.CameraVelocityThreshold - playerVelocity) / constants.CameraVelocityThreshold
		newCameraTargetX = cameraTargetX + (position.X-cameraTargetX)*centeringFactor*constants.CameraCenteringStrength
		newCameraTargetY = cameraTargetY + (position.Y-cameraTargetY)*centeringFactor*constants.CameraCenteringStrength
	}

	targetCameraX := -newCameraTargetX
	targetCameraY := -newCameraTargetY

	// Frame-rate independent camera smoothing
	interpolationFactor := 1.0 - stdmath.Pow(1.0-constants.CameraInterpolationFactor, state.DeltaTime*60.0)
	s.cameraPosition.X += (targetCameraX - s.cameraPosition.X) * interpolationFactor
	s.cameraPosition.Y += (targetCameraY - s.cameraPosition.Y) * interpolationFactor

	return nil
}

// Draw renders the game scene with background, world, entities and lighting effects
func (s *GameScene) Draw(screen *ebiten.Image, state *State) {
	worldWidth, worldHeight := state.World.GetWidth(), state.World.GetHeight()
	tempScreen := ebiten.NewImage(worldWidth, worldHeight)

	// Scale background to fit screen while maintaining aspect ratio
	bgOp := &ebiten.DrawImageOptions{}
	gameBg := assets.GetGameBackground()
	bgWidth := float64(gameBg.Bounds().Dx())
	bgHeight := float64(gameBg.Bounds().Dy())

	scaleX := float64(worldWidth) / bgWidth
	scaleY := float64(worldHeight) / bgHeight
	scale := scaleX
	if scaleY > scaleX {
		scale = scaleY
	}

	bgOp.GeoM.Scale(scale, scale)

	scaledWidth := bgWidth * scale
	scaledHeight := bgHeight * scale
	bgOp.GeoM.Translate((float64(worldWidth)-scaledWidth)/2, (float64(worldHeight)-scaledHeight)/2)

	tempScreen.DrawImage(gameBg, bgOp)

	if s.generatedWorld != nil {
		s.generatedWorld.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y)
	}

	for _, enemy := range s.enemies {
		enemy.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y, worldWidth, worldHeight)
	}

	for _, b := range s.bullets {
		b.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y, worldWidth, worldHeight)
	}

	s.player.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y)

	// Apply lighting effect with brightness shader
	if s.brightnessShader != nil {
		playerPos := s.player.GetPosition()
		screenPosX := playerPos.X + s.cameraPosition.X + float64(worldWidth)/2
		screenPosY := playerPos.Y + s.cameraPosition.Y + float64(worldHeight)/2

		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = tempScreen
		op.Uniforms = map[string]any{
			"PlayerPos": []float32{float32(screenPosX), float32(screenPosY)},
			"Radius":    float32(float64(worldWidth) * 0.75),
		}

		screen.DrawRectShader(worldWidth, worldHeight, s.brightnessShader.Shader(), op)
	} else {
		screen.DrawImage(tempScreen, nil)
	}
}

const fireInterval = 0.25
const enemyFireInterval = 0.5
const enemyShootRadius = 150.0

// handleShooting manages bullet firing based on keyboard or touch input
func (s *GameScene) handleShooting(state *State) {
	touch := state.Input.Touch()

	if inpututil.IsKeyJustPressed(input.KeySpace) || (touch != nil && touch.IsFireJustSwiped()) {
		s.spawnBullet()
		s.timeSinceLastShot = 0
		return
	}

	holding := state.Input.Keyboard().IsKeyPressed(input.KeySpace)
	if touch != nil && touch.IsFireHolding() {
		holding = true
	}

	if holding {
		s.timeSinceLastShot += state.DeltaTime
		if s.timeSinceLastShot >= fireInterval {
			s.spawnBullet()
			s.timeSinceLastShot = 0
		}
	} else {
		s.timeSinceLastShot = fireInterval
	}
}

// handleEnemyShooting makes enemies fire bullets at the player when in range
func (s *GameScene) handleEnemyShooting(state *State) {
	playerPos := s.player.GetPosition()
	for _, enemy := range s.enemies {
		dx := playerPos.X - enemy.Position.X
		dy := playerPos.Y - enemy.Position.Y
		if dx*dx+dy*dy <= enemyShootRadius*enemyShootRadius {
			enemy.TimeSinceLastShot += state.DeltaTime
			if enemy.TimeSinceLastShot >= enemyFireInterval {
				rot := stdmath.Atan2(-dy, dx)
				// Create a copy of the enemy position to ensure bullets spawn from the enemy
				// Add an offset in the direction of the player to make bullets spawn from the edge of the enemy
				offsetDistance := 10.0 // Distance from enemy center to spawn the bullet
				offsetX := stdmath.Sin(rot) * offsetDistance
				offsetY := stdmath.Cos(rot) * -offsetDistance
				enemyPos := math.Vector{X: enemy.Position.X + offsetX, Y: enemy.Position.Y + offsetY}
				s.bullets = append(s.bullets, projectiles.NewLinearBullet(enemyPos, rot, assets.EnemyBullet))
				enemy.TimeSinceLastShot = 0
			}
		} else if enemy.TimeSinceLastShot > enemyFireInterval {
			enemy.TimeSinceLastShot = enemyFireInterval
		}
	}
}

// spawnBullet creates a new bullet at the player's position with the player's rotation
func (s *GameScene) spawnBullet() {
	pos := s.player.GetPosition()
	rot := s.player.GetRotation()
	s.bullets = append(s.bullets, projectiles.NewBullet(pos, rot, assets.PlayerBullet))
}

// resolveBulletCollisions checks all bullets against enemies and the player.
func (s *GameScene) resolveBulletCollisions() {
	var remaining []*projectiles.Bullet
	for _, b := range s.bullets {
		collided := false
		if b.FromPlayer {
			for _, e := range s.enemies {
				if distanceSquared(b.Position, e.Position) <= constants.BulletRadius*constants.BulletRadius+constants.EnemyRadius*constants.EnemyRadius {
					e.TakeDamage(b.Damage)
					collided = true
					break
				}
			}
		} else {
			if distanceSquared(b.Position, s.player.GetPosition()) <= constants.BulletRadius*constants.BulletRadius+constants.PlayerRadius*constants.PlayerRadius {
				s.player.TakeDamage(b.Damage)
				collided = true
			}
		}

		if s.bulletHitsWall(b.Position) {
			collided = true
		}

		if !collided {
			remaining = append(remaining, b)
		}
	}
	s.bullets = remaining
}

// distanceSquared returns squared distance between two vectors.
func distanceSquared(a, b math.Vector) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return dx*dx + dy*dy
}

// bulletHitsWall checks if a bullet collides with any nearby wall.
func (s *GameScene) bulletHitsWall(pos math.Vector) bool {
	cell := s.generatedWorld.GetCellAt(int(pos.X), int(pos.Y))
	if cell == nil {
		return false
	}
	walls := cell.GetWallsInWorldCoordinates()
	for _, w := range walls {
		dx := pos.X - w.X
		dy := pos.Y - w.Y
		if dx*dx+dy*dy <= constants.BulletRadius*constants.BulletRadius {
			return true
		}
	}
	return false
}

// resolvePlayerWallCollision prevents the player from moving through walls.
func (s *GameScene) resolvePlayerWallCollision() {
	pos := s.player.GetPosition()
	cell := s.generatedWorld.GetCellAt(int(pos.X), int(pos.Y))
	if cell == nil {
		return
	}
	walls := cell.GetWallsInWorldCoordinates()
	for _, w := range walls {
		dx := pos.X - w.X
		dy := pos.Y - w.Y
		if dx*dx+dy*dy <= constants.PlayerRadius*constants.PlayerRadius {
			// Push player out along wall normal
			pos.X = w.X + w.Normal.X*constants.PlayerRadius
			pos.Y = w.Y + w.Normal.Y*constants.PlayerRadius
			s.player.TakeDamage(constants.WallCollisionDamage)
		}
	}
	s.player.SetPosition(pos)
}

// resolveEnemyDeaths removes dead enemies from the scene.
func (s *GameScene) resolveEnemyDeaths() {
	var alive []*enemies.Enemy
	for _, e := range s.enemies {
		if !e.IsDead() {
			alive = append(alive, e)
		}
	}
	s.enemies = alive
}
