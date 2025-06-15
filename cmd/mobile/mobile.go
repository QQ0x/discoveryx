package mobile

import (
	"github.com/hajimehoshi/ebiten/v2/mobile"

	"discoveryx/internal/core/game"
)

func init() {
	game := game.New()
	if game == nil {
		panic("Failed to initialize game")
	}
	mobile.SetGame(game)
}

// Dummy is a dummy exported function.
//
// gomobile doesn't compile a package that doesn't include any exported function.
// Dummy forces gomobile to compile this package.
func Dummy() {}
