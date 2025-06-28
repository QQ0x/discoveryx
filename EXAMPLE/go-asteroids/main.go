package main

import (
	"go-asteroids/goasteroids"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {

	ebiten.SetWindowTitle("Go Asteroids")
	ebiten.SetWindowSize(goasteroids.ScreenWidth, goasteroids.ScreenHeight)

	err := ebiten.RunGame(&goasteroids.Game{})
	if err != nil {
		panic(err)
	}
}