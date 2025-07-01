package scenes

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/gameplay/player"
	"discoveryx/internal/core/worldgen"
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
)

type GameScene struct {
	player         *player.Player
	generatedWorld *worldgen.GeneratedWorld
	cameraPosition math.Vector // Camera position for following the player
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

	// Get the player's position and update the generated world
	position := s.player.GetPosition()

	// Calculate screen dimensions
	screenWidth := float64(state.World.GetWidth())
	screenHeight := float64(state.World.GetHeight())

	s.generatedWorld.SetPlayerPosition(position.X, position.Y)

	// Update camera position to follow player
	// Reuse screen dimensions calculated above

	// Calculate how far the player is from the center of the screen (as a ratio)
	// This will be used to determine how strongly the camera should follow the player
	distanceFromCenterX := stdmath.Abs(position.X) / (screenWidth / 2)
	distanceFromCenterY := stdmath.Abs(position.Y) / (screenHeight / 2)

	// Apply non-linear scaling to make camera movement stronger near edges
	// and weaker near center - using a much higher factor (4.0) to ensure player stays on screen
	followStrengthX := stdmath.Pow(distanceFromCenterX, 2) * 4.0
	followStrengthY := stdmath.Pow(distanceFromCenterY, 2) * 4.0

	// If player is getting close to the edge, use full strength to prevent them from leaving the screen
	// Reduced threshold from 0.9 to 0.8 to start full strength follow earlier
	if distanceFromCenterX > 0.8 {
		followStrengthX = 1.0 // Full strength when near edge
	}
	if distanceFromCenterY > 0.8 {
		followStrengthY = 1.0 // Full strength when near edge
	}

	// Calculate target camera position (negative because we're moving the world in the opposite direction)
	targetCameraX := -position.X * followStrengthX
	targetCameraY := -position.Y * followStrengthY

	// Smoothly interpolate current camera position toward target position
	// Using a much higher interpolation factor (0.5) for very fast camera movement
	s.cameraPosition.X += (targetCameraX - s.cameraPosition.X) * 0.5
	s.cameraPosition.Y += (targetCameraY - s.cameraPosition.Y) * 0.5

	return nil
}

// Draw renders the game scene to the screen.
// It satisfies the Scene interface.
func (s *GameScene) Draw(screen *ebiten.Image, state *State) {
	// Get world dimensions from the state
	worldWidth, worldHeight := state.World.GetWidth(), state.World.GetHeight()

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

	screen.DrawImage(gameBg, bgOp)

	// Draw the generated world if it has been initialized
	if s.generatedWorld != nil {
		// Apply camera offset when drawing the world (convert float64 to int)
		s.generatedWorld.Draw(screen, int(s.cameraPosition.X), int(s.cameraPosition.Y))
	}

	// Draw the player with camera offset
	s.player.Draw(screen, s.cameraPosition.X, s.cameraPosition.Y)
}
