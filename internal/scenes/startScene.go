package scenes

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/gameplay/player"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type StartScene struct {
	buttonX         float64
	buttonY         float64
	buttonWidth     float64
	buttonHeight    float64
	lastScreenWidth  int
	lastScreenHeight int
}

func NewStartScene() *StartScene {
	return &StartScene{}
}

func (s *StartScene) Update(state *State) error {
	// Get current screen dimensions from the world
	screenWidth, screenHeight := state.World.GetWidth(), state.World.GetHeight()

	// Calculate button position and size if not set yet or if screen dimensions have changed
	if s.buttonWidth == 0 || screenWidth != s.lastScreenWidth || screenHeight != s.lastScreenHeight {
		// Set button dimensions (adjust size as needed)
		s.buttonWidth = float64(assets.PlayButton.Bounds().Dx()) * 0.2  // Scale to 20% of original size
		s.buttonHeight = float64(assets.PlayButton.Bounds().Dy()) * 0.2 // Scale to 20% of original size

		// Center the button on screen
		s.buttonX = float64(screenWidth)/2 - s.buttonWidth/2
		s.buttonY = float64(screenHeight)/2 - s.buttonHeight/2

		// Update last known screen dimensions
		s.lastScreenWidth = screenWidth
		s.lastScreenHeight = screenHeight
	}

	// Check for touch input (for mobile)
	// First check for new touches (just pressed) - this is the most important part for mobile
	justPressedIDs := inpututil.AppendJustPressedTouchIDs(nil)
	for _, id := range justPressedIDs {
		// Get touch position
		x, y := ebiten.TouchPosition(id)

		// Check if touch is within button bounds
		if float64(x) >= s.buttonX && float64(x) <= s.buttonX+s.buttonWidth &&
			float64(y) >= s.buttonY && float64(y) <= s.buttonY+s.buttonHeight {

			// If button is touched, transition to game scene immediately
			// This is critical for mobile responsiveness
			gameScene := NewGameScene(player.NewPlayer(state.World))
			state.SceneManager.GoToScene(gameScene)
			return nil // Return immediately to ensure the scene transition happens
		}
	}

	// Also check for existing touches (for compatibility)
	for _, id := range ebiten.TouchIDs() {
		// Skip touches that were just checked above
		isJustPressed := false
		for _, justPressedID := range justPressedIDs {
			if id == justPressedID {
				isJustPressed = true
				break
			}
		}
		if isJustPressed {
			continue
		}

		// Get touch position
		x, y := ebiten.TouchPosition(id)

		// Check if touch is within button bounds
		if float64(x) >= s.buttonX && float64(x) <= s.buttonX+s.buttonWidth &&
			float64(y) >= s.buttonY && float64(y) <= s.buttonY+s.buttonHeight {

			// For existing touches, only trigger on release to prevent accidental clicks
			if inpututil.IsTouchJustReleased(id) {
				// If button is clicked, transition to game scene
				gameScene := NewGameScene(player.NewPlayer(state.World))
				state.SceneManager.GoToScene(gameScene)
				break
			}
		}
	}

	// Check for mouse click (for desktop)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		// Check if click is within button bounds
		if float64(x) >= s.buttonX && float64(x) <= s.buttonX+s.buttonWidth &&
			float64(y) >= s.buttonY && float64(y) <= s.buttonY+s.buttonHeight {

			// If button is clicked, transition to game scene
			gameScene := NewGameScene(player.NewPlayer(state.World))
			state.SceneManager.GoToScene(gameScene)
		}
	}

	return nil
}

// Draw renders the start scene to the screen.
// It satisfies the Scene interface.
func (s *StartScene) Draw(screen *ebiten.Image, state *State) {
	// Get world dimensions from the state
	worldWidth, worldHeight := state.World.GetWidth(), state.World.GetHeight()

	// Draw background image scaled to fit screen
	bgOp := &ebiten.DrawImageOptions{}

	// Calculate scale to fit screen while maintaining aspect ratio
	startBg := assets.GetStartBackground()
	bgWidth := float64(startBg.Bounds().Dx())
	bgHeight := float64(startBg.Bounds().Dy())

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

	screen.DrawImage(assets.GetStartBackground(), bgOp)

	// Draw play button
	if s.buttonWidth > 0 {
		buttonOp := &ebiten.DrawImageOptions{}
		buttonOp.GeoM.Scale(s.buttonWidth/float64(assets.PlayButton.Bounds().Dx()), 
		                    s.buttonHeight/float64(assets.PlayButton.Bounds().Dy()))
		buttonOp.GeoM.Translate(s.buttonX, s.buttonY)
		screen.DrawImage(assets.PlayButton, buttonOp)
	}
}
