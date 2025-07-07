package physics

import (
	"discoveryx/internal/utils/math"
	"testing"
)

func TestCollisionSystem(t *testing.T) {
	// Create a new collision system
	cs := NewEbitenCollisionSystem(100.0)
	
	// Create two circle shapes
	circle1 := &CircleShape{
		Position: math.Vector{X: 0, Y: 0},
		Radius:   10,
	}
	
	circle2 := &CircleShape{
		Position: math.Vector{X: 15, Y: 0},
		Radius:   10,
	}
	
	// Add the shapes to the collision system
	id1 := cs.AddShape(circle1)
	id2 := cs.AddShape(circle2)
	
	// Check for collisions
	collisions, err := cs.Resolve(nil)
	if err != nil {
		t.Fatalf("Error resolving collisions: %v", err)
	}
	
	// There should be one collision
	if len(collisions) != 1 {
		t.Fatalf("Expected 1 collision, got %d", len(collisions))
	}
	
	// Check the collision properties
	collision := collisions[0]
	if collision.ShapeA != circle1 && collision.ShapeA != circle2 {
		t.Fatalf("Collision ShapeA is not one of the circles")
	}
	if collision.ShapeB != circle1 && collision.ShapeB != circle2 {
		t.Fatalf("Collision ShapeB is not one of the circles")
	}
	
	// Move circle2 away from circle1
	circle2.Position = math.Vector{X: 30, Y: 0}
	cs.UpdateShape(id2, circle2)
	
	// Check for collisions again
	collisions, err = cs.Resolve(nil)
	if err != nil {
		t.Fatalf("Error resolving collisions: %v", err)
	}
	
	// There should be no collisions
	if len(collisions) != 0 {
		t.Fatalf("Expected 0 collisions, got %d", len(collisions))
	}
	
	// Test AABB shapes
	aabb1 := &AABBShape{
		Position: math.Vector{X: 0, Y: 0},
		Width:    20,
		Height:   20,
	}
	
	aabb2 := &AABBShape{
		Position: math.Vector{X: 15, Y: 0},
		Width:    20,
		Height:   20,
	}
	
	// Add the shapes to the collision system
	id3 := cs.AddShape(aabb1)
	id4 := cs.AddShape(aabb2)
	
	// Check for collisions
	collisions, err = cs.Resolve(nil)
	if err != nil {
		t.Fatalf("Error resolving collisions: %v", err)
	}
	
	// There should be one collision
	if len(collisions) != 1 {
		t.Fatalf("Expected 1 collision, got %d", len(collisions))
	}
	
	// Check the collision properties
	collision = collisions[0]
	if collision.ShapeA != aabb1 && collision.ShapeA != aabb2 {
		t.Fatalf("Collision ShapeA is not one of the AABBs")
	}
	if collision.ShapeB != aabb1 && collision.ShapeB != aabb2 {
		t.Fatalf("Collision ShapeB is not one of the AABBs")
	}
	
	// Move aabb2 away from aabb1
	aabb2.Position = math.Vector{X: 30, Y: 0}
	cs.UpdateShape(id4, aabb2)
	
	// Check for collisions again
	collisions, err = cs.Resolve(nil)
	if err != nil {
		t.Fatalf("Error resolving collisions: %v", err)
	}
	
	// There should be no collisions
	if len(collisions) != 0 {
		t.Fatalf("Expected 0 collisions, got %d", len(collisions))
	}
	
	// Test collision filtering
	// Move circle2 back to collide with circle1
	circle2.Position = math.Vector{X: 15, Y: 0}
	cs.UpdateShape(id2, circle2)
	
	// Create a filter that only allows circle-circle collisions
	filter := func(self, other Shape) bool {
		return self.GetType() == ShapeTypeCircle && other.GetType() == ShapeTypeCircle
	}
	
	// Check for collisions with the filter
	collisions, err = cs.Resolve(filter)
	if err != nil {
		t.Fatalf("Error resolving collisions: %v", err)
	}
	
	// There should be one collision
	if len(collisions) != 1 {
		t.Fatalf("Expected 1 collision, got %d", len(collisions))
	}
	
	// Check the collision properties
	collision = collisions[0]
	if collision.ShapeA != circle1 && collision.ShapeA != circle2 {
		t.Fatalf("Collision ShapeA is not one of the circles")
	}
	if collision.ShapeB != circle1 && collision.ShapeB != circle2 {
		t.Fatalf("Collision ShapeB is not one of the circles")
	}
	
	// Test removing shapes
	cs.RemoveShape(id1)
	cs.RemoveShape(id2)
	
	// Check for collisions again
	collisions, err = cs.Resolve(nil)
	if err != nil {
		t.Fatalf("Error resolving collisions: %v", err)
	}
	
	// There should be no collisions
	if len(collisions) != 0 {
		t.Fatalf("Expected 0 collisions, got %d", len(collisions))
	}
}