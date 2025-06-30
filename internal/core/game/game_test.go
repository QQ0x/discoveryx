package game

import (
	"discoveryx/internal/scenes"
	"testing"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
)

// TestNewGame tests that a new game instance is created correctly
func TestNewGame(t *testing.T) {
	g := New()

	// Check that the game instance is not nil
	if g == nil {
		t.Fatal("New() returned nil")
	}

	// Check that the game dimensions are set correctly
	if g.GetWidth() <= 0 || g.GetHeight() <= 0 {
		t.Errorf("Game dimensions not set correctly: width=%d, height=%d", g.GetWidth(), g.GetHeight())
	}

	// Check that the scene manager is initialized
	if g.sceneManager == nil {
		t.Fatal("Scene manager not initialized")
	}

	// Check that the input manager is initialized
	if g.inputManager == nil {
		t.Fatal("Input manager not initialized")
	}

	// Check that the player is initialized
	if g.player == nil {
		t.Fatal("Player not initialized")
	}
}

// MockScene is a simple mock implementation of the Scene interface for testing
type MockScene struct {
	updateCalled bool
	drawCalled   bool
}

func (m *MockScene) Update(state *scenes.State) error {
	m.updateCalled = true
	return nil
}

func (m *MockScene) Draw(screen *ebiten.Image) {
	m.drawCalled = true
	screen.Fill(color.RGBA{0, 0, 0, 255})
}

// TestGoToScene tests that the GoToScene method works correctly
func TestGoToScene(t *testing.T) {
	g := New()
	mockScene := &MockScene{}

	// Change to the mock scene
	g.GoToScene(mockScene)

	// Update the game to trigger the scene transition
	err := g.Update()
	if err != nil {
		t.Fatalf("Update() returned an error: %v", err)
	}

	// Create a dummy screen for drawing
	screen := ebiten.NewImage(640, 480)

	// Draw the game to trigger the scene's Draw method
	g.Draw(screen)

	// Check that the mock scene's methods were called
	if !mockScene.updateCalled {
		t.Error("Mock scene's Update method was not called")
	}

	if !mockScene.drawCalled {
		t.Error("Mock scene's Draw method was not called")
	}
}
