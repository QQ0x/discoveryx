package physics

import (
	"discoveryx/internal/core/ecs"
	"discoveryx/internal/utils/math"
)

// ClampToWorld clamps a position vector so that it stays within the world
// boundaries. It returns the clamped position and whether a clamp was
// necessary (i.e., a collision with the boundary occurred).
func ClampToWorld(pos math.Vector, world ecs.World) (math.Vector, bool) {
	halfW := float64(world.GetWidth()) / 2
	halfH := float64(world.GetHeight()) / 2
	collided := false

	if pos.X < -halfW {
		pos.X = -halfW
		collided = true
	} else if pos.X > halfW {
		pos.X = halfW
		collided = true
	}

	if pos.Y < -halfH {
		pos.Y = -halfH
		collided = true
	} else if pos.Y > halfH {
		pos.Y = halfH
		collided = true
	}

	return pos, collided
}
