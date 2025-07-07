// Package physics provides collision detection and physics simulation for the game.
package physics

import (
	"discoveryx/internal/utils/math"
)

// CollisionManager centralizes collision detection and provides optimized methods
// for detecting collisions between different types of entities. It uses the
// EbitenCollisionSystem for efficient collision detection.
type CollisionManager struct {
	// The underlying collision system
	collisionSystem CollisionSystem

	// Maps entities to their shape IDs
	entityShapeIDs map[interface{}]int

	// Maps wall colliders to their shape IDs
	wallShapeIDs map[RectCollider]int
}

// NewCollisionManager creates a new collision manager with the specified cell size.
func NewCollisionManager(cellSize float64) *CollisionManager {
	return &CollisionManager{
		collisionSystem: NewEbitenCollisionSystem(cellSize),
		entityShapeIDs:  make(map[interface{}]int),
		wallShapeIDs:    make(map[RectCollider]int),
	}
}

// RegisterEntity adds an entity to the collision system with its collider.
func (cm *CollisionManager) RegisterEntity(entity interface{}, collider CircleCollider) {
	// Create a circle shape from the collider
	circleShape := &CircleShape{
		Position: collider.Position,
		Radius:   collider.Radius,
	}

	// Add the shape to the collision system
	shapeID := cm.collisionSystem.AddShape(circleShape)

	// Store the shape ID for later lookup
	cm.entityShapeIDs[entity] = shapeID
}

// UpdateEntity updates an entity's position and collider in the collision system.
func (cm *CollisionManager) UpdateEntity(entity interface{}, collider CircleCollider) {
	// Get the shape ID for this entity
	shapeID, exists := cm.entityShapeIDs[entity]
	if !exists {
		// If the entity doesn't exist, register it
		cm.RegisterEntity(entity, collider)
		return
	}

	// Create a circle shape from the collider
	circleShape := &CircleShape{
		Position: collider.Position,
		Radius:   collider.Radius,
	}

	// Update the shape in the collision system
	cm.collisionSystem.UpdateShape(shapeID, circleShape)
}

// RemoveEntity removes an entity from the collision system.
func (cm *CollisionManager) RemoveEntity(entity interface{}) {
	// Get the shape ID for this entity
	shapeID, exists := cm.entityShapeIDs[entity]
	if !exists {
		return
	}

	// Remove the shape from the collision system
	cm.collisionSystem.RemoveShape(shapeID)

	// Remove the entity from our map
	delete(cm.entityShapeIDs, entity)
}

// RegisterWall adds a wall to the collision system.
func (cm *CollisionManager) RegisterWall(wall RectCollider) {
	// Create an AABB shape from the wall
	aabbShape := &AABBShape{
		Position: wall.Position,
		Width:    wall.Width,
		Height:   wall.Height,
	}

	// Add the shape to the collision system
	shapeID := cm.collisionSystem.AddShape(aabbShape)

	// Store the shape ID for later lookup
	cm.wallShapeIDs[wall] = shapeID
}

// ClearWalls removes all walls from the collision system.
func (cm *CollisionManager) ClearWalls() {
	// Remove all wall shapes from the collision system
	for _, shapeID := range cm.wallShapeIDs {
		cm.collisionSystem.RemoveShape(shapeID)
	}

	// Clear our map
	cm.wallShapeIDs = make(map[RectCollider]int)
}

// CheckCollision checks if the specified entity collides with any other entity.
func (cm *CollisionManager) CheckCollision(entity interface{}, radius float64) (bool, interface{}) {
	// Get the shape ID for this entity
	shapeID, exists := cm.entityShapeIDs[entity]
	if !exists {
		return false, nil
	}

	// Create a filter that excludes the entity itself
	filter := func(self, other Shape) bool {
		// Skip self
		if self == other {
			return false
		}

		// Only check entity-entity collisions (circle-circle)
		if self.GetType() != ShapeTypeCircle || other.GetType() != ShapeTypeCircle {
			return false
		}

		return true
	}

	// Check for collisions
	collisions, err := cm.collisionSystem.Resolve(filter)
	if err != nil {
		return false, nil
	}

	// Find collisions involving this entity
	for _, collision := range collisions {
		// Find the entity associated with the other shape
		for e, id := range cm.entityShapeIDs {
			if e == entity {
				continue
			}

			// Check if this entity is involved in the collision
			if (collision.ShapeA == cm.collisionSystem.(*EbitenCollisionSystem).shapes[shapeID] &&
				collision.ShapeB == cm.collisionSystem.(*EbitenCollisionSystem).shapes[id]) ||
				(collision.ShapeB == cm.collisionSystem.(*EbitenCollisionSystem).shapes[shapeID] &&
					collision.ShapeA == cm.collisionSystem.(*EbitenCollisionSystem).shapes[id]) {
				return true, e
			}
		}
	}

	return false, nil
}

// CheckAABBWallCollision performs AABB collision detection between an entity and walls.
func (cm *CollisionManager) CheckAABBWallCollision(entity interface{}, plannedPos math.Vector) (bool, math.Vector, bool) {
	// Check if the entity implements GetAABBCollider
	type aabbGetter interface {
		GetAABBCollider() AABBCollider
	}

	aabbEntity, ok := entity.(aabbGetter)
	if !ok {
		return false, math.Vector{}, false
	}

	// Get the entity's AABB collider
	entityCollider := aabbEntity.GetAABBCollider()

	// Update the position to the planned position
	entityCollider.Position = plannedPos

	// Create an AABB shape from the collider
	aabbShape := &AABBShape{
		Position: entityCollider.Position,
		Width:    entityCollider.Width,
		Height:   entityCollider.Height,
	}

	// Add the shape to the collision system temporarily
	shapeID := cm.collisionSystem.AddShape(aabbShape)

	// Create a filter that only checks AABB-AABB collisions
	filter := func(self, other Shape) bool {
		// Only check AABB-AABB collisions
		return self.GetType() == ShapeTypeAABB && other.GetType() == ShapeTypeAABB
	}

	// Check for collisions
	collisions, err := cm.collisionSystem.Resolve(filter)
	if err != nil {
		cm.collisionSystem.RemoveShape(shapeID)
		return false, math.Vector{}, false
	}

	// Remove the temporary shape
	cm.collisionSystem.RemoveShape(shapeID)

	// If there are no collisions, return false
	if len(collisions) == 0 {
		return false, math.Vector{}, false
	}

	// Find the collision with the smallest depth
	var smallestDepth float64 = 1000000
	var separationVector math.Vector
	var isXAxis bool

	for _, collision := range collisions {
		if collision.Depth < smallestDepth {
			smallestDepth = collision.Depth

			// Determine the separation vector and axis
			if collision.Normal.X != 0 && collision.Normal.Y == 0 {
				// X-axis collision
				separationVector = math.Vector{X: collision.Normal.X * collision.Depth, Y: 0}
				isXAxis = true
			} else if collision.Normal.X == 0 && collision.Normal.Y != 0 {
				// Y-axis collision
				separationVector = math.Vector{X: 0, Y: collision.Normal.Y * collision.Depth}
				isXAxis = false
			} else {
				// Diagonal collision - determine which axis has the smaller overlap
				overlapX := collision.Depth * collision.Normal.X
				overlapY := collision.Depth * collision.Normal.Y

				if abs(overlapX) < abs(overlapY) {
					separationVector = math.Vector{X: overlapX, Y: 0}
					isXAxis = true
				} else {
					separationVector = math.Vector{X: 0, Y: overlapY}
					isXAxis = false
				}
			}
		}
	}

	return true, separationVector, isXAxis
}

// GetNearbyEntities returns all entities within the specified radius of the position.
func (cm *CollisionManager) GetNearbyEntities(position math.Vector, radius float64) []interface{} {
	// Get nearby shapes
	nearbyShapes := cm.collisionSystem.GetNearbyShapes(position, radius)

	// Convert shapes to entities
	var entities []interface{}
	for _, shape := range nearbyShapes {
		// Find the entity associated with this shape
		for entity, shapeID := range cm.entityShapeIDs {
			if cm.collisionSystem.(*EbitenCollisionSystem).shapes[shapeID] == shape {
				entities = append(entities, entity)
				break
			}
		}
	}

	return entities
}

// GetNearbyWalls returns all walls within the specified radius of the position.
func (cm *CollisionManager) GetNearbyWalls(position math.Vector, radius float64) []RectCollider {
	// Get nearby shapes
	nearbyShapes := cm.collisionSystem.GetNearbyShapes(position, radius)

	// Convert shapes to walls
	var walls []RectCollider
	for _, shape := range nearbyShapes {
		// Only consider AABB shapes
		if shape.GetType() != ShapeTypeAABB {
			continue
		}

		// Find the wall associated with this shape
		for wall, shapeID := range cm.wallShapeIDs {
			if cm.collisionSystem.(*EbitenCollisionSystem).shapes[shapeID] == shape {
				walls = append(walls, wall)
				break
			}
		}
	}

	return walls
}

// OptimizeWalls simplifies the wall colliders by merging adjacent walls.
// This is a no-op in the new system as the collision library handles optimization internally.
func (cm *CollisionManager) OptimizeWalls() {
	// No-op - the collision library handles optimization internally
}
