package scenes

import (
	"discoveryx/internal/assets"
	"discoveryx/internal/core/gameplay/player"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// StartScene represents the initial menu screen with a play button
type StartScene struct {
	buttonX         float64
	buttonY         float64
	buttonWidth     float64
	buttonHeight    float64
	lastScreenWidth  int
	lastScreenHeight int
}

// NewStartScene creates a new start menu scene
func NewStartScene() *StartScene {
	return &StartScene{}
}

// Update handles input processing and scene transitions
func (s *StartScene) Update(state *State) error {
	screenWidth, screenHeight := state.World.GetWidth(), state.World.GetHeight()

	if s.buttonWidth == 0 || screenWidth != s.lastScreenWidth || screenHeight != s.lastScreenHeight {
		s.buttonWidth = float64(assets.PlayButton.Bounds().Dx()) * 0.2
		s.buttonHeight = float64(assets.PlayButton.Bounds().Dy()) * 0.2

		s.buttonX = float64(screenWidth)/2 - s.buttonWidth/2
		s.buttonY = float64(screenHeight)/2 - s.buttonHeight/2

		s.lastScreenWidth = screenWidth
		s.lastScreenHeight = screenHeight
	}

	// Handle mobile touch input
	justPressedIDs := inpututil.AppendJustPressedTouchIDs(nil)
	for _, id := range justPressedIDs {
		x, y := ebiten.TouchPosition(id)

		if float64(x) >= s.buttonX && float64(x) <= s.buttonX+s.buttonWidth &&
			float64(y) >= s.buttonY && float64(y) <= s.buttonY+s.buttonHeight {

			gameScene := NewGameScene(player.NewPlayer(state.World))
			state.SceneManager.GoToScene(gameScene)
			return nil
		}
	}

	// Handle existing touches for compatibility
	for _, id := range ebiten.TouchIDs() {
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

		x, y := ebiten.TouchPosition(id)

		if float64(x) >= s.buttonX && float64(x) <= s.buttonX+s.buttonWidth &&
			float64(y) >= s.buttonY && float64(y) <= s.buttonY+s.buttonHeight {

			if inpututil.IsTouchJustReleased(id) {
				gameScene := NewGameScene(player.NewPlayer(state.World))
				state.SceneManager.GoToScene(gameScene)
				break
			}
		}
	}

	// Handle desktop mouse input
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		if float64(x) >= s.buttonX && float64(x) <= s.buttonX+s.buttonWidth &&
			float64(y) >= s.buttonY && float64(y) <= s.buttonY+s.buttonHeight {

			gameScene := NewGameScene(player.NewPlayer(state.World))
			state.SceneManager.GoToScene(gameScene)
		}
	}

	return nil
}

// Draw renders the start scene with background and play button
func (s *StartScene) Draw(screen *ebiten.Image, state *State) {
	worldWidth, worldHeight := state.World.GetWidth(), state.World.GetHeight()

	bgOp := &ebiten.DrawImageOptions{}
	startBg := assets.GetStartBackground()
	bgWidth := float64(startBg.Bounds().Dx())
	bgHeight := float64(startBg.Bounds().Dy())

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

	screen.DrawImage(assets.GetStartBackground(), bgOp)

	if s.buttonWidth > 0 {
		buttonOp := &ebiten.DrawImageOptions{}
		buttonOp.GeoM.Scale(s.buttonWidth/float64(assets.PlayButton.Bounds().Dx()), 
		                    s.buttonHeight/float64(assets.PlayButton.Bounds().Dy()))
		buttonOp.GeoM.Translate(s.buttonX, s.buttonY)
		screen.DrawImage(assets.PlayButton, buttonOp)
	}
}
