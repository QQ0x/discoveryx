package goasteroids

import "github.com/hajimehoshi/ebiten/v2"

var (
	transitionFrom = ebiten.NewImage(ScreenWidth, ScreenHeight)
	transitionTo   = ebiten.NewImage(ScreenWidth, ScreenHeight)
)

const transitionMaxCount = 25

// Scene is the interface for all scenes. In order to be a scene, we have to implement all the functions
// for this interface.
type Scene interface {
	Update(state *State) error
	Draw(screen *ebiten.Image)
}

// State is the type for game state. All we need to keep track of is the Scene (with SceneManager)
// and Input (so we can get the keys pressed on the keyboard).
type State struct {
	SceneManager *SceneManager
	Input        *Input
}

// SceneManager is the type used to manage scenes. We keep track of what scene we are on, and when
// going to a new scene, what that scene is (so we can fade things in & out nicely).
type SceneManager struct {
	current         Scene
	next            Scene
	transitionCount int
}

// Draw draws the scene. Note that it transitions between scenes using transitionCount (which is set to
// transitionMaxCount initially when moving between scenes).
func (s *SceneManager) Draw(r *ebiten.Image) {
	if s.transitionCount == 0 {
		s.current.Draw(r)
		return
	}

	transitionFrom.Clear()
	s.current.Draw(transitionFrom)

	transitionTo.Clear()
	s.next.Draw(transitionTo)

	r.DrawImage(transitionFrom, nil)

	alpha := 1 - float32(s.transitionCount)/float32(transitionMaxCount)
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(alpha)
	r.DrawImage(transitionTo, op)
}

// Update updates the scene for the next draw.
func (s *SceneManager) Update(_ *Input) error {
	if s.transitionCount == 0 {
		return s.current.Update(&State{
			SceneManager: s,
		})
	}

	s.transitionCount--
	if s.transitionCount > 0 {
		return nil
	}

	s.current = s.next
	s.next = nil
	return nil
}

// GoToScene takes us to another scene.
func (s *SceneManager) GoToScene(scene Scene) {
	if s.current == nil {
		s.current = scene
	} else {
		s.next = scene
		s.transitionCount = transitionMaxCount
	}
}
