package physics

import (
	"discoveryx/internal/utils/math"
	"fmt"
)

// SpatialGrid is a simple spatial partitioning system that divides the world into a grid
// of cells to optimize collision detection. This allows for O(1) lookup of nearby entities
// instead of checking all entities against each other (O(n²)).
type SpatialGrid struct {
	cellSize  float64                      // Size of each grid cell
	grid      map[string][]interface{}     // Grid cells containing entities
	positions map[interface{}]math.Vector  // Entity positions for quick lookup
	cellKeys  map[interface{}]string       // Cell keys for each entity for quick removal
}

// NewSpatialGrid creates a new spatial grid with the specified cell size.
// The cell size should be chosen based on the typical size and distribution of entities.
// A good rule of thumb is to use a cell size that's 2-4 times the size of the largest entity.
func NewSpatialGrid(cellSize float64) *SpatialGrid {
	return &SpatialGrid{
		cellSize:  cellSize,
		grid:      make(map[string][]interface{}),
		positions: make(map[interface{}]math.Vector),
		cellKeys:  make(map[interface{}]string),
	}
}

// getCellKey returns a string key for a grid cell based on its coordinates
func getCellKey(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

// getCellCoords returns the grid cell coordinates for a world position
func (sg *SpatialGrid) getCellCoords(position math.Vector) (int, int) {
	return int(position.X / sg.cellSize), int(position.Y / sg.cellSize)
}

// Insert adds an entity to the spatial grid at the specified position
func (sg *SpatialGrid) Insert(entity interface{}, position math.Vector) {
	// Get grid cell coordinates
	cellX, cellY := sg.getCellCoords(position)
	cellKey := getCellKey(cellX, cellY)

	// Store the entity's position
	sg.positions[entity] = position

	// Store the entity's cell key
	sg.cellKeys[entity] = cellKey

	// Add the entity to the grid cell
	if _, exists := sg.grid[cellKey]; !exists {
		sg.grid[cellKey] = make([]interface{}, 0, 8) // Pre-allocate for efficiency
	}
	sg.grid[cellKey] = append(sg.grid[cellKey], entity)
}

// Remove removes an entity from the spatial grid
func (sg *SpatialGrid) Remove(entity interface{}) {
	// Get the entity's cell key
	cellKey, exists := sg.cellKeys[entity]
	if !exists {
		return
	}

	// Remove the entity from the grid cell
	entities := sg.grid[cellKey]
	for i, e := range entities {
		if e == entity {
			// Remove by swapping with the last element and truncating
			entities[i] = entities[len(entities)-1]
			sg.grid[cellKey] = entities[:len(entities)-1]
			break
		}
	}

	// Clean up empty cells
	if len(sg.grid[cellKey]) == 0 {
		delete(sg.grid, cellKey)
	}

	// Remove from tracking maps
	delete(sg.positions, entity)
	delete(sg.cellKeys, entity)
}

// Update updates an entity's position in the spatial grid
func (sg *SpatialGrid) Update(entity interface{}, newPosition math.Vector) {
	// Get the entity's current cell key
	oldCellKey, exists := sg.cellKeys[entity]
	if !exists {
		// If the entity doesn't exist, insert it
		sg.Insert(entity, newPosition)
		return
	}

	// Get the new cell key
	newCellX, newCellY := sg.getCellCoords(newPosition)
	newCellKey := getCellKey(newCellX, newCellY)

	// If the cell hasn't changed, just update the position
	if oldCellKey == newCellKey {
		sg.positions[entity] = newPosition
		return
	}

	// Remove from old cell
	sg.Remove(entity)

	// Insert into new cell
	sg.Insert(entity, newPosition)
}

// QueryNearby returns all entities in the same cell as the specified position
func (sg *SpatialGrid) QueryNearby(position math.Vector) []interface{} {
	cellX, cellY := sg.getCellCoords(position)
	cellKey := getCellKey(cellX, cellY)

	if entities, exists := sg.grid[cellKey]; exists {
		return entities
	}

	return nil
}

// QueryRadius returns all entities within the specified radius of the position
func (sg *SpatialGrid) QueryRadius(position math.Vector, radius float64) []interface{} {
	// Calculate the grid cells that could contain entities within the radius
	minCellX := int((position.X - radius) / sg.cellSize)
	maxCellX := int((position.X + radius) / sg.cellSize)
	minCellY := int((position.Y - radius) / sg.cellSize)
	maxCellY := int((position.Y + radius) / sg.cellSize)

	// Collect all entities in those cells
	var result []interface{}
	for cellX := minCellX; cellX <= maxCellX; cellX++ {
		for cellY := minCellY; cellY <= maxCellY; cellY++ {
			cellKey := getCellKey(cellX, cellY)
			if entities, exists := sg.grid[cellKey]; exists {
				result = append(result, entities...)
			}
		}
	}

	// Filter entities by actual distance
	var filtered []interface{}
	for _, entity := range result {
		entityPos := sg.positions[entity]
		dx := entityPos.X - position.X
		dy := entityPos.Y - position.Y
		distSquared := dx*dx + dy*dy
		if distSquared <= radius*radius {
			filtered = append(filtered, entity)
		}
	}

	return filtered
}

// Clear removes all entities from the spatial grid
func (sg *SpatialGrid) Clear() {
	sg.grid = make(map[string][]interface{})
	sg.positions = make(map[interface{}]math.Vector)
	sg.cellKeys = make(map[interface{}]string)
}
