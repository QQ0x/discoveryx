# Spatial Hashing Collision System

This package implements a spatial hashing system for efficient collision detection. The system divides the space into a grid of cells and assigns objects to cells based on their position. When checking for collisions, only objects in the same cell and neighboring cells are considered, which significantly reduces the number of collision checks required.

## How It Works

1. The space is divided into a grid of cells, each with a size specified when creating the collision space.
2. Objects are added to cells based on their position and size:
   - For point-like objects, they are added to the cell containing their center point.
   - For objects with size (like circles and rectangles), they are added to all cells they overlap.
3. When checking for collisions, only objects in the same cell and neighboring cells are considered.

## Performance Improvements

The spatial hashing system significantly improves the performance of collision detection:

- **Without spatial hashing**: O(n²) collision checks, where n is the total number of objects in the space.
- **With spatial hashing**: O(k) collision checks, where k is the number of objects in the nearby cells.

For a game with hundreds of objects (like walls) distributed across the space, this can reduce the number of collision checks from tens of thousands to just a few dozen, resulting in a massive performance improvement.

## Usage

### Creating a Collision Space

```
// Create a new collision space with a cell size of 100 units
space := collisions.NewSpace(100.0)
```

### Adding Objects

```
// Create a circle
circle := collisions.NewCircle(x, y, radius)

// Add the circle to the space
space.Add(circle)

// Create a rectangle
rect := collisions.NewRectangle(x, y, width, height)

// Add the rectangle to the space
space.Add(rect)
```

### Updating Object Positions

When an object's position changes, you need to update its position in the spatial hash:

```
// Store the old position
oldX, oldY := obj.GetX(), obj.GetY()

// Update the object's position
obj.SetX(newX)
obj.SetY(newY)

// Update the object in the space
space.UpdateShape(obj, oldX, oldY)
```

### Checking for Collisions

```
// Get all objects that collide with a specific object
collisions := space.GetCollisions(obj)

// Process the collisions
for _, other := range collisions {
    // Handle the collision
}
```

## Implementation Details

The spatial hashing system consists of the following components:

1. **SpatialHash**: A structure that manages the spatial partitioning of objects.
2. **Space**: A collision space that uses the spatial hash for efficient collision detection.
3. **Object**: An interface for collision objects (Circle, Rectangle, etc.).

The system handles objects that span multiple cells by adding them to all cells they overlap. When updating an object's position, it removes the object from its old cells and adds it to its new cells. When checking for collisions, it considers all cells that the object spans and their neighboring cells.

## Best Practices

1. Choose an appropriate cell size:
   - Too small: Objects span many cells, increasing overhead.
   - Too large: Too many objects per cell, reducing the benefit of spatial partitioning.
   - A good rule of thumb is to use a cell size that's roughly the size of your average object.

2. Update object positions correctly:
   - Always call `UpdateShape` when an object's position changes.
   - Provide the old position to ensure the object is removed from its old cells.

3. Consider object size:
   - The system handles objects that span multiple cells, but very large objects might still cause performance issues.
   - For very large objects, consider breaking them down into smaller objects if possible.
