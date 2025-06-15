//go:build !mobile
// Non-mobile platforms only

package main

import (
	"discoveryx/internal/core/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("DiscoveryX")

	if err := ebiten.RunGame(game.New()); err != nil {
		panic(err)
	}
}
