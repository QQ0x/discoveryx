// Package physics provides collision detection and physics simulation for the game.
package physics

import (
	"discoveryx/internal/utils/math"
)

// Shape is the interface for all collision shapes.
// It defines the methods that all collision shapes must implement.
type Shape interface {
	// GetPosition returns the position of the shape.
	GetPosition() math.Vector
	
	// SetPosition updates the position of the shape.
	SetPosition(position math.Vector)
	
	// GetType returns the type of the shape.
	GetType() ShapeType
}

// ShapeType represents the type of a collision shape.
type ShapeType int

const (
	// ShapeTypeCircle represents a circular collision shape.
	ShapeTypeCircle ShapeType = iota
	
	// ShapeTypeAABB represents an axis-aligned bounding box collision shape.
	ShapeTypeAABB
	
	// ShapeTypePolygon represents a polygon collision shape.
	ShapeTypePolygon
)

// CircleShape represents a circular collision shape.
type CircleShape struct {
	Position math.Vector
	Radius   float64
}

// GetPosition returns the position of the circle.
func (c *CircleShape) GetPosition() math.Vector {
	return c.Position
}

// SetPosition updates the position of the circle.
func (c *CircleShape) SetPosition(position math.Vector) {
	c.Position = position
}

// GetType returns the type of the shape.
func (c *CircleShape) GetType() ShapeType {
	return ShapeTypeCircle
}

// AABBShape represents an axis-aligned bounding box collision shape.
type AABBShape struct {
	Position math.Vector
	Width    float64
	Height   float64
}

// GetPosition returns the position of the AABB.
func (a *AABBShape) GetPosition() math.Vector {
	return a.Position
}

// SetPosition updates the position of the AABB.
func (a *AABBShape) SetPosition(position math.Vector) {
	a.Position = position
}

// GetType returns the type of the shape.
func (a *AABBShape) GetType() ShapeType {
	return ShapeTypeAABB
}

// GetMin returns the minimum point of the AABB.
func (a *AABBShape) GetMin() math.Vector {
	return math.Vector{
		X: a.Position.X - a.Width/2,
		Y: a.Position.Y - a.Height/2,
	}
}

// GetMax returns the maximum point of the AABB.
func (a *AABBShape) GetMax() math.Vector {
	return math.Vector{
		X: a.Position.X + a.Width/2,
		Y: a.Position.Y + a.Height/2,
	}
}

// Collision represents a collision between two shapes.
type Collision struct {
	// ShapeA is the first shape involved in the collision.
	ShapeA Shape
	
	// ShapeB is the second shape involved in the collision.
	ShapeB Shape
	
	// Normal is the collision normal vector (direction to resolve the collision).
	Normal math.Vector
	
	// Depth is the penetration depth of the collision.
	Depth float64
	
	// Point is the point of collision.
	Point math.Vector
}

// CollisionFilter is a function that determines whether two shapes should collide.
type CollisionFilter func(self, other Shape) bool

// CollisionSystem is the interface for collision detection systems.
// It defines the methods that all collision systems must implement.
type CollisionSystem interface {
	// AddShape adds a shape to the collision system.
	AddShape(shape Shape) int
	
	// RemoveShape removes a shape from the collision system.
	RemoveShape(id int)
	
	// UpdateShape updates a shape in the collision system.
	UpdateShape(id int, shape Shape)
	
	// Resolve checks for collisions and returns a list of collisions.
	// The filter parameter can be used to selectively enable/disable collisions.
	Resolve(filter CollisionFilter) ([]Collision, error)
	
	// ResolveWithMovement checks for collisions with movement and returns a list of collisions.
	// This is used for continuous collision detection.
	ResolveWithMovement(dx, dy float64, filter CollisionFilter) ([]Collision, error)
	
	// GetNearbyShapes returns all shapes within the specified radius of the position.
	GetNearbyShapes(position math.Vector, radius float64) []Shape
}