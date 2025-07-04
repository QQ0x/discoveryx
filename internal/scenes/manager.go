package scenes

import (
	"discoveryx/internal/core/ecs"
	"discoveryx/internal/input"
	"discoveryx/internal/screen"
	"github.com/hajimehoshi/ebiten/v2"
)

const transitionMaxCount = 25

// Scene represents a game scene that can be updated and drawn.
// Each scene has access to the game state through the State parameter.
// The Scene interface is implemented by different game screens such as:
// - Start/title screen
// - Main gameplay screen
// - Pause menu
// - Game over screen
// - Loading screen
//
// This interface-based design allows for easy addition of new game screens
// and smooth transitions between them.
type Scene interface {
	// Update updates the scene state.
	// It returns an error if the update fails.
	// This method is called once per frame and should handle:
	// - Scene-specific logic
	// - Entity updates
	// - Input processing
	// - State transitions
	Update(state *State) error

	// Draw renders the scene to the screen.
	// It has access to the game state through the State parameter.
	// This method is responsible for:
	// - Rendering the scene background
	// - Drawing all visible entities
	// - Rendering UI elements
	// - Applying visual effects
	Draw(screen *ebiten.Image, state *State)
}

// State encapsulates all the game state needed by scenes.
// This struct is passed to both Update and Draw methods of scenes,
// providing access to all necessary game systems and state.
type State struct {
	SceneManager *SceneManager  // Reference to the scene manager for scene transitions
	Input        *input.Manager // Access to input state for handling user interactions
	DeltaTime    float64        // Time elapsed since last frame for frame-rate independent updates
	World        ecs.World      // The ECS world containing all game entities
}

// SceneManager handles scene transitions and manages the currently active scene.
// It provides a smooth fade transition effect between scenes and ensures
// proper initialization and cleanup of scene resources.
//
// The SceneManager is a core component of the game architecture, allowing
// different game states to be encapsulated in separate Scene implementations
// while providing a consistent interface for updating and rendering them.
type SceneManager struct {
	current         Scene             // The currently active scene
	next            Scene             // The scene being transitioned to (if a transition is in progress)
	transitionCount int               // Counter for tracking transition progress (0 = no transition)
	transitionFrom  *ebiten.Image     // Render target for the current scene during transitions
	transitionTo    *ebiten.Image     // Render target for the next scene during transitions
	screenManager   *screen.Manager   // Reference to the screen manager for dimension information
}

// Draw renders the current scene or a transition between scenes.
// This method is called by the Game's Draw method each frame and handles:
// 1. Creating a state object with all necessary game state
// 2. Drawing the current scene directly if no transition is in progress
// 3. Managing the transition effect between scenes if a transition is active
//
// During transitions, both the current and next scenes are rendered to separate
// images, and then composited with alpha blending to create a smooth fade effect.
func (s *SceneManager) Draw(r *ebiten.Image, inputManager *input.Manager, deltaTime float64, world ecs.World) {
	// Create state object with all necessary game state for the scene to use
	state := &State{
		SceneManager: s,
		Input:        inputManager,
		DeltaTime:    deltaTime,
		World:        world,
	}

	// If no transition is in progress, simply draw the current scene directly
	if s.transitionCount == 0 {
		s.current.Draw(r, state)
		return
	}

	// A transition is in progress, so we need to handle the fade effect
	// First, ensure transition images are properly initialized with the correct size
	width, height := r.Size()
	s.ensureTransitionImages(width, height)

	// Render the current scene to its transition image
	s.transitionFrom.Clear()
	s.current.Draw(s.transitionFrom, state)

	// Render the next scene to its transition image
	s.transitionTo.Clear()
	s.next.Draw(s.transitionTo, state)

	// Draw the current scene at full opacity
	r.DrawImage(s.transitionFrom, nil)

	// Calculate alpha for the next scene based on transition progress
	// As transitionCount decreases, alpha increases (fade in)
	alpha := 1 - float32(s.transitionCount)/float32(transitionMaxCount)
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)

	// Draw the next scene with calculated alpha on top of the current scene
	r.DrawImage(s.transitionTo, op)
}

// ensureTransitionImages makes sure transition images are initialized with the correct size.
// This method handles both initial creation and resizing of transition images
// to match the current screen dimensions. It's called during scene transitions
// to ensure the transition effect renders correctly at the current resolution.
//
// The method performs several important functions:
// 1. Initializes the screen manager if needed
// 2. Detects changes in screen dimensions
// 3. Updates the screen manager with new dimensions
// 4. Creates or resizes transition images as needed
//
// Proper management of these images is important for both visual quality
// and memory efficiency, especially on mobile devices with limited resources.
func (s *SceneManager) ensureTransitionImages(width, height int) {
	// Initialize screen manager if it doesn't exist
	// This is a fallback in case SetScreenManager wasn't called
	if s.screenManager == nil {
		s.screenManager = screen.New()
	}

	// Check if dimensions have changed significantly enough to require resizing
	// This prevents unnecessary image recreations when dimensions change slightly
	currentWidth := s.screenManager.GetWidth()
	currentHeight := s.screenManager.GetHeight()
	needsResize := width != currentWidth || height != currentHeight

	// Update screen manager dimensions to match the current screen size
	if needsResize {
		s.screenManager.SetDimensions(width, height)
	}

	// Handle the "from" transition image (current scene)
	// Either create it if it doesn't exist or resize it if dimensions changed
	if s.transitionFrom == nil {
		s.transitionFrom = ebiten.NewImage(width, height)
	} else if needsResize {
		// Dispose the old image to free GPU memory before creating a new one
		s.transitionFrom.Dispose()
		s.transitionFrom = ebiten.NewImage(width, height)
	}

	// Handle the "to" transition image (next scene)
	// Either create it if it doesn't exist or resize it if dimensions changed
	if s.transitionTo == nil {
		s.transitionTo = ebiten.NewImage(width, height)
	} else if needsResize {
		// Dispose the old image to free GPU memory before creating a new one
		s.transitionTo.Dispose()
		s.transitionTo = ebiten.NewImage(width, height)
	}
}

// Update updates the current scene or advances a scene transition.
// This method is called by the Game's Update method each frame and handles:
// 1. Updating the current scene if no transition is in progress
// 2. Advancing the transition progress if a transition is active
// 3. Finalizing the transition when complete by making the next scene current
//
// During transitions, the current scene's Update method is not called,
// allowing for a clean handoff between scenes without interference.
func (s *SceneManager) Update(inputManager *input.Manager, deltaTime float64, world ecs.World) error {
	// If no transition is in progress, update the current scene
	if s.transitionCount == 0 {
		// Create state object with all necessary game state for the scene to use
		return s.current.Update(&State{
			SceneManager: s,
			Input:        inputManager,
			DeltaTime:    deltaTime,
			World:        world,
		})
	}

	// A transition is in progress, decrement the transition counter
	s.transitionCount--

	// If the transition is still in progress, do nothing else
	if s.transitionCount > 0 {
		return nil
	}

	// Transition is complete, make the next scene the current scene
	s.current = s.next
	s.next = nil

	// Clean up transition images to free memory and prepare for next transition
	s.Cleanup()

	return nil
}

// Cleanup clears transition images but doesn't dispose them for reuse.
// This is called after a transition completes to prepare for the next transition.
// The images are kept in memory but cleared to avoid unnecessary allocations,
// which improves performance when multiple scene transitions occur.
func (s *SceneManager) Cleanup() {
	// Instead of disposing, we just clear the images for reuse
	// This is more efficient than creating new images for each transition
	if s.transitionFrom != nil {
		s.transitionFrom.Clear()
	}

	if s.transitionTo != nil {
		s.transitionTo.Clear()
	}
}

// FinalCleanup should be called when the scene manager is no longer needed.
// This method properly disposes of all resources to prevent memory leaks.
// It should be called when shutting down the game or when the scene manager
// will not be used again.
func (s *SceneManager) FinalCleanup() {
	// Properly dispose of transition images to free GPU memory
	if s.transitionFrom != nil {
		s.transitionFrom.Dispose()
		s.transitionFrom = nil
	}

	if s.transitionTo != nil {
		s.transitionTo.Dispose()
		s.transitionTo = nil
	}
}

// SetScreenManager sets the screen manager to be used by this scene manager.
// The screen manager provides information about screen dimensions that is
// needed for properly sizing transition images and other rendering operations.
func (s *SceneManager) SetScreenManager(manager *screen.Manager) {
	s.screenManager = manager
}

// GoToScene changes to a new scene with a smooth transition effect.
// If this is the first scene being set (no current scene), it becomes
// the current scene immediately without a transition.
// Otherwise, a fade transition is initiated between the current scene
// and the new scene.
func (s *SceneManager) GoToScene(scene Scene) {
	// If there's no current scene, set the new scene directly without transition
	if s.current == nil {
		s.current = scene
	} else {
		// Store the next scene and start the transition
		s.next = scene
		s.transitionCount = transitionMaxCount

		// Pre-allocate transition images if they don't exist yet
		// We use a default size initially, they'll be resized if needed in Draw
		if s.transitionFrom == nil || s.transitionTo == nil {
			// Use standard resolution as default size for transition images
			defaultWidth, defaultHeight := 640, 480
			s.ensureTransitionImages(defaultWidth, defaultHeight)
		}
	}
}
