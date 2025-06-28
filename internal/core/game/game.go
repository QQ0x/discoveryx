// Package game provides the main game loop and core functionality.
package game

import (
	"discoveryx/internal/constants"
	// "discoveryx/internal/core/ecs"
	"discoveryx/internal/core/gameplay/player"
	"discoveryx/internal/input"
	"discoveryx/internal/scenes"
	"github.com/hajimehoshi/ebiten/v2"
)

// Game is the main game struct that implements ebiten.Game interface.
// It manages the game state, scene transitions, and input handling.
type Game struct {
	player       *player.Player
	width        int
	height       int
	sceneManager *scenes.SceneManager
	inputManager *input.Manager
}

// GetWidth returns the width of the game world.
func (g *Game) GetWidth() int {
	return g.width
}

// GetHeight returns the height of the game world.
func (g *Game) GetHeight() int {
	return g.height
}

// New creates a new game instance with initialized components.
func New() *Game {
	g := &Game{
		width:        constants.ScreenWidth,
		height:       constants.ScreenHeight,
		sceneManager: &scenes.SceneManager{},
		inputManager: input.NewManager(),
	}

	// Load player sprite and create player
	g.player = player.NewPlayer(g)

	// Initialize with the default scene
	gameScene := scenes.NewGameScene(g.player)
	g.sceneManager.GoToScene(gameScene)

	return g
}

// Update updates the game state.
// It handles input processing and delegates to the current scene.
func (g *Game) Update() error {
	// Update input handlers
	g.inputManager.Update()

	// Update the current scene through the scene manager
	return g.sceneManager.Update(g.inputManager)
}

// Draw renders the game to the screen.
// It delegates rendering to the current scene.
func (g *Game) Draw(screen *ebiten.Image) {
	// Let the scene manager handle drawing
	g.sceneManager.Draw(screen)
}

// Layout implements ebiten.Game's Layout method.
// It updates the game dimensions and input manager when the window is resized.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// Update game dimensions
	g.width = outsideWidth
	g.height = outsideHeight

	// Update input manager with screen dimensions
	g.inputManager.SetScreenDimensions(outsideWidth, outsideHeight)

	return outsideWidth, outsideHeight
}

// GoToScene changes to a new scene.
// It delegates to the scene manager's GoToScene method.
func (g *Game) GoToScene(scene scenes.Scene) {
	g.sceneManager.GoToScene(scene)
}
