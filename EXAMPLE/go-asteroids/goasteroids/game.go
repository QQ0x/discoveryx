package goasteroids

import "github.com/hajimehoshi/ebiten/v2"

// Game is the type for the overall game. It holds a scene manager, used to change scenes,
// and a stub input type (required to use this as a parameter in ebiten.RunGame) which is
// used to pass keyboard input to scenes.
type Game struct {
	sceneManager *SceneManager
	input Input
}

// Input is a stub type so that we can use Game as an ebiten.Game interface in the call to ebiten.RunGame.
type Input struct{}

// Update is required to satisfy the interface requirements for calling ebiten.RunGame. It's a stub, but we need it.
func (i *Input) Update() {}

// Update manages scenes, and updates input (which is sent to each scene).
func (g *Game) Update() error {
	if g.sceneManager == nil {
		g.sceneManager = &SceneManager{}
		g.sceneManager.GoToScene(NewGameScene())
	}

	g.input.Update()
	if err := g.sceneManager.Update(&g.input); err != nil {
		return err
	}
	return nil
}

// Draw draws the game using the current scene.
func (g *Game) Draw(screen *ebiten.Image) {
	g.sceneManager.Draw(screen)
}

// Layout is required in order to satisfy the ebiten.Game interface.
func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return ScreenWidth, ScreenHeight
}