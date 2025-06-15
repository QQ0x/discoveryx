package ecs

// World represents the game world with dimensions
type World interface {
	GetWidth() int
	GetHeight() int
}
