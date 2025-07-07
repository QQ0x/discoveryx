// Package physics provides collision detection and physics simulation for the game.
package physics

import (
	"discoveryx/internal/core/physics/collisions"
	"discoveryx/internal/utils/math"
	// "errors"
	"fmt"
	stdmath "math"
)

// EbitenCollisionSystem is an implementation of the CollisionSystem interface
// that uses the ebiten-collisions library for collision detection.
type EbitenCollisionSystem struct {
	// The underlying collision space from the ebiten-collisions library
	space *collisions.Space

	// Map of shape IDs to shapes for quick lookup
	shapes map[int]Shape

	// Next available shape ID
	nextID int

	// Map of shape IDs to collisions.Object for quick lookup
	objects map[int]collisions.Object

	// Spatial partitioning cell size
	cellSize float64
}

// NewEbitenCollisionSystem creates a new collision system using the ebiten-collisions library.
// The cellSize parameter determines the size of the spatial partitioning cells.
func NewEbitenCollisionSystem(cellSize float64) *EbitenCollisionSystem {
	return &EbitenCollisionSystem{
		space:    collisions.NewSpace(cellSize),
		shapes:   make(map[int]Shape),
		objects:  make(map[int]collisions.Object),
		nextID:   1,
		cellSize: cellSize,
	}
}

// AddShape adds a shape to the collision system and returns its ID.
func (ecs *EbitenCollisionSystem) AddShape(shape Shape) int {
	id := ecs.nextID
	ecs.nextID++

	// Create a collisions.Object based on the shape type
	var obj collisions.Object

	switch shape.GetType() {
	case ShapeTypeCircle:
		circleShape, ok := shape.(*CircleShape)
		if !ok {
			panic("Shape with type ShapeTypeCircle is not a *CircleShape")
		}

		obj = collisions.NewCircle(
			circleShape.Position.X,
			circleShape.Position.Y,
			circleShape.Radius,
		)

	case ShapeTypeAABB:
		aabbShape, ok := shape.(*AABBShape)
		if !ok {
			panic("Shape with type ShapeTypeAABB is not a *AABBShape")
		}

		obj = collisions.NewRectangle(
			aabbShape.Position.X-aabbShape.Width/2,
			aabbShape.Position.Y-aabbShape.Height/2,
			aabbShape.Width,
			aabbShape.Height,
		)

	default:
		panic(fmt.Sprintf("Unsupported shape type: %v", shape.GetType()))
	}

	// Add the object to the space
	ecs.space.Add(obj)

	// Store the shape and object for later lookup
	ecs.shapes[id] = shape
	ecs.objects[id] = obj

	return id
}

// RemoveShape removes a shape from the collision system.
func (ecs *EbitenCollisionSystem) RemoveShape(id int) {
	obj, exists := ecs.objects[id]
	if !exists {
		return
	}

	// Remove the object from the space
	ecs.space.Remove(obj)

	// Remove the shape and object from our maps
	delete(ecs.shapes, id)
	delete(ecs.objects, id)
}

// UpdateShape updates a shape in the collision system.
func (ecs *EbitenCollisionSystem) UpdateShape(id int, shape Shape) {
	// Get the existing object
	obj, exists := ecs.objects[id]
	if !exists {
		return
	}

	// Update the object based on the shape type
	switch shape.GetType() {
	case ShapeTypeCircle:
		circleShape, ok := shape.(*CircleShape)
		if !ok {
			panic("Shape with type ShapeTypeCircle is not a *CircleShape")
		}

		circle, ok := obj.(*collisions.Circle)
		if !ok {
			panic("Object is not a *collisions.Circle")
		}

		circle.X = circleShape.Position.X
		circle.Y = circleShape.Position.Y
		circle.Radius = circleShape.Radius

	case ShapeTypeAABB:
		aabbShape, ok := shape.(*AABBShape)
		if !ok {
			panic("Shape with type ShapeTypeAABB is not a *AABBShape")
		}

		rect, ok := obj.(*collisions.Rectangle)
		if !ok {
			panic("Object is not a *collisions.Rectangle")
		}

		rect.X = aabbShape.Position.X - aabbShape.Width/2
		rect.Y = aabbShape.Position.Y - aabbShape.Height/2
		rect.W = aabbShape.Width
		rect.H = aabbShape.Height

	default:
		panic(fmt.Sprintf("Unsupported shape type: %v", shape.GetType()))
	}

	// Update our shape map
	ecs.shapes[id] = shape
}

// Resolve checks for collisions and returns a list of collisions.
func (ecs *EbitenCollisionSystem) Resolve(filter CollisionFilter) ([]Collision, error) {
	// Update the space to check for collisions
	ecs.space.Update()

	// Get all collisions from the space
	var collisions []Collision

	// Check each object against all other objects
	for idA, objA := range ecs.objects {
		shapeA := ecs.shapes[idA]

		// Get all objects that collide with this object
		for _, other := range ecs.space.GetCollisions(objA) {
			// Find the shape of the other object
			var shapeB Shape

			for id, obj := range ecs.objects {
				if obj == other {
					shapeB = ecs.shapes[id]
					break
				}
			}

			// Skip if we couldn't find the other object
			if shapeB == nil {
				continue
			}

			// Apply the filter if provided
			if filter != nil && !filter(shapeA, shapeB) {
				continue
			}

			// Create a collision object
			collision := Collision{
				ShapeA: shapeA,
				ShapeB: shapeB,
			}

			// Calculate the collision normal and depth
			posA := shapeA.GetPosition()
			posB := shapeB.GetPosition()

			// Calculate the vector from B to A
			dx := posA.X - posB.X
			dy := posA.Y - posB.Y

			// Normalize the vector
			length := stdmath.Sqrt(dx*dx + dy*dy)
			if length > 0 {
				dx /= length
				dy /= length
			} else {
				// If the shapes are at the same position, use a default normal
				dx = 0
				dy = -1
			}

			collision.Normal = math.Vector{X: dx, Y: dy}

			// Calculate the collision depth based on the shape types
			switch {
			case shapeA.GetType() == ShapeTypeCircle && shapeB.GetType() == ShapeTypeCircle:
				// Circle-circle collision
				circleA := shapeA.(*CircleShape)
				circleB := shapeB.(*CircleShape)

				// Calculate the distance between the centers
				distance := math.Distance(posA, posB)

				// The depth is the sum of the radii minus the distance
				collision.Depth = circleA.Radius + circleB.Radius - distance

				// Calculate the collision point
				collision.Point = math.Vector{
					X: posB.X + dx*circleB.Radius,
					Y: posB.Y + dy*circleB.Radius,
				}

			case shapeA.GetType() == ShapeTypeAABB && shapeB.GetType() == ShapeTypeAABB:
				// AABB-AABB collision
				aabbA := shapeA.(*AABBShape)
				aabbB := shapeB.(*AABBShape)

				// Calculate the overlap on each axis
				minA, maxA := aabbA.GetMin(), aabbA.GetMax()
				minB, maxB := aabbB.GetMin(), aabbB.GetMax()

				overlapX := stdmath.Min(maxA.X, maxB.X) - stdmath.Max(minA.X, minB.X)
				overlapY := stdmath.Min(maxA.Y, maxB.Y) - stdmath.Max(minA.Y, minB.Y)

				// The depth is the minimum overlap
				if overlapX < overlapY {
					collision.Depth = overlapX
				} else {
					collision.Depth = overlapY
				}

				// Calculate the collision point (center of the overlap)
				collision.Point = math.Vector{
					X: stdmath.Max(minA.X, minB.X) + overlapX/2,
					Y: stdmath.Max(minA.Y, minB.Y) + overlapY/2,
				}

			default:
				// Mixed shape types - use a simplified approach
				// This is a simplification and might not be accurate for all cases
				collision.Depth = 1.0
				collision.Point = math.Vector{
					X: (posA.X + posB.X) / 2,
					Y: (posA.Y + posB.Y) / 2,
				}
			}

			collisions = append(collisions, collision)
		}
	}

	return collisions, nil
}

// ResolveWithMovement checks for collisions with movement and returns a list of collisions.
func (ecs *EbitenCollisionSystem) ResolveWithMovement(dx, dy float64, filter CollisionFilter) ([]Collision, error) {
	// Store the original positions
	originalPositions := make(map[int]math.Vector)
	for id, shape := range ecs.shapes {
		originalPositions[id] = shape.GetPosition()
	}

	// Move all shapes
	for id, shape := range ecs.shapes {
		pos := shape.GetPosition()
		shape.SetPosition(math.Vector{X: pos.X + dx, Y: pos.Y + dy})
		ecs.UpdateShape(id, shape)
	}

	// Check for collisions
	collisions, err := ecs.Resolve(filter)
	if err != nil {
		return nil, err
	}

	// Restore original positions
	for id, pos := range originalPositions {
		shape := ecs.shapes[id]
		shape.SetPosition(pos)
		ecs.UpdateShape(id, shape)
	}

	return collisions, nil
}

// GetNearbyShapes returns all shapes within the specified radius of the position.
func (ecs *EbitenCollisionSystem) GetNearbyShapes(position math.Vector, radius float64) []Shape {
	// Create a temporary circle to query the space
	circle := collisions.NewCircle(position.X, position.Y, radius)

	// Add the circle to the space
	ecs.space.Add(circle)

	// Update the space
	ecs.space.Update()

	// Get all objects that collide with the circle
	colliding := ecs.space.GetCollisions(circle)

	// Remove the circle from the space
	ecs.space.Remove(circle)

	// Convert the colliding objects to shapes
	var shapes []Shape
	for _, obj := range colliding {
		for id, o := range ecs.objects {
			if o == obj {
				shapes = append(shapes, ecs.shapes[id])
				break
			}
		}
	}

	return shapes
}

// Clear removes all shapes from the collision system.
func (ecs *EbitenCollisionSystem) Clear() {
	// Create a new space
	ecs.space = collisions.NewSpace(ecs.cellSize)

	// Clear our maps
	ecs.shapes = make(map[int]Shape)
	ecs.objects = make(map[int]collisions.Object)

	// Reset the next ID
	ecs.nextID = 1
}
