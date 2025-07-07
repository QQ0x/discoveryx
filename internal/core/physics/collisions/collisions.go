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

// Space represents a collision space that contains objects.
type Space struct {
	cellSize float64
	objects  []Object
}

// NewSpace creates a new collision space with the specified cell size.
func NewSpace(cellSize float64) *Space {
	return &Space{
		cellSize: cellSize,
		objects:  make([]Object, 0),
	}
}

// Add adds an object to the space.
func (s *Space) Add(obj Object) {
	s.objects = append(s.objects, obj)
}

// Remove removes an object from the space.
func (s *Space) Remove(obj Object) {
	for i, o := range s.objects {
		if o == obj {
			s.objects = append(s.objects[:i], s.objects[i+1:]...)
			return
		}
	}
}

// Update updates the space.
func (s *Space) Update() {
	// No-op in this minimal implementation
}

// GetCollisions returns all objects that collide with the specified object.
func (s *Space) GetCollisions(obj Object) []Object {
	var collisions []Object
	for _, o := range s.objects {
		if o != obj && obj.Collides(o) {
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