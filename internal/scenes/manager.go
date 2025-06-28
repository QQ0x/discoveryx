package scenes

import (
	"discoveryx/internal/input"
	"github.com/hajimehoshi/ebiten/v2"
)

const transitionMaxCount = 25

type Scene interface {
	Update(state *State) error
	Draw(screen *ebiten.Image)
}

type State struct {
	SceneManager *SceneManager
	Input *input.Manager
}

type SceneManager struct{
	current Scene
	next Scene
	transitionCount int
	transitionFrom *ebiten.Image
	transitionTo *ebiten.Image
}

func (s *SceneManager) Draw(r *ebiten.Image) {
	if s.transitionCount == 0 {
		s.current.Draw(r)
		return
	}

	// Create transition images on demand with the correct size
	width, height := r.Size()
	if s.transitionFrom == nil {
		s.transitionFrom = ebiten.NewImage(width, height)
	} else {
		fromWidth, fromHeight := s.transitionFrom.Size()
		if fromWidth != width || fromHeight != height {
			s.transitionFrom.Dispose()
			s.transitionFrom = ebiten.NewImage(width, height)
		}
	}

	if s.transitionTo == nil {
		s.transitionTo = ebiten.NewImage(width, height)
	} else {
		toWidth, toHeight := s.transitionTo.Size()
		if toWidth != width || toHeight != height {
			s.transitionTo.Dispose()
			s.transitionTo = ebiten.NewImage(width, height)
		}
	}

	s.transitionFrom.Clear()
	s.current.Draw(s.transitionFrom)

	s.transitionTo.Clear()
	s.next.Draw(s.transitionTo)

	// Draw directly without redundant operations
	r.DrawImage(s.transitionFrom, nil)

	alpha := 1 - float32(s.transitionCount)/float32(transitionMaxCount)
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)
	r.DrawImage(s.transitionTo, op)
}

func (s *SceneManager) Update(inputManager *input.Manager) error {
	if s.transitionCount == 0 {
		return s.current.Update(&State{
			SceneManager: s,
			Input: inputManager,
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

// Cleanup disposes of transition images when they're no longer needed
func (s *SceneManager) Cleanup() {
	if s.transitionFrom != nil {
		s.transitionFrom.Dispose()
		s.transitionFrom = nil
	}

	if s.transitionTo != nil {
		s.transitionTo.Dispose()
		s.transitionTo = nil
	}
}

func (s *SceneManager) GoToScene(scene Scene) {
	if s.current == nil {
		s.current = scene
	} else {
		s.next = scene
		s.transitionCount = transitionMaxCount
	}
}
