// Package game provides the main game loop and core functionality.
// It serves as the central coordinator for all game systems including
// scene management, input handling, rendering, and game state.
// This package implements the ebiten.Game interface required by the Ebiten game engine.
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
// The Game struct serves as the central hub that coordinates all game subsystems:
// - Player management and control
// - Scene management for different game states (menu, gameplay, etc.)
// - Input handling for keyboard, mouse, and touch
// - Screen management for resolution and scaling
// - World management using an Entity Component System (ECS)
// - Time management for frame-rate independent updates
type Game struct {
	player          *player.Player    // The player entity with its state and behavior
	sceneManager    *scenes.SceneManager // Manages scene transitions and rendering
	inputManager    *input.Manager    // Handles and processes user input
	screenManager   *screen.Manager   // Manages screen dimensions and scaling
	lastUpdateTime  time.Time         // Timestamp of the last update for delta time calculation
	deltaTime       float64           // Time elapsed since the last frame in seconds
	world           ecs.World         // The ECS world containing all game entities
}

// GetWidth returns the width of the game world.
// This is used by game entities to determine boundaries and positioning.
func (g *Game) GetWidth() int {
	return g.world.GetWidth()
}

// GetHeight returns the height of the game world.
// This is used by game entities to determine boundaries and positioning.
func (g *Game) GetHeight() int {
	return g.world.GetHeight()
}

// SetWidth sets the width of the game world.
// This affects collision detection, entity spawning, and camera boundaries.
func (g *Game) SetWidth(width int) {
	g.world.SetWidth(width)
}

// SetHeight sets the height of the game world.
// This affects collision detection, entity spawning, and camera boundaries.
func (g *Game) SetHeight(height int) {
	g.world.SetHeight(height)
}

// ShouldMatchScreen returns true if the world dimensions should
// automatically match the screen dimensions.
// When true, the world will resize whenever the screen size changes.
// This is typically used for UI scenes or when the game world should
// fill the entire screen regardless of resolution.
func (g *Game) ShouldMatchScreen() bool {
	return g.world.ShouldMatchScreen()
}

// SetMatchScreen sets whether the world dimensions should
// automatically match the screen dimensions.
// Setting this to true makes the world dynamically resize with the screen,
// while setting it to false maintains a fixed world size regardless of screen size.
// Fixed world sizes are typically used for gameplay scenes with defined boundaries.
func (g *Game) SetMatchScreen(match bool) {
	g.world.SetMatchScreen(match)
}

// New creates a new game instance with initialized components.
// This function serves as the entry point for the game initialization process.
// It performs the following steps:
// 1. Creates and initializes all manager components (screen, scene, input)
// 2. Sets up the game world with the appropriate dimensions
// 3. Creates the player entity
// 4. Sets up the initial game scene (start screen)
// The initialization order is important to ensure all dependencies are properly set up.
func New() *Game {
	// Create screen manager first so it can be shared with other components
	// The screen manager handles window dimensions and scaling
	screenMgr := screen.New()

	// Create scene manager for handling different game states and transitions
	sceneManager := &scenes.SceneManager{}

	// Get initial screen dimensions to set up the game world
	width, height := screenMgr.GetWidth(), screenMgr.GetHeight()

	// Initialize the game struct with all required components
	g := &Game{
		sceneManager:    sceneManager,
		inputManager:    input.NewManager(),  // Create input manager for handling user input
		screenManager:   screenMgr,
		lastUpdateTime:  time.Now(),          // Initialize time tracking
		deltaTime:       1.0 / 60.0,          // Default to 60 FPS for first frame
		world:           ecs.NewBasicWorld(width, height), // Create ECS world with screen dimensions
	}

	// Connect the screen manager to the scene manager for proper rendering
	sceneManager.SetScreenManager(screenMgr)

	// Create the player entity and connect it to the game
	// This loads player sprites and initializes player state
	g.player = player.NewPlayer(g)

	// Set up the initial scene (start/title screen)
	// The game always begins at the start scene before transitioning to gameplay
	startScene := scenes.NewStartScene()
	g.sceneManager.GoToScene(startScene)

	return g
}

// Update updates the game state.
// It handles input processing and delegates to the current scene.
// This method is called by the Ebiten engine once per frame before drawing.
// The update process follows these steps:
// 1. Calculate and clamp delta time to ensure frame-rate independent updates
// 2. Update input state to process user interactions
// 3. Delegate scene-specific updates to the current active scene
//
// Delta time clamping prevents physics or animation issues during lag spikes
// or when the application loses focus (which could cause very large time steps).
func (g *Game) Update() error {
	// Calculate delta time (time since last frame) for frame-rate independent movement
	now := time.Now()
	g.deltaTime = now.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = now

	// Clamp deltaTime to avoid extreme values that could cause physics/animation issues
	if g.deltaTime > 0.1 {
		g.deltaTime = 0.1 // Cap at 100ms to prevent huge jumps during lag spikes
	} else if g.deltaTime < 0.001 {
		g.deltaTime = 0.001 // Minimum of 1ms to prevent division by zero issues
	}

	// Update input handlers to process keyboard, mouse, or touch events
	g.inputManager.Update()

	// Delegate to the current scene's update method with all necessary state
	// This allows each scene to handle its specific update logic
	return g.sceneManager.Update(g.inputManager, g.deltaTime, g.world)
}

// Draw renders the game to the screen.
// It delegates rendering to the current scene.
// This method is called by the Ebiten engine once per frame after updating.
// The drawing process is handled by the scene manager, which:
// 1. Determines which scene is currently active
// 2. Handles any scene transitions with fade effects
// 3. Calls the active scene's Draw method with the necessary state
//
// All rendering is done to the provided screen image, which is then presented
// to the display by the Ebiten engine.
func (g *Game) Draw(screen *ebiten.Image) {
	// Delegate drawing to the scene manager with all necessary game state
	// This allows each scene to handle its specific rendering logic
	// and enables smooth transitions between scenes
	g.sceneManager.Draw(screen, g.inputManager, g.deltaTime, g.world)
}

// Layout implements ebiten.Game's Layout method.
// It updates the game dimensions and input manager when the window is resized.
// This method is called by the Ebiten engine whenever the game window is resized
// or when the display properties change (e.g., when rotating a mobile device).
//
// The Layout method performs several important functions:
// 1. Calculates the logical screen dimensions based on the physical window size
// 2. Updates the input manager to correctly map input coordinates
// 3. Optionally resizes the game world to match the new screen dimensions
//
// The returned values (screenWidth, screenHeight) tell Ebiten the logical
// resolution to use for rendering, which may differ from the physical window size
// when using scaling or maintaining aspect ratios.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// Calculate logical screen dimensions using the screen manager
	// This handles scaling, aspect ratio, and other display considerations
	screenWidth, screenHeight = g.screenManager.CalculateLayout(outsideWidth, outsideHeight)

	// Update input manager with physical screen dimensions
	// This ensures input coordinates are correctly mapped to game coordinates
	g.inputManager.SetScreenDimensions(outsideWidth, outsideHeight)

	// Update world dimensions only if they should match the screen
	// This allows the game world to either:
	// - Dynamically resize with the screen (for UI-heavy scenes)
	// - Maintain a fixed size (for gameplay scenes with fixed boundaries)
	if g.world.ShouldMatchScreen() {
		g.world.SetWidth(screenWidth)
		g.world.SetHeight(screenHeight)
	}

	return screenWidth, screenHeight
}

// GoToScene changes to a new scene.
// It delegates to the scene manager's GoToScene method.
// This method triggers a scene transition with a fade effect between
// the current scene and the new scene. During the transition, both scenes
// are updated and drawn, with the new scene gradually fading in.
//
// Scene transitions are used to move between different game states such as:
// - From the start screen to the gameplay
// - From gameplay to a pause menu
// - From one level to another
// - From gameplay to game over screen
func (g *Game) GoToScene(scene scenes.Scene) {
	g.sceneManager.GoToScene(scene)
}

// SetDynamicResizing enables or disables dynamic screen resizing.
// When disabled, the game will maintain a fixed size regardless of window size.
//
// Dynamic resizing affects how the game appears when the window is resized:
// - When enabled: The game content scales to fill the available space while
//   maintaining aspect ratio, using the full window area
// - When disabled: The game maintains a fixed logical resolution and is
//   centered in the window, possibly with letterboxing or pillarboxing
//
// This setting is particularly important for supporting different screen sizes
// and orientations across desktop and mobile platforms.
func (g *Game) SetDynamicResizing(enabled bool) {
	g.screenManager.SetDynamicResizing(enabled)
}

// IsDynamicResizingEnabled returns whether dynamic resizing is enabled.
// This can be used by UI components to determine how to position elements
// based on whether the screen will dynamically resize or maintain fixed dimensions.
func (g *Game) IsDynamicResizingEnabled() bool {
	return g.screenManager.IsDynamicResizingEnabled()
}
