// Package game provides the main game loop and core functionality.
package game

import (
	"discoveryx/internal/core/ecs"
	"discoveryx/internal/core/gameplay/player"
	"discoveryx/internal/input"
	"discoveryx/internal/scenes"
	"discoveryx/internal/screen"
	"github.com/hajimehoshi/ebiten/v2"
	"time"
)

// Game is the main game struct that implements ebiten.Game interface.
// It manages the game state, scene transitions, and input handling.
type Game struct {
	player          *player.Player
	sceneManager    *scenes.SceneManager
	inputManager    *input.Manager
	screenManager   *screen.Manager
	lastUpdateTime  time.Time
	deltaTime       float64
	world           ecs.World
}

// GetWidth returns the width of the game world.
func (g *Game) GetWidth() int {
	return g.world.GetWidth()
}

// GetHeight returns the height of the game world.
func (g *Game) GetHeight() int {
	return g.world.GetHeight()
}

// SetWidth sets the width of the game world.
func (g *Game) SetWidth(width int) {
	g.world.SetWidth(width)
}

// SetHeight sets the height of the game world.
func (g *Game) SetHeight(height int) {
	g.world.SetHeight(height)
}

// ShouldMatchScreen returns true if the world dimensions should
// automatically match the screen dimensions.
func (g *Game) ShouldMatchScreen() bool {
	return g.world.ShouldMatchScreen()
}

// SetMatchScreen sets whether the world dimensions should
// automatically match the screen dimensions.
func (g *Game) SetMatchScreen(match bool) {
	g.world.SetMatchScreen(match)
}

// New creates a new game instance with initialized components.
func New() *Game {
	// Create screen manager first so it can be shared
	screenMgr := screen.New()

	// Create scene manager
	sceneManager := &scenes.SceneManager{}

	// Get initial screen dimensions
	width, height := screenMgr.GetWidth(), screenMgr.GetHeight()

	g := &Game{
		sceneManager:    sceneManager,
		inputManager:    input.NewManager(),
		screenManager:   screenMgr,
		lastUpdateTime:  time.Now(),
		deltaTime:       1.0 / 60.0, // Default to 60 FPS
		world:           ecs.NewBasicWorld(width, height),
	}

	// Set the screen manager for the scene manager
	sceneManager.SetScreenManager(screenMgr)

	// Load player sprite and create player
	g.player = player.NewPlayer(g)

	// Initialize with the start scene
	startScene := scenes.NewStartScene()
	g.sceneManager.GoToScene(startScene)

	return g
}

// Update updates the game state.
// It handles input processing and delegates to the current scene.
func (g *Game) Update() error {
	// Calculate delta time
	now := time.Now()
	g.deltaTime = now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	// Clamp deltaTime to avoid extreme values
	if g.deltaTime > 0.1 {
		g.deltaTime = 0.1 // Cap at 100ms to prevent huge jumps
	} else if g.deltaTime < 0.001 {
		g.deltaTime = 0.001 // Minimum of 1ms
	}

	// Update input handlers
	g.inputManager.Update()

	// Update the current scene through the scene manager
	return g.sceneManager.Update(g.inputManager, g.deltaTime, g.world)
}

// Draw renders the game to the screen.
// It delegates rendering to the current scene.
func (g *Game) Draw(screen *ebiten.Image) {
	// Let the scene manager handle drawing with all necessary state
	g.sceneManager.Draw(screen, g.inputManager, g.deltaTime, g.world)
}

// Layout implements ebiten.Game's Layout method.
// It updates the game dimensions and input manager when the window is resized.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// Calculate layout using screen manager
	screenWidth, screenHeight = g.screenManager.CalculateLayout(outsideWidth, outsideHeight)

	// Update input manager with screen dimensions
	g.inputManager.SetScreenDimensions(outsideWidth, outsideHeight)

	// Update world dimensions only if they should match the screen
	if g.world.ShouldMatchScreen() {
		g.world.SetWidth(screenWidth)
		g.world.SetHeight(screenHeight)
	}

	return screenWidth, screenHeight
}

// GoToScene changes to a new scene.
// It delegates to the scene manager's GoToScene method.
func (g *Game) GoToScene(scene scenes.Scene) {
	g.sceneManager.GoToScene(scene)
}

// SetDynamicResizing enables or disables dynamic screen resizing.
// When disabled, the game will maintain a fixed size regardless of window size.
func (g *Game) SetDynamicResizing(enabled bool) {
	g.screenManager.SetDynamicResizing(enabled)
}

// IsDynamicResizingEnabled returns whether dynamic resizing is enabled.
func (g *Game) IsDynamicResizingEnabled() bool {
	return g.screenManager.IsDynamicResizingEnabled()
}
