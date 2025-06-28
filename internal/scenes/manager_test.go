package scenes

import (
	"discoveryx/internal/input"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"testing"
)

// MockScene is a simple mock implementation of the Scene interface for testing
type MockScene struct {
	updateCalled bool
	drawCalled   bool
	name         string
}

// NewMockScene creates a new mock scene with the given name
func NewMockScene(name string) *MockScene {
	return &MockScene{
		name: name,
	}
}

// Update implements the Scene interface
func (m *MockScene) Update(state *State) error {
	m.updateCalled = true
	return nil
}

// Draw implements the Scene interface
func (m *MockScene) Draw(screen *ebiten.Image) {
	m.drawCalled = true
	screen.Fill(color.RGBA{0, 0, 0, 255})
}

// TestSceneManagerInitialization tests that a new SceneManager is initialized correctly
func TestSceneManagerInitialization(t *testing.T) {
	sm := &SceneManager{}

	// Check that the scene manager is not nil
	if sm == nil {
		t.Fatal("SceneManager is nil")
	}

	// Check that the current scene is nil
	if sm.current != nil {
		t.Errorf("Initial current scene is not nil: %v", sm.current)
	}

	// Check that the next scene is nil
	if sm.next != nil {
		t.Errorf("Initial next scene is not nil: %v", sm.next)
	}

	// Check that the transition count is 0
	if sm.transitionCount != 0 {
		t.Errorf("Initial transition count is not 0: %d", sm.transitionCount)
	}
}

// TestGoToScene tests that the GoToScene method works correctly
func TestGoToScene(t *testing.T) {
	sm := &SceneManager{}
	scene1 := NewMockScene("Scene1")
	scene2 := NewMockScene("Scene2")

	// Set the first scene
	sm.GoToScene(scene1)

	// Check that the current scene is set correctly
	if sm.current != scene1 {
		t.Errorf("Current scene not set correctly: expected %v, got %v", scene1, sm.current)
	}

	// Check that the next scene is nil
	if sm.next != nil {
		t.Errorf("Next scene is not nil: %v", sm.next)
	}

	// Set the second scene
	sm.GoToScene(scene2)

	// Check that the current scene is still the first scene
	if sm.current != scene1 {
		t.Errorf("Current scene changed unexpectedly: expected %v, got %v", scene1, sm.current)
	}

	// Check that the next scene is set correctly
	if sm.next != scene2 {
		t.Errorf("Next scene not set correctly: expected %v, got %v", scene2, sm.next)
	}

	// Check that the transition count is set correctly
	if sm.transitionCount != transitionMaxCount {
		t.Errorf("Transition count not set correctly: expected %d, got %d", transitionMaxCount, sm.transitionCount)
	}
}

// TestSceneTransition tests that scene transitions work correctly
func TestSceneTransition(t *testing.T) {
	sm := &SceneManager{}
	scene1 := NewMockScene("Scene1")
	scene2 := NewMockScene("Scene2")

	// Create a real input manager
	inputManager := input.NewManager()

	// Set the first scene
	sm.GoToScene(scene1)

	// Set the second scene to trigger a transition
	sm.GoToScene(scene2)

	// Update the scene manager until the transition is complete
	for i := 0; i < transitionMaxCount; i++ {
		err := sm.Update(inputManager)
		if err != nil {
			t.Fatalf("Update returned an error: %v", err)
		}
	}

	// Check that the current scene is now the second scene
	if sm.current != scene2 {
		t.Errorf("Current scene not updated after transition: expected %v, got %v", scene2, sm.current)
	}

	// Check that the next scene is nil
	if sm.next != nil {
		t.Errorf("Next scene not reset after transition: %v", sm.next)
	}

	// Check that the transition count is 0
	if sm.transitionCount != 0 {
		t.Errorf("Transition count not reset after transition: %d", sm.transitionCount)
	}
}
