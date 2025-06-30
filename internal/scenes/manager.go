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
type Scene interface {
	// Update updates the scene state.
	// It returns an error if the update fails.
	Update(state *State) error

	// Draw renders the scene to the screen.
	// It has access to the game state through the State parameter.
	Draw(screen *ebiten.Image, state *State)
}

type State struct {
	SceneManager *SceneManager
	Input        *input.Manager
	DeltaTime    float64
	World        ecs.World
}

type SceneManager struct {
	current         Scene
	next            Scene
	transitionCount int
	transitionFrom  *ebiten.Image
	transitionTo    *ebiten.Image
	screenManager   *screen.Manager
}

func (s *SceneManager) Draw(r *ebiten.Image, inputManager *input.Manager, deltaTime float64, world ecs.World) {
	// Create state for drawing
	state := &State{
		SceneManager: s,
		Input:        inputManager,
		DeltaTime:    deltaTime,
		World:        world,
	}

	if s.transitionCount == 0 {
		s.current.Draw(r, state)
		return
	}

	// Check if we need to initialize or resize transition images
	width, height := r.Size()
	s.ensureTransitionImages(width, height)

	s.transitionFrom.Clear()
	s.current.Draw(s.transitionFrom, state)

	s.transitionTo.Clear()
	s.next.Draw(s.transitionTo, state)

	// Draw directly without redundant operations
	r.DrawImage(s.transitionFrom, nil)

	alpha := 1 - float32(s.transitionCount)/float32(transitionMaxCount)
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)
	r.DrawImage(s.transitionTo, op)
}

// ensureTransitionImages makes sure transition images are initialized with the correct size
func (s *SceneManager) ensureTransitionImages(width, height int) {
	// Initialize screen manager if it doesn't exist
	if s.screenManager == nil {
		s.screenManager = screen.New()
	}

	// Check if dimensions have changed significantly
	currentWidth := s.screenManager.GetWidth()
	currentHeight := s.screenManager.GetHeight()
	needsResize := width != currentWidth || height != currentHeight

	// Update screen manager dimensions
	if needsResize {
		s.screenManager.SetDimensions(width, height)
	}

	// Initialize images if they don't exist
	if s.transitionFrom == nil {
		s.transitionFrom = ebiten.NewImage(width, height)
	} else if needsResize {
		s.transitionFrom.Dispose()
		s.transitionFrom = ebiten.NewImage(width, height)
	}

	if s.transitionTo == nil {
		s.transitionTo = ebiten.NewImage(width, height)
	} else if needsResize {
		s.transitionTo.Dispose()
		s.transitionTo = ebiten.NewImage(width, height)
	}
}

func (s *SceneManager) Update(inputManager *input.Manager, deltaTime float64, world ecs.World) error {
	if s.transitionCount == 0 {
		return s.current.Update(&State{
			SceneManager: s,
			Input:        inputManager,
			DeltaTime:    deltaTime,
			World:        world,
		})
	}

	s.transitionCount--
	if s.transitionCount > 0 {
		return nil
	}

	s.current = s.next
	s.next = nil

	// Clean up transition images when transition is complete
	s.Cleanup()

	return nil
}

// Cleanup clears transition images but doesn't dispose them for reuse
func (s *SceneManager) Cleanup() {
	// Instead of disposing, we just clear the images for reuse
	if s.transitionFrom != nil {
		s.transitionFrom.Clear()
	}

	if s.transitionTo != nil {
		s.transitionTo.Clear()
	}
}

// FinalCleanup should be called when the scene manager is no longer needed
func (s *SceneManager) FinalCleanup() {
	if s.transitionFrom != nil {
		s.transitionFrom.Dispose()
		s.transitionFrom = nil
	}

	if s.transitionTo != nil {
		s.transitionTo.Dispose()
		s.transitionTo = nil
	}
}

// SetScreenManager sets the screen manager to be used by this scene manager
func (s *SceneManager) SetScreenManager(manager *screen.Manager) {
	s.screenManager = manager
}

func (s *SceneManager) GoToScene(scene Scene) {
	if s.current == nil {
		s.current = scene
	} else {
		s.next = scene
		s.transitionCount = transitionMaxCount

		// Pre-allocate transition images if they don't exist yet
		// We use a default size initially, they'll be resized if needed in Draw
		if s.transitionFrom == nil || s.transitionTo == nil {
			defaultWidth, defaultHeight := 640, 480
			s.ensureTransitionImages(defaultWidth, defaultHeight)
		}
	}
}
