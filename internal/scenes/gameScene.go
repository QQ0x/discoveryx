package scenes

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/constants"
	"discoveryx/internal/core/gameplay/enemies"
	"discoveryx/internal/core/gameplay/player"
	"discoveryx/internal/core/worldgen"
	"discoveryx/internal/rendering/shaders"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
)

type GameScene struct {
	player          *player.Player
	generatedWorld  *worldgen.GeneratedWorld
	cameraPosition  math.Vector      // Camera position for following the player
	enemies         []*enemies.Enemy // List of enemies in the scene
	brightnessShader *shaders.BrightnessShader // Shader for brightness effect
}

func NewGameScene(player *player.Player) *GameScene {
	return &GameScene{
		player:         player,
		cameraPosition: math.Vector{X: 0, Y: 0},
	}
}

// Initialize initializes the game scene with the generated world
func (s *GameScene) Initialize(state *State) error {
	// Create a world generator
	generator, err := worldgen.NewWorldGenerator()
	if err != nil {
		return err
	}

	// Use default configuration
	config := worldgen.DefaultWorldGenConfig()

	// Create a new generated world with the same dimensions as the state world
	s.generatedWorld, err = worldgen.NewGeneratedWorld(
		state.World.GetWidth(),
		state.World.GetHeight(),
		generator,
		config,
	)
	if err != nil {
		return err
	}

	// Initialize the brightness shader
	s.brightnessShader, err = shaders.NewBrightnessShader()
	if err != nil {
		return err
	}

	// Spawn objects (enemies) on walls
	objectTypes := []string{"enemy_1"}
	s.enemies = enemies.SpawnObjectsOnWalls(s.generatedWorld, objectTypes, 1.0, 32.0) // 100% chance per wall, minimum distance 32 units

	// Provisionally set player position to the first enemy spawn point if enemies were spawned
	if len(s.enemies) > 0 {
		firstEnemyPos := s.enemies[0].Position
		s.player.SetPosition(firstEnemyPos)
	}

	return nil
}

// Update handles the game logic for the game scene.
// It satisfies the Scene interface.
func (s *GameScene) Update(state *State) error {
	// Initialize the generated world if it hasn't been initialized yet
	if s.generatedWorld == nil {
		if err := s.Initialize(state); err != nil {
			return err
		}
	}

	// Update the player
	if err := s.player.Update(state.Input, state.DeltaTime); err != nil {
		return err
	}

	// Update enemies
	for _, enemy := range s.enemies {
		if err := enemy.Update(); err != nil {
			return err
		}
	}

	// Get the player's position and update the generated world
	position := s.player.GetPosition()

	// Calculate screen dimensions
	screenWidth := float64(state.World.GetWidth())
	screenHeight := float64(state.World.GetHeight())

	s.generatedWorld.SetPlayerPosition(position.X, position.Y)

	// Update camera position to follow player
	// Reuse screen dimensions calculated above

	// Get player velocity to determine if we should center the camera
	playerVelocity := s.player.GetVelocity()

	// Calculate the current camera target (where the camera is looking at)
	// This is the inverse of the camera position since camera position is an offset
	cameraTargetX := -s.cameraPosition.X
	cameraTargetY := -s.cameraPosition.Y

	// Calculate how far the player is from the camera target
	offsetX := position.X - cameraTargetX
	offsetY := position.Y - cameraTargetY

	// Calculate the deadzone dimensions
	deadZoneWidth := screenWidth * constants.CameraDeadZoneX
	deadZoneHeight := screenHeight * constants.CameraDeadZoneY

	// Calculate new camera target position
	var newCameraTargetX, newCameraTargetY float64

	// When player is moving at normal speed, use deadzone-based following
	if playerVelocity >= constants.CameraVelocityThreshold {
		// Start with current camera target
		newCameraTargetX = cameraTargetX
		newCameraTargetY = cameraTargetY

		// X-axis deadzone calculation
		if stdmath.Abs(offsetX) > deadZoneWidth/2 {
			// Player is outside deadzone, move camera target toward player
			// but only by the amount they're outside the deadzone
			if offsetX > 0 {
				// Player is to the right of deadzone
				newCameraTargetX = position.X - deadZoneWidth/2
			} else {
				// Player is to the left of deadzone
				newCameraTargetX = position.X + deadZoneWidth/2
			}
		}

		// Y-axis deadzone calculation
		if stdmath.Abs(offsetY) > deadZoneHeight/2 {
			// Player is outside deadzone, move camera target toward player
			// but only by the amount they're outside the deadzone
			if offsetY > 0 {
				// Player is below deadzone
				newCameraTargetY = position.Y - deadZoneHeight/2
			} else {
				// Player is above deadzone
				newCameraTargetY = position.Y + deadZoneHeight/2
			}
		}
	} else {
		// When player is slow/stopped, gradually center on player
		// Calculate centering strength based on player velocity
		centeringFactor := (constants.CameraVelocityThreshold - playerVelocity) / constants.CameraVelocityThreshold

		// Blend between current target and player position based on centering factor
		newCameraTargetX = cameraTargetX + (position.X-cameraTargetX)*centeringFactor*constants.CameraCenteringStrength
		newCameraTargetY = cameraTargetY + (position.Y-cameraTargetY)*centeringFactor*constants.CameraCenteringStrength
	}

	// Convert camera target to camera position (which is the negative of the target)
	targetCameraX := -newCameraTargetX
	targetCameraY := -newCameraTargetY

	// Apply frame-rate independent smoothing
	// This ensures consistent camera movement regardless of frame rate
	interpolationFactor := 1.0 - stdmath.Pow(1.0-constants.CameraInterpolationFactor, state.DeltaTime*60.0)

	// Smoothly interpolate current camera position toward target position
	s.cameraPosition.X += (targetCameraX - s.cameraPosition.X) * interpolationFactor
	s.cameraPosition.Y += (targetCameraY - s.cameraPosition.Y) * interpolationFactor

	return nil
}

// Draw renders the game scene to the screen.
// It satisfies the Scene interface.
func (s *GameScene) Draw(screen *ebiten.Image, state *State) {
	// Get world dimensions from the state
	worldWidth, worldHeight := state.World.GetWidth(), state.World.GetHeight()

	// Create a temporary image to draw all scene elements
	tempScreen := ebiten.NewImage(worldWidth, worldHeight)

	// Draw background image scaled to fit screen
	bgOp := &ebiten.DrawImageOptions{}

	// Calculate scale to fit screen while maintaining aspect ratio
	gameBg := assets.GetGameBackground()
	bgWidth := float64(gameBg.Bounds().Dx())
	bgHeight := float64(gameBg.Bounds().Dy())

	scaleX := float64(worldWidth) / bgWidth
	scaleY := float64(worldHeight) / bgHeight

	// Use the larger scale to ensure the image covers the entire screen
	scale := scaleX
	if scaleY > scaleX {
		scale = scaleY
	}

	bgOp.GeoM.Scale(scale, scale)

	// Center the background
	scaledWidth := bgWidth * scale
	scaledHeight := bgHeight * scale
	bgOp.GeoM.Translate((float64(worldWidth)-scaledWidth)/2, (float64(worldHeight)-scaledHeight)/2)

	tempScreen.DrawImage(gameBg, bgOp)

	// Draw the generated world if it has been initialized
	if s.generatedWorld != nil {
		// Apply camera offset when drawing the world (using float64 for smoother movement)
		s.generatedWorld.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y)
	}

	// Draw the enemies with camera offset
	for _, enemy := range s.enemies {
		// Draw the enemy with camera offset and world dimensions
		enemy.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y, worldWidth, worldHeight)
	}

	// Draw the player with camera offset
	s.player.Draw(tempScreen, s.cameraPosition.X, s.cameraPosition.Y)

	// Apply brightness shader if initialized
	if s.brightnessShader != nil {
		// Get player position in screen coordinates
		playerPos := s.player.GetPosition()
		screenPosX := playerPos.X + s.cameraPosition.X + float64(worldWidth)/2
		screenPosY := playerPos.Y + s.cameraPosition.Y + float64(worldHeight)/2

		// Set up shader options
		op := &ebiten.DrawRectShaderOptions{}
		op.Images[0] = tempScreen
		op.Uniforms = map[string]any{
			"PlayerPos": []float32{float32(screenPosX), float32(screenPosY)},
			"Radius":    float32(float64(worldWidth) * 0.75), // Use 75% of screen width as radius
		}

		// Apply shader to the screen
		screen.DrawRectShader(worldWidth, worldHeight, s.brightnessShader.Shader(), op)
	} else {
		// If shader not initialized, just draw the temp screen directly
		screen.DrawImage(tempScreen, nil)
	}
}
