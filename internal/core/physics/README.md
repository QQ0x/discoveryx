# Collision Detection System

This document outlines the collision detection system, including the recent complete redesign to use specialized external libraries.

## Overview of Changes

### Initial Implementation
1. **Spatial Partitioning**: Grid-based spatial partitioning system to reduce collision checks from O(n²) to O(n).
2. **Broad Phase and Narrow Phase**: Two-phase collision detection approach:
   - Broad phase: Quickly identify potential collisions using spatial partitioning
   - Narrow phase: Perform precise collision checks only on nearby entities
3. **Wall Collider Optimization**: System to convert pixel-level wall points into optimized rectangular colliders.
4. **Centralized Collision Management**: CollisionManager to centralize all collision detection logic.
5. **Comprehensive Collision Detection**:
   - Player-enemy collisions
   - Player bullet-enemy collisions
   - Player-wall collisions

### Intermediate Improvements
6. **Continuous Collision Detection**: Swept-sphere collision detection for fast-moving objects to prevent tunneling:
   - Circle-Rectangle continuous collision detection for bullet-wall collisions
   - Circle-Circle continuous collision detection for bullet-enemy and bullet-player collisions
7. **Improved Wall Collider Generation**: Enhanced wall collider generator for merging adjacent wall points.
8. **Wall Optimization**: OptimizeWalls method to reduce the number of wall colliders.
9. **Consistent Collision Management**: Consistent use of CollisionManager for all collision detection.

### Complete Redesign (Current Version)
10. **Modular Library Integration**: Integrated the [ebiten-collisions](https://github.com/tducasse/ebiten-collisions) library for robust collision detection.
11. **Interface-Based Architecture**: Designed a flexible, interface-based architecture:
    - `Shape` interface for all collision shapes
    - `CollisionSystem` interface for collision detection systems
    - `CollisionFilter` for selective collision detection
12. **Expanded Shape Support**: Added support for multiple collision shape types:
    - CircleShape for entity-entity collisions
    - AABBShape for entity-wall collisions
13. **Enhanced Collision Response**: Improved collision response with:
    - Precise separation vectors
    - Axis-separated collision resolution
    - Minimal movement adjustments
14. **Testable Design**: Created a testable design with unit tests for the collision system.

## Architecture

The collision system is designed with modularity and extensibility in mind:

### Interfaces

- `Shape`: Base interface for all collision shapes
- `CollisionSystem`: Interface for collision detection systems
- `CollisionFilter`: Function type for selective collision detection

### Shapes

- `CircleShape`: Circular collision shape
- `AABBShape`: Axis-aligned bounding box collision shape

### Collision Detection

- `EbitenCollisionSystem`: Implementation using the ebiten-collisions library
- `CollisionManager`: High-level manager providing a convenient API

## Usage Examples

### Checking for Collisions

```go
// Check if an entity collides with any other entity
collision, collidedEntity := collisionManager.CheckCollision(player, player.GetCollider().Radius)

// Check if an entity collides with any wall
collision, separationVector, isXAxis := collisionManager.CheckAABBWallCollision(player, plannedPosition)
```

### Selective Collision Detection

```go
// Create a filter that only allows specific collision types
filter := func(self, other Shape) bool {
    // Only check collisions between specific shape types
    return self.GetType() == ShapeTypeCircle && other.GetType() == ShapeTypeCircle
}

// Check for collisions with the filter
collisions, _ := collisionSystem.Resolve(filter)
```

## Performance Improvements

The redesigned collision system provides significant performance improvements:

1. **Optimized Spatial Partitioning**: The ebiten-collisions library provides efficient spatial partitioning.
2. **Reduced Collision Checks**: Only checks for collisions between entities that are close to each other.
3. **Efficient Shape Management**: Optimized shape creation and management.
4. **Selective Collision Detection**: Filters allow checking only relevant collision pairs.

## Extensibility

The system is designed to be extensible:

1. **New Shape Types**: Add new shape types by implementing the `Shape` interface.
2. **Custom Collision Filters**: Create custom filters for selective collision detection.
3. **Alternative Implementations**: Implement the `CollisionSystem` interface with different libraries.
4. **Enhanced Collision Response**: Extend with additional collision response behaviors.

## Future Improvements

While the current implementation provides significant improvements, there are still opportunities for enhancement:

1. **Additional Shape Types**: Add support for polygon shapes and other complex geometries.
2. **Physics Integration**: Integrate with a physics engine for more realistic collision response.
3. **Parallel Processing**: Use goroutines to parallelize collision detection.
4. **Visualization Tools**: Add debug visualization for collision shapes and detections.
