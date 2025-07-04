package physics

import "discoveryx/internal/utils/math"

// CircleCollision checks for overlap between two circular colliders defined by
// their centre positions and radii.
func CircleCollision(aPos, bPos math.Vector, aRadius, bRadius float64) bool {
	dx := aPos.X - bPos.X
	dy := aPos.Y - bPos.Y
	r := aRadius + bRadius
	return dx*dx+dy*dy <= r*r
}
