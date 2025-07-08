// Package collisions provides a minimal implementation of the ebiten-collisions package.
package collisions

// Object is the interface for all collision objects.
type Object interface {
	// GetX returns the X coordinate of the object.
	GetX() float64

	// GetY returns the Y coordinate of the object.
	GetY() float64

	// SetX sets the X coordinate of the object.
	SetX(x float64)

	// SetY sets the Y coordinate of the object.
	SetY(y float64)

	// Collides checks if this object collides with another object.
	Collides(other Object) bool
}

// SpatialHash represents a spatial partitioning system for efficient collision detection.
type SpatialHash struct {
	cellSize float64
	cells    map[[2]int][]Object // Key is (cellX, cellY)
}

// NewSpatialHash creates a new spatial hash with the specified cell size.
func NewSpatialHash(cellSize float64) *SpatialHash {
	return &SpatialHash{
		cellSize: cellSize,
		cells:    make(map[[2]int][]Object),
	}
}

// getCellCoords returns the cell coordinates for a given position.
func (sh *SpatialHash) getCellCoords(x, y float64) [2]int {
	cellX := int(x / sh.cellSize)
	cellY := int(y / sh.cellSize)
	return [2]int{cellX, cellY}
}

// getObjectCells returns all cell coordinates that an object occupies.
// This handles objects that span multiple cells based on their type.
func (sh *SpatialHash) getObjectCells(obj Object) [][2]int {
	var cells [][2]int

	switch o := obj.(type) {
	case *Circle:
		// For circles, we need to check cells that the circle might overlap
		// Get the cell coordinates for the top-left and bottom-right corners of the bounding box
		minX := o.X - o.Radius
		minY := o.Y - o.Radius
		maxX := o.X + o.Radius
		maxY := o.Y + o.Radius

		// Calculate the cell coordinates for the corners
		minCellX := int(minX / sh.cellSize)
		minCellY := int(minY / sh.cellSize)
		maxCellX := int(maxX / sh.cellSize)
		maxCellY := int(maxY / sh.cellSize)

		// Add all cells in the range
		for x := minCellX; x <= maxCellX; x++ {
			for y := minCellY; y <= maxCellY; y++ {
				cells = append(cells, [2]int{x, y})
			}
		}

	case *Rectangle:
		// For rectangles, we need to check cells that the rectangle might overlap
		// Get the cell coordinates for the top-left and bottom-right corners
		minCellX := int(o.X / sh.cellSize)
		minCellY := int(o.Y / sh.cellSize)
		maxCellX := int((o.X + o.W) / sh.cellSize)
		maxCellY := int((o.Y + o.H) / sh.cellSize)

		// Add all cells in the range
		for x := minCellX; x <= maxCellX; x++ {
			for y := minCellY; y <= maxCellY; y++ {
				cells = append(cells, [2]int{x, y})
			}
		}

	default:
		// For other object types, just use the center point
		cells = append(cells, sh.getCellCoords(obj.GetX(), obj.GetY()))
	}

	return cells
}

// addObject adds an object to the spatial hash.
func (sh *SpatialHash) addObject(obj Object) {
	cells := sh.getObjectCells(obj)
	for _, cell := range cells {
		sh.cells[cell] = append(sh.cells[cell], obj)
	}
}

// removeObject removes an object from the spatial hash.
func (sh *SpatialHash) removeObject(obj Object) {
	cells := sh.getObjectCells(obj)
	for _, cell := range cells {
		objects := sh.cells[cell]
		for i, o := range objects {
			if o == obj {
				// Remove the object from this cell
				sh.cells[cell] = append(objects[:i], objects[i+1:]...)
				break
			}
		}
		// If the cell is now empty, remove it from the map
		if len(sh.cells[cell]) == 0 {
			delete(sh.cells, cell)
		}
	}
}

// updateObject updates an object's position in the spatial hash.
func (sh *SpatialHash) updateObject(obj Object, oldX, oldY float64) {
	// Store the original position
	originalX, originalY := obj.GetX(), obj.GetY()

	// Temporarily set the object's position to the old position
	obj.SetX(oldX)
	obj.SetY(oldY)

	// Get the old cells
	oldCells := sh.getObjectCells(obj)

	// Restore the object's position
	obj.SetX(originalX)
	obj.SetY(originalY)

	// Get the new cells
	newCells := sh.getObjectCells(obj)

	// If the cells haven't changed, no need to update
	if len(oldCells) == len(newCells) {
		allSame := true
		for i, oldCell := range oldCells {
			if oldCell != newCells[i] {
				allSame = false
				break
			}
		}
		if allSame {
			return
		}
	}

	// Remove from old cells
	for _, cell := range oldCells {
		for i, o := range sh.cells[cell] {
			if o == obj {
				sh.cells[cell] = append(sh.cells[cell][:i], sh.cells[cell][i+1:]...)
				break
			}
		}
		// If the cell is now empty, remove it from the map
		if len(sh.cells[cell]) == 0 {
			delete(sh.cells, cell)
		}
	}

	// Add to new cells
	for _, cell := range newCells {
		sh.cells[cell] = append(sh.cells[cell], obj)
	}
}

// getPotentialCollisions returns all objects that could potentially collide with the given object.
func (sh *SpatialHash) getPotentialCollisions(obj Object) []Object {
	var potentialCollisions []Object

	// Get all cells that the object spans
	objCells := sh.getObjectCells(obj)

	// Create a map to track which objects we've already added
	// This prevents adding the same object multiple times if it spans multiple cells
	addedObjects := make(map[Object]bool)

	// Check all cells that the object spans and their neighboring cells
	for _, cell := range objCells {
		// Check the cell and all neighboring cells
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				neighborCell := [2]int{cell[0] + dx, cell[1] + dy}
				for _, o := range sh.cells[neighborCell] {
					// Skip the object itself and objects we've already added
					if o != obj && !addedObjects[o] {
						potentialCollisions = append(potentialCollisions, o)
						addedObjects[o] = true
					}
				}
			}
		}
	}

	return potentialCollisions
}

// Space represents a collision space that contains objects.
type Space struct {
	cellSize float64
	objects  []Object
	spatialHash *SpatialHash
}

// NewSpace creates a new collision space with the specified cell size.
func NewSpace(cellSize float64) *Space {
	return &Space{
		cellSize: cellSize,
		objects:  make([]Object, 0),
		spatialHash: NewSpatialHash(cellSize),
	}
}

// Add adds an object to the space.
func (s *Space) Add(obj Object) {
	s.objects = append(s.objects, obj)
	s.spatialHash.addObject(obj)
}

// Remove removes an object from the space.
func (s *Space) Remove(obj Object) {
	for i, o := range s.objects {
		if o == obj {
			s.objects = append(s.objects[:i], s.objects[i+1:]...)
			s.spatialHash.removeObject(obj)
			return
		}
	}
}

// UpdateShape updates an object's position in the space.
// This should be called whenever an object's position changes.
func (s *Space) UpdateShape(obj Object, oldX, oldY float64) {
	s.spatialHash.updateObject(obj, oldX, oldY)
}

// Update updates the space.
func (s *Space) Update() {
	// No-op in this minimal implementation
}

// GetCollisions returns all objects that collide with the specified object.
func (s *Space) GetCollisions(obj Object) []Object {
	var collisions []Object

	// Get potential collisions from the spatial hash
	potentialCollisions := s.spatialHash.getPotentialCollisions(obj)

	// Check actual collisions
	for _, o := range potentialCollisions {
		if obj.Collides(o) {
			collisions = append(collisions, o)
		}
	}

	return collisions
}

// Circle represents a circular collision object.
type Circle struct {
	X      float64
	Y      float64
	Radius float64
}

// GetX returns the X coordinate of the circle.
func (c *Circle) GetX() float64 {
	return c.X
}

// GetY returns the Y coordinate of the circle.
func (c *Circle) GetY() float64 {
	return c.Y
}

// SetX sets the X coordinate of the circle.
func (c *Circle) SetX(x float64) {
	c.X = x
}

// SetY sets the Y coordinate of the circle.
func (c *Circle) SetY(y float64) {
	c.Y = y
}

// Collides checks if this circle collides with another object.
func (c *Circle) Collides(other Object) bool {
	switch o := other.(type) {
	case *Circle:
		dx := c.X - o.X
		dy := c.Y - o.Y
		distance := dx*dx + dy*dy
		return distance < (c.Radius+o.Radius)*(c.Radius+o.Radius)
	case *Rectangle:
		// Find the closest point on the rectangle to the circle
		closestX := max(o.X, min(c.X, o.X+o.W))
		closestY := max(o.Y, min(c.Y, o.Y+o.H))

		// Calculate the distance between the circle's center and the closest point
		dx := closestX - c.X
		dy := closestY - c.Y
		distance := dx*dx + dy*dy

		// If the distance is less than the circle's radius, there is a collision
		return distance < c.Radius*c.Radius
	default:
		return false
	}
}

// NewCircle creates a new circle with the specified position and radius.
func NewCircle(x, y, radius float64) *Circle {
	return &Circle{
		X:      x,
		Y:      y,
		Radius: radius,
	}
}

// Rectangle represents a rectangular collision object.
type Rectangle struct {
	X float64
	Y float64
	W float64
	H float64
}

// GetX returns the X coordinate of the rectangle.
func (r *Rectangle) GetX() float64 {
	return r.X
}

// GetY returns the Y coordinate of the rectangle.
func (r *Rectangle) GetY() float64 {
	return r.Y
}

// SetX sets the X coordinate of the rectangle.
func (r *Rectangle) SetX(x float64) {
	r.X = x
}

// SetY sets the Y coordinate of the rectangle.
func (r *Rectangle) SetY(y float64) {
	r.Y = y
}

// Collides checks if this rectangle collides with another object.
func (r *Rectangle) Collides(other Object) bool {
	switch o := other.(type) {
	case *Circle:
		// Find the closest point on the rectangle to the circle
		closestX := max(r.X, min(o.X, r.X+r.W))
		closestY := max(r.Y, min(o.Y, r.Y+r.H))

		// Calculate the distance between the circle's center and the closest point
		dx := closestX - o.X
		dy := closestY - o.Y
		distance := dx*dx + dy*dy

		// If the distance is less than the circle's radius, there is a collision
		return distance < o.Radius*o.Radius
	case *Rectangle:
		// Check if the rectangles overlap
		return r.X < o.X+o.W && r.X+r.W > o.X && r.Y < o.Y+o.H && r.Y+r.H > o.Y
	default:
		return false
	}
}

// NewRectangle creates a new rectangle with the specified position and size.
func NewRectangle(x, y, w, h float64) *Rectangle {
	return &Rectangle{
		X: x,
		Y: y,
		W: w,
		H: h,
	}
}

// Helper functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
