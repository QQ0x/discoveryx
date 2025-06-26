package mobile

import (
	"github.com/hajimehoshi/ebiten/v2/mobile"

	"example.com/go_astroids/internal/game"
)

func init() {
	inogame := game.New()
	if inogame == nil {
		panic("Failed to initialize game")
	}
	mobile.SetGame(inogame)
}

// Dummy is a dummy exported function.
//
// gomobile doesn't compile a package that doesn't include any exported function.
// Dummy forces gomobile to compile this package.
func Dummy() {}
