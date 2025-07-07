package physics

import (
	"discoveryx/internal/utils/math"
)

// WallColliderGenerator generates wall colliders from wall points.
// It converts the pixel-level wall points into optimized rectangular colliders
// that can be used for efficient collision detection.
type WallColliderGenerator struct {
	minWallSize float64 // Minimum size of a wall collider
}

// NewWallColliderGenerator creates a new wall collider generator.
func NewWallColliderGenerator(minWallSize float64) *WallColliderGenerator {
	return &WallColliderGenerator{
		minWallSize: minWallSize,
	}
}

// GenerateWallColliders generates wall colliders from wall points.
// It takes a slice of wall points and returns a slice of rectangular colliders.
//
// Parameters:
// - wallPoints: A slice of wall points with position and normal vectors
// - cellSize: The size of a cell in the world grid
//
// Returns:
// - []RectCollider: A slice of rectangular colliders representing the walls
func (wcg *WallColliderGenerator) GenerateWallColliders(wallPoints []WallPoint, cellSize float64) []RectCollider {
	if len(wallPoints) == 0 {
		return nil
	}

	// Create a map to group wall points by their normal vector
	wallsByNormal := make(map[string][]WallPoint)

	// Group wall points by their normal vector (rounded to nearest 90 degrees)
	for _, wp := range wallPoints {
		// Normalize the normal vector to one of the four cardinal directions
		normalKey := wcg.normalizeNormal(wp.Normal)
		wallsByNormal[normalKey] = append(wallsByNormal[normalKey], wp)
	}

	// Generate colliders for each group of wall points
	var colliders []RectCollider

	for normalKey, points := range wallsByNormal {
		// Skip groups with too few points
		if len(points) < 3 {
			continue
		}

		// Generate colliders for this group
		groupColliders := wcg.generateCollidersForGroup(points, normalKey, cellSize)
		colliders = append(colliders, groupColliders...)
	}

	return colliders
}

// normalizeNormal normalizes a normal vector to one of the four cardinal directions.
// This helps group wall points with similar normals together.
func (wcg *WallColliderGenerator) normalizeNormal(normal math.Vector) string {
	// Determine the dominant axis
	if abs(normal.X) > abs(normal.Y) {
		// X-dominant
		if normal.X > 0 {
			return "right"
		} else {
			return "left"
		}
	} else {
		// Y-dominant
		if normal.Y > 0 {
			return "down"
		} else {
			return "up"
		}
	}
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// generateCollidersForGroup generates colliders for a group of wall points with the same normal.
// It merges adjacent wall points into larger rectangular colliders for better performance.
func (wcg *WallColliderGenerator) generateCollidersForGroup(points []WallPoint, normalKey string, cellSize float64) []RectCollider {
	var colliders []RectCollider

	// Sort points to make merging adjacent points easier
	sortedPoints := wcg.sortPointsByNormal(points, normalKey)

	// Merge adjacent points into larger rectangles
	var currentRect *rect

	for _, wp := range sortedPoints {
		if currentRect == nil {
			// Start a new rectangle
			currentRect = &rect{
				minX: wp.X - wcg.minWallSize/2,
				maxX: wp.X + wcg.minWallSize/2,
				minY: wp.Y - wcg.minWallSize/2,
				maxY: wp.Y + wcg.minWallSize/2,
			}
		} else {
			// Check if this point can be merged with the current rectangle
			if wcg.canMergePoint(wp, currentRect, normalKey) {
				// Expand the current rectangle to include this point
				wcg.expandRect(currentRect, wp)
			} else {
				// Create a collider from the current rectangle
				colliders = append(colliders, wcg.rectToCollider(currentRect))

				// Start a new rectangle with this point
				currentRect = &rect{
					minX: wp.X - wcg.minWallSize/2,
					maxX: wp.X + wcg.minWallSize/2,
					minY: wp.Y - wcg.minWallSize/2,
					maxY: wp.Y + wcg.minWallSize/2,
				}
			}
		}
	}

	// Don't forget the last rectangle
	if currentRect != nil {
		colliders = append(colliders, wcg.rectToCollider(currentRect))
	}

	return colliders
}

// rect represents a rectangle with min/max coordinates
type rect struct {
	minX, maxX, minY, maxY float64
}

// sortPointsByNormal sorts wall points to make merging adjacent points easier
func (wcg *WallColliderGenerator) sortPointsByNormal(points []WallPoint, normalKey string) []WallPoint {
	// Create a copy of the points slice to avoid modifying the original
	sortedPoints := make([]WallPoint, len(points))
	copy(sortedPoints, points)

	// Sort based on the normal direction
	switch normalKey {
	case "up", "down":
		// Sort horizontally (by X) for vertical walls
		for i := 0; i < len(sortedPoints); i++ {
			for j := i + 1; j < len(sortedPoints); j++ {
				if sortedPoints[i].X > sortedPoints[j].X {
					sortedPoints[i], sortedPoints[j] = sortedPoints[j], sortedPoints[i]
				}
			}
		}
	case "left", "right":
		// Sort vertically (by Y) for horizontal walls
		for i := 0; i < len(sortedPoints); i++ {
			for j := i + 1; j < len(sortedPoints); j++ {
				if sortedPoints[i].Y > sortedPoints[j].Y {
					sortedPoints[i], sortedPoints[j] = sortedPoints[j], sortedPoints[i]
				}
			}
		}
	}

	return sortedPoints
}

// canMergePoint checks if a point can be merged with the current rectangle
func (wcg *WallColliderGenerator) canMergePoint(wp WallPoint, r *rect, normalKey string) bool {
	const mergeThreshold = 2.0 // Maximum distance for merging points

	switch normalKey {
	case "up", "down":
		// For vertical walls, check if the X coordinate is close enough
		return wp.Y >= r.minY-mergeThreshold && wp.Y <= r.maxY+mergeThreshold &&
			(wp.X >= r.minX-mergeThreshold && wp.X <= r.maxX+mergeThreshold)
	case "left", "right":
		// For horizontal walls, check if the Y coordinate is close enough
		return wp.X >= r.minX-mergeThreshold && wp.X <= r.maxX+mergeThreshold &&
			(wp.Y >= r.minY-mergeThreshold && wp.Y <= r.maxY+mergeThreshold)
	}

	return false
}

// expandRect expands a rectangle to include a new point
func (wcg *WallColliderGenerator) expandRect(r *rect, wp WallPoint) {
	// Expand the rectangle to include the point
	r.minX = min(r.minX, wp.X-wcg.minWallSize/2)
	r.maxX = max(r.maxX, wp.X+wcg.minWallSize/2)
	r.minY = min(r.minY, wp.Y-wcg.minWallSize/2)
	r.maxY = max(r.maxY, wp.Y+wcg.minWallSize/2)
}

// rectToCollider converts a rectangle to a RectCollider
func (wcg *WallColliderGenerator) rectToCollider(r *rect) RectCollider {
	width := r.maxX - r.minX
	height := r.maxY - r.minY

	return RectCollider{
		Position: math.Vector{
			X: (r.minX + r.maxX) / 2, // Center X
			Y: (r.minY + r.maxY) / 2, // Center Y
		},
		Width:  width,
		Height: height,
	}
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// WallPoint represents a point on a wall with position and normal vector.
type WallPoint struct {
	X, Y   float64     // Coordinates
	Normal math.Vector // Normal vector of the wall (points away from the solid part)
}
