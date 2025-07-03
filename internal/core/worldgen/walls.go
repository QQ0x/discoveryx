package worldgen

import (
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
	"sync"
)

// WallPoint represents a point on a wall with world coordinates
type WallPoint struct {
	X, Y   float64     // Coordinates
	Normal math.Vector // Normal vector of the wall (points away from the rock)
}

// wallCache stores detected walls for snippets to avoid reprocessing unchanged snippets
var wallCache = struct {
	sync.RWMutex
	data map[string][]WallPoint
}{
	data: make(map[string][]WallPoint),
}

// DetectWallsInSnippet detects all wall points within a snippet
// If the snippet hasn't changed (based on filename), it returns cached wall data
func DetectWallsInSnippet(snippet *WorldSnippet) []WallPoint {
	// Check if we have this snippet's walls in the cache
	wallCache.RLock()
	cachedWalls, exists := wallCache.data[snippet.Filename]
	wallCache.RUnlock()

	// If walls exist in cache, return them
	if exists {
		return cachedWalls
	}

	// Otherwise, detect walls
	walls := []WallPoint{}

	// Get image dimensions
	width, height := snippet.Image.Bounds().Dx(), snippet.Image.Bounds().Dy()

	// Create temporary RGBA image to extract pixel data
	rgba := ebiten.NewImage(width, height)
	op := &ebiten.DrawImageOptions{}
	rgba.DrawImage(snippet.Image, op)

	// Get pixel data
	imgData := make([]byte, width*height*4)
	rgba.ReadPixels(imgData)

	// Directions for neighboring pixels (top, right, bottom, left)
	directions := [][2]int{
		{0, -1}, // Top
		{1, 0},  // Right
		{0, 1},  // Bottom
		{-1, 0}, // Left
	}

	// Check each pixel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get alpha value of current pixel (4th byte in RGBA format)
			idx := (y*width + x) * 4
			alpha := imgData[idx+3]

			// If pixel is not transparent (rock)
			if alpha > 0 {
				// Check all neighboring pixels
				for _, dir := range directions {
					nx, ny := x+dir[0], y+dir[1]

					// Check if neighbor is within image bounds
					if nx >= 0 && nx < width && ny >= 0 && ny < height {
						// Get alpha value of neighboring pixel
						nIdx := (ny*width + nx) * 4
						nAlpha := imgData[nIdx+3]

						// If neighbor is transparent, we found a wall
						if nAlpha == 0 {
							// Calculate wall position (between the two pixels)
							wallX := float64(x) + float64(dir[0])*0.5
							wallY := float64(y) + float64(dir[1])*0.5

							// Calculate normal vector (points away from rock)
							normal := math.Vector{
								X: float64(dir[0]),
								Y: float64(dir[1]),
							}

							// Add wall point
							walls = append(walls, WallPoint{
								X:      wallX,
								Y:      wallY,
								Normal: normal,
							})
						}
					}
				}
			}
 	}
	}

	// Store the detected walls in the cache
	wallCache.Lock()
	wallCache.data[snippet.Filename] = walls
	wallCache.Unlock()

	return walls
}

// GetWallsInWorldCoordinates calculates the world coordinates of all wall points in a cell
func (cell *WorldCell) GetWallsInWorldCoordinates() []WallPoint {
	if cell.Snippet == nil || len(cell.Snippet.Walls) == 0 {
		return nil
	}

	worldWalls := make([]WallPoint, len(cell.Snippet.Walls))

	// Calculate world position of cell
	cellWorldX := cell.X * CellSize
	cellWorldY := cell.Y * CellSize

	// For each wall in the snippet
	for i, wall := range cell.Snippet.Walls {
		// Copy the wall
		worldWall := wall

		// Apply rotation if needed
		if cell.Rotation != 0 {
			// Rotation angle in radians
			angle := float64(cell.Rotation) * (stdmath.Pi / 180.0)

			// Apply rotation matrix
			cosA := stdmath.Cos(angle)
			sinA := stdmath.Sin(angle)

			// Rotate around snippet center
			snippetCenterX := float64(CellSize) / 2
			snippetCenterY := float64(CellSize) / 2

			// Point relative to center
			relX := worldWall.X - snippetCenterX
			relY := worldWall.Y - snippetCenterY

			// Apply rotation
			rotX := relX*cosA - relY*sinA
			rotY := relX*sinA + relY*cosA

			// Back to absolute point
			worldWall.X = rotX + snippetCenterX
			worldWall.Y = rotY + snippetCenterY

			// Also rotate normal vector
			// Store original normal vector components
			origNormalX := worldWall.Normal.X
			origNormalY := worldWall.Normal.Y

			// Apply rotation matrix to normal vector
			normalX := origNormalX*cosA - origNormalY*sinA
			normalY := origNormalX*sinA + origNormalY*cosA
			worldWall.Normal = math.Vector{X: normalX, Y: normalY}
		}

		// Calculate world coordinates
		worldWall.X += float64(cellWorldX)
		worldWall.Y += float64(cellWorldY)

		worldWalls[i] = worldWall
	}

	return worldWalls
}
