package ecs

import "github.com/hajimehoshi/ebiten/v2"

// Entity represents a game entity with update and draw capabilities
type Entity interface {
	Update() error
	Draw(screen *ebiten.Image)
}
