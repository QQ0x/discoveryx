package scenes

import (
	"discoveryx/internal/core/gameplay/player"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
)

type GameScene struct {
	player *player.Player
}

func NewGameScene(player *player.Player) *GameScene {
	return &GameScene{
		player: player,
	}
}

// Update handles the game logic for the game scene.
// It satisfies the Scene interface.
func (s *GameScene) Update(state *State) error {
	return s.player.Update(state.Input)
}

// Draw renders the game scene to the screen.
// It satisfies the Scene interface.
func (s *GameScene) Draw(screen *ebiten.Image) {
	// Clear screen with a dark color
	screen.Fill(color.RGBA{20, 20, 20, 255})

	// Draw the player
	s.player.Draw(screen)
}
