package physics

import (
	"discoveryx/internal/constants"
	"discoveryx/internal/utils/math"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
	"time"
)

// CircleCollider represents a circular collision area.
// It is used for collision detection between entities like players, enemies, and projectiles.
type CircleCollider struct {
	Position math.Vector // Center position of the circle
	Radius   float64     // Radius of the circle
}

// RectCollider represents a rectangular collision area.
// It is used for collision detection with walls and other rectangular objects.
type RectCollider struct {
	Position math.Vector // Center position of the rectangle
	Width    float64     // Width of the rectangle
	Height   float64     // Height of the rectangle
}

// AABBCollider represents an Axis-Aligned Bounding Box for collision detection.
// It is used for precise collision detection between the player and walls.
type AABBCollider struct {
	Position math.Vector // Center position of the box
	Width    float64     // Width of the box
	Height   float64     // Height of the box
}

// GetAABB returns the min and max points of the AABB.
func (c *AABBCollider) GetAABB() (math.Vector, math.Vector) {
	halfWidth := c.Width / 2
	halfHeight := c.Height / 2

	min := math.Vector{
		X: c.Position.X - halfWidth,
		Y: c.Position.Y - halfHeight,
	}

	max := math.Vector{
		X: c.Position.X + halfWidth,
		Y: c.Position.Y + halfHeight,
	}

	return min, max
}

// CheckCircleCollision determines if two circular colliders are intersecting.
// This is used for entity-entity collision detection (player-enemy, projectile-enemy, etc.).
//
// Parameters:
// - c1, c2: The two circular colliders to check for collision
//
// Returns:
// - bool: True if the colliders are intersecting, false otherwise
func CheckCircleCollision(c1, c2 CircleCollider) bool {
	// Calculate the distance between the centers of the circles
	dx := c1.Position.X - c2.Position.X
	dy := c1.Position.Y - c2.Position.Y
	distance := stdmath.Sqrt(dx*dx + dy*dy)

	// If the distance is less than the sum of the radii, the circles are colliding
	return distance < (c1.Radius + c2.Radius)
}

// CheckCircleRectCollision determines if a circular collider and a rectangular collider are intersecting.
// This is used for entity-wall collision detection.
//
// Parameters:
// - circle: The circular collider
// - rect: The rectangular collider
//
// Returns:
// - bool: True if the colliders are intersecting, false otherwise
// - math.Vector: The normal vector of the collision (direction to resolve the collision)
func CheckCircleRectCollision(circle CircleCollider, rect RectCollider) (bool, math.Vector) {
	// Find the closest point on the rectangle to the circle
	closestX := stdmath.Max(rect.Position.X-rect.Width/2, stdmath.Min(circle.Position.X, rect.Position.X+rect.Width/2))
	closestY := stdmath.Max(rect.Position.Y-rect.Height/2, stdmath.Min(circle.Position.Y, rect.Position.Y+rect.Height/2))

	// Calculate the distance between the circle's center and the closest point
	dx := closestX - circle.Position.X
	dy := closestY - circle.Position.Y
	distance := stdmath.Sqrt(dx*dx + dy*dy)

	// If the distance is less than the circle's radius, there is a collision
	if distance < circle.Radius {
		// Calculate the normal vector (direction to resolve the collision)
		var normal math.Vector
		if dx == 0 && dy == 0 {
			// Circle is inside the rectangle, use a default normal
			normal = math.Vector{X: 0, Y: -1}
		} else {
			// Normalize the vector from the closest point to the circle's center
			length := stdmath.Sqrt(dx*dx + dy*dy)
			normal = math.Vector{X: -dx / length, Y: -dy / length}
		}
		return true, normal
	}

	return false, math.Vector{}
}

// GetEntityCollider creates a circular collider for an entity based on its image and position.
// This is a utility function to simplify collision detection setup.
//
// Parameters:
// - position: The entity's position
// - img: The entity's image (used to determine size)
// - scale: The scale factor applied to the entity's image
//
// Returns:
// - CircleCollider: A circular collider for the entity
func GetEntityCollider(position math.Vector, img *ebiten.Image, scale float64) CircleCollider {
	// Default radius in case the image is nil
	defaultRadius := 15.0 * scale

	// Check if the image is nil to prevent panic
	if img == nil {
		return CircleCollider{
			Position: position,
			Radius:   defaultRadius,
		}
	}

	// Calculate the radius based on the image dimensions and scale
	width, height := float64(img.Bounds().Dx()), float64(img.Bounds().Dy())
	// Use the average of width and height, multiplied by scale, divided by 2 for radius
	radius := (width + height) / 4 * scale

	// Create and return the collider
	return CircleCollider{
		Position: position,
		Radius:   radius,
	}
}

// ResolveCollision calculates the new position after a collision to prevent overlap.
// This is used to push entities away from each other or from walls.
//
// Parameters:
// - position: The current position of the entity
// - normal: The normal vector of the collision (direction to resolve)
// - depth: The overlap depth to resolve
//
// Returns:
// - math.Vector: The new position after resolving the collision
func ResolveCollision(position math.Vector, normal math.Vector, depth float64) math.Vector {
	// Move the position along the normal vector by the overlap depth
	// Apply a correction factor to prevent oscillation (use 95% of the full correction)
	correctionFactor := 0.95
	return math.Vector{
		X: position.X + normal.X * depth * correctionFactor,
		Y: position.Y + normal.Y * depth * correctionFactor,
	}
}

// CheckAABBCollision determines if two AABB colliders are intersecting.
// This is used for precise collision detection between the player and walls.
//
// Parameters:
// - box1, box2: The two AABB colliders to check for collision
//
// Returns:
// - bool: True if the colliders are intersecting, false otherwise
// - math.Vector: The overlap vector (how much box1 overlaps box2)
func CheckAABBCollision(box1, box2 AABBCollider) (bool, math.Vector) {
	// Get the min and max points of both boxes
	min1, max1 := box1.GetAABB()
	min2, max2 := box2.GetAABB()

	// Check for intersection on both axes
	if max1.X < min2.X || min1.X > max2.X || max1.Y < min2.Y || min1.Y > max2.Y {
		return false, math.Vector{}
	}

	// Calculate overlap on each axis
	overlapX := stdmath.Min(max1.X, max2.X) - stdmath.Max(min1.X, min2.X)
	overlapY := stdmath.Min(max1.Y, max2.Y) - stdmath.Max(min1.Y, min2.Y)

	return true, math.Vector{X: overlapX, Y: overlapY}
}

// CheckAABBCollisionWithSeparation checks for collision between two AABB colliders
// and returns information about the collision, including the separation vector.
//
// Parameters:
// - box1, box2: The two AABB colliders to check for collision
//
// Returns:
// - bool: True if the colliders are intersecting, false otherwise
// - math.Vector: The separation vector (how to move box1 to resolve the collision)
// - bool: True if the collision is primarily on the X-axis, false if on the Y-axis
func CheckAABBCollisionWithSeparation(box1, box2 AABBCollider) (bool, math.Vector, bool) {
	// Get the min and max points of both boxes
	min1, max1 := box1.GetAABB()
	min2, max2 := box2.GetAABB()

	// Check for intersection on both axes
	if max1.X < min2.X || min1.X > max2.X || max1.Y < min2.Y || min1.Y > max2.Y {
		return false, math.Vector{}, false
	}

	// Calculate overlap on each axis
	overlapX := stdmath.Min(max1.X, max2.X) - stdmath.Max(min1.X, min2.X)
	overlapY := stdmath.Min(max1.Y, max2.Y) - stdmath.Max(min1.Y, min2.Y)

	// Determine which axis has the smaller overlap (for minimal separation)
	isXAxis := overlapX < overlapY

	// Calculate the separation vector
	var separationVector math.Vector
	if isXAxis {
		// X-axis separation
		if box1.Position.X < box2.Position.X {
			// box1 is to the left of box2
			separationVector = math.Vector{X: -overlapX, Y: 0}
		} else {
			// box1 is to the right of box2
			separationVector = math.Vector{X: overlapX, Y: 0}
		}
	} else {
		// Y-axis separation
		if box1.Position.Y < box2.Position.Y {
			// box1 is above box2
			separationVector = math.Vector{X: 0, Y: -overlapY}
		} else {
			// box1 is below box2
			separationVector = math.Vector{X: 0, Y: overlapY}
		}
	}

	return true, separationVector, isXAxis
}

// GetAABBColliderFromSprite creates an AABB collider for an entity based on its image and position.
//
// Parameters:
// - position: The entity's position
// - img: The entity's image (used to determine size)
// - scale: The scale factor applied to the entity's image
//
// Returns:
// - AABBCollider: An AABB collider for the entity
func GetAABBColliderFromSprite(position math.Vector, img *ebiten.Image, scale float64) AABBCollider {
	// Default size in case the image is nil
	defaultWidth := 30.0 * scale
	defaultHeight := 30.0 * scale

	// Check if the image is nil to prevent panic
	if img == nil {
		return AABBCollider{
			Position: position,
			Width:    defaultWidth,
			Height:   defaultHeight,
		}
	}

	// Calculate the size based on the image dimensions and scale
	width := float64(img.Bounds().Dx()) * scale
	height := float64(img.Bounds().Dy()) * scale

	// Create and return the collider
	return AABBCollider{
		Position: position,
		Width:    width,
		Height:   height,
	}
}

// CheckContinuousCircleCollision performs continuous collision detection for a moving circle.
// This is used to detect collisions for fast-moving objects like bullets that might
// otherwise pass through thin walls or enemies between frames.
//
// Parameters:
// - startPos: The starting position of the circle
// - endPos: The ending position of the circle
// - radius: The radius of the circle
// - rect: The rectangle to check for collision with
//
// Returns:
// - bool: True if a collision was detected, false otherwise
// - math.Vector: The point of collision (only valid if a collision was detected)
// - math.Vector: The normal vector of the collision (only valid if a collision was detected)
// - float64: The time of collision (0-1, where 0 is startPos and 1 is endPos)
func CheckContinuousCircleCollision(startPos, endPos math.Vector, radius float64, rect RectCollider) (bool, math.Vector, math.Vector, float64) {
	// Debug logging for detailed collision detection
	if constants.DebugCollisionDetailed {
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] CONTINUOUS COLLISION CHECK: Circle(start=%v, end=%v, radius=%v) vs Rect(pos=%v, size=%vx%v)\n",
			timestamp, startPos, endPos, radius, rect.Position, rect.Width, rect.Height)
	}

	// Calculate the movement vector
	moveX := endPos.X - startPos.X
	moveY := endPos.Y - startPos.Y

	// Calculate the expanded rectangle (expanded by the circle's radius)
	expandedRect := RectCollider{
		Position: rect.Position,
		Width: rect.Width + radius*2,
		Height: rect.Height + radius*2,
	}

	// Calculate the rectangle's bounds
	minX := expandedRect.Position.X - expandedRect.Width/2
	maxX := expandedRect.Position.X + expandedRect.Width/2
	minY := expandedRect.Position.Y - expandedRect.Height/2
	maxY := expandedRect.Position.Y + expandedRect.Height/2

	if constants.DebugCollisionDetailed {
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] EXPANDED RECT: MinX=%v, MaxX=%v, MinY=%v, MaxY=%v, Movement=(%v,%v)\n",
			timestamp, minX, maxX, minY, maxY, moveX, moveY)
	}

	// Calculate entry and exit times for X axis
	var txMin, txMax float64
	if moveX > 0 {
		txMin = (minX - startPos.X) / moveX
		txMax = (maxX - startPos.X) / moveX

		if constants.DebugCollisionDetailed {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] X-AXIS (moveX > 0): txMin=%v, txMax=%v\n", timestamp, txMin, txMax)
		}
	} else if moveX < 0 {
		txMin = (maxX - startPos.X) / moveX
		txMax = (minX - startPos.X) / moveX

		if constants.DebugCollisionDetailed {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] X-AXIS (moveX < 0): txMin=%v, txMax=%v\n", timestamp, txMin, txMax)
		}
	} else {
		// No movement in X direction
		if startPos.X < minX || startPos.X > maxX {
			if constants.DebugCollisionDetailed {
				timestamp := time.Now().Format("15:04:05.000")
				fmt.Printf("[%s] X-AXIS (moveX = 0): No collision possible (outside X bounds)\n", timestamp)
			}
			return false, math.Vector{}, math.Vector{}, 0
		}
		txMin = 0
		txMax = 1

		if constants.DebugCollisionDetailed {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] X-AXIS (moveX = 0): Inside X bounds, txMin=%v, txMax=%v\n", timestamp, txMin, txMax)
		}
	}

	// Calculate entry and exit times for Y axis
	var tyMin, tyMax float64
	if moveY > 0 {
		tyMin = (minY - startPos.Y) / moveY
		tyMax = (maxY - startPos.Y) / moveY

		if constants.DebugCollisionDetailed {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] Y-AXIS (moveY > 0): tyMin=%v, tyMax=%v\n", timestamp, tyMin, tyMax)
		}
	} else if moveY < 0 {
		tyMin = (maxY - startPos.Y) / moveY
		tyMax = (minY - startPos.Y) / moveY

		if constants.DebugCollisionDetailed {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] Y-AXIS (moveY < 0): tyMin=%v, tyMax=%v\n", timestamp, tyMin, tyMax)
		}
	} else {
		// No movement in Y direction
		if startPos.Y < minY || startPos.Y > maxY {
			if constants.DebugCollisionDetailed {
				timestamp := time.Now().Format("15:04:05.000")
				fmt.Printf("[%s] Y-AXIS (moveY = 0): No collision possible (outside Y bounds)\n", timestamp)
			}
			return false, math.Vector{}, math.Vector{}, 0
		}
		tyMin = 0
		tyMax = 1

		if constants.DebugCollisionDetailed {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] Y-AXIS (moveY = 0): Inside Y bounds, tyMin=%v, tyMax=%v\n", timestamp, tyMin, tyMax)
		}
	}

	// Find the latest entry time and earliest exit time
	tMin := stdmath.Max(txMin, tyMin)
	tMax := stdmath.Min(txMax, tyMax)

	if constants.DebugCollisionDetailed {
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] COLLISION TIMES: tMin=%v, tMax=%v\n", timestamp, tMin, tMax)
	}

	// Check if there's a collision
	if tMin > tMax || tMin > 1 || tMax < 0 {
		if constants.DebugCollisionDetailed {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] NO COLLISION: tMin > tMax (%v > %v) or tMin > 1 (%v > 1) or tMax < 0 (%v < 0)\n", 
				timestamp, tMin, tMax, tMin, tMax)
		}
		return false, math.Vector{}, math.Vector{}, 0
	}

	// Calculate the collision point
	collisionTime := stdmath.Max(0, tMin) // Clamp to 0 if already colliding
	collisionPoint := math.Vector{
		X: startPos.X + moveX*collisionTime,
		Y: startPos.Y + moveY*collisionTime,
	}

	// Calculate the collision normal
	var normal math.Vector
	if txMin > tyMin {
		// Collision on X axis
		if moveX > 0 {
			normal = math.Vector{X: -1, Y: 0} // Left face
			if constants.DebugCollisionDetailed {
				timestamp := time.Now().Format("15:04:05.000")
				fmt.Printf("[%s] COLLISION ON X-AXIS (LEFT FACE): txMin=%v > tyMin=%v, moveX=%v\n", 
					timestamp, txMin, tyMin, moveX)
			}
		} else {
			normal = math.Vector{X: 1, Y: 0} // Right face
			if constants.DebugCollisionDetailed {
				timestamp := time.Now().Format("15:04:05.000")
				fmt.Printf("[%s] COLLISION ON X-AXIS (RIGHT FACE): txMin=%v > tyMin=%v, moveX=%v\n", 
					timestamp, txMin, tyMin, moveX)
			}
		}
	} else {
		// Collision on Y axis
		if moveY > 0 {
			normal = math.Vector{X: 0, Y: -1} // Top face
			if constants.DebugCollisionDetailed {
				timestamp := time.Now().Format("15:04:05.000")
				fmt.Printf("[%s] COLLISION ON Y-AXIS (TOP FACE): txMin=%v <= tyMin=%v, moveY=%v\n", 
					timestamp, txMin, tyMin, moveY)
			}
		} else {
			normal = math.Vector{X: 0, Y: 1} // Bottom face
			if constants.DebugCollisionDetailed {
				timestamp := time.Now().Format("15:04:05.000")
				fmt.Printf("[%s] COLLISION ON Y-AXIS (BOTTOM FACE): txMin=%v <= tyMin=%v, moveY=%v\n", 
					timestamp, txMin, tyMin, moveY)
			}
		}
	}

	if constants.DebugCollisionDetailed {
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] COLLISION DETECTED: Time=%v, Point=%v, Normal=%v\n", 
			timestamp, collisionTime, collisionPoint, normal)
	}

	return true, collisionPoint, normal, collisionTime
}

// CheckContinuousCircleCircleCollision performs continuous collision detection between two moving circles.
// This is used to detect collisions between fast-moving objects like bullets and enemies.
//
// Parameters:
// - startPos1, endPos1: The starting and ending positions of the first circle
// - radius1: The radius of the first circle
// - startPos2, endPos2: The starting and ending positions of the second circle
// - radius2: The radius of the second circle
//
// Returns:
// - bool: True if a collision was detected, false otherwise
// - math.Vector: The point of collision for the first circle (only valid if a collision was detected)
// - math.Vector: The normal vector of the collision (only valid if a collision was detected)
// - float64: The time of collision (0-1, where 0 is startPos and 1 is endPos)
func CheckContinuousCircleCircleCollision(startPos1, endPos1 math.Vector, radius1 float64, startPos2, endPos2 math.Vector, radius2 float64) (bool, math.Vector, math.Vector, float64) {
	// Calculate the relative movement vector
	relStartPos := math.Vector{
		X: startPos1.X - startPos2.X,
		Y: startPos1.Y - startPos2.Y,
	}
	relEndPos := math.Vector{
		X: endPos1.X - endPos2.X,
		Y: endPos1.Y - endPos2.Y,
	}

	// Calculate the movement vector
	moveX := relEndPos.X - relStartPos.X
	moveY := relEndPos.Y - relStartPos.Y

	// Calculate the combined radius
	combinedRadius := radius1 + radius2

	// Solve the quadratic equation for the time of collision
	// (p + vt)^2 = r^2, where p is the relative position, v is the relative velocity, and r is the combined radius
	a := moveX*moveX + moveY*moveY
	b := 2 * (relStartPos.X*moveX + relStartPos.Y*moveY)
	c := relStartPos.X*relStartPos.X + relStartPos.Y*relStartPos.Y - combinedRadius*combinedRadius

	// Calculate the discriminant
	discriminant := b*b - 4*a*c

	// Check if there's a collision
	if discriminant < 0 || a == 0 {
		return false, math.Vector{}, math.Vector{}, 0
	}

	// Calculate the collision time
	t1 := (-b - stdmath.Sqrt(discriminant)) / (2 * a)
	t2 := (-b + stdmath.Sqrt(discriminant)) / (2 * a)

	// Use the earliest collision time that's within the movement range
	var collisionTime float64
	if t1 >= 0 && t1 <= 1 {
		collisionTime = t1
	} else if t2 >= 0 && t2 <= 1 {
		collisionTime = t2
	} else {
		return false, math.Vector{}, math.Vector{}, 0
	}

	// Calculate the collision point for the first circle
	collisionPoint := math.Vector{
		X: startPos1.X + (endPos1.X-startPos1.X)*collisionTime,
		Y: startPos1.Y + (endPos1.Y-startPos1.Y)*collisionTime,
	}

	// Calculate the collision point for the second circle
	collisionPoint2 := math.Vector{
		X: startPos2.X + (endPos2.X-startPos2.X)*collisionTime,
		Y: startPos2.Y + (endPos2.Y-startPos2.Y)*collisionTime,
	}

	// Calculate the collision normal (from circle2 to circle1)
	normal := math.Vector{
		X: collisionPoint.X - collisionPoint2.X,
		Y: collisionPoint.Y - collisionPoint2.Y,
	}

	// Normalize the normal vector
	length := stdmath.Sqrt(normal.X*normal.X + normal.Y*normal.Y)
	if length > 0 {
		normal.X /= length
		normal.Y /= length
	} else {
		// If the circles are exactly on top of each other, use a default normal
		normal = math.Vector{X: 1, Y: 0}
	}

	return true, collisionPoint, normal, collisionTime
}
