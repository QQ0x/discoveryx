package worldgen

import (
	"discoveryx/internal/utils/math"
	"github.com/hajimehoshi/ebiten/v2"
	stdmath "math"
	"math/rand"
	"time"
)

// WallPoint represents a point on a wall with world coordinates
type WallPoint struct {
	X, Y   float64     // Coordinates
	Normal math.Vector // Normal vector of the wall (points away from the rock)
}

// DetectWallsInSnippet detects all wall points within a snippet
func DetectWallsInSnippet(snippet *WorldSnippet) []WallPoint {
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
			normalX := worldWall.Normal.X*cosA - worldWall.Normal.Y*sinA
			normalY := worldWall.Normal.X*sinA + worldWall.Normal.Y*cosA
			worldWall.Normal = math.Vector{X: normalX, Y: normalY}
		}

		// Calculate world coordinates
		worldWall.X += float64(cellWorldX)
		worldWall.Y += float64(cellWorldY)

		worldWalls[i] = worldWall
	}

	return worldWalls
}

// SpawnObjectsOnWalls spawns objects on walls in the visible world
func (w *GeneratedWorld) SpawnObjectsOnWalls(objectTypes []string, spawnChance float64, minDistanceBetweenObjects float64) {
	// Random generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// List of all spawned objects to ensure minimum distance
	spawnedObjects := []math.Vector{}

	// For each visible chunk
	for _, chunk := range w.chunks {
		if !chunk.IsLoaded {
			continue
		}

		// For each cell in the chunk
		for _, cell := range chunk.Cells {
			// Get wall points in world coordinates
			wallPoints := cell.GetWallsInWorldCoordinates()

			// For each wall point
			for _, wall := range wallPoints {
				// Check if an object should be spawned (random chance)
				if rng.Float64() <= spawnChance {
					// Check minimum distance to other spawned objects
					tooClose := false
					for _, obj := range spawnedObjects {
						dx := wall.X - obj.X
						dy := wall.Y - obj.Y
						distSq := dx*dx + dy*dy

						if distSq < minDistanceBetweenObjects*minDistanceBetweenObjects {
							tooClose = true
							break
						}
					}

					if !tooClose {
						// Choose random object type
						objectType := objectTypes[rng.Intn(len(objectTypes))]

						// Spawn object at wall position
						// Offset slightly in direction of normal vector so it sits on the wall
						spawnX := wall.X + wall.Normal.X*0.1
						spawnY := wall.Y + wall.Normal.Y*0.1

						// Log what type of object we're spawning (in a real implementation, we would use this)
						_ = objectType // Using objectType to avoid unused variable warning

						// Here would be the actual object creation
						// For now, just add to the list of spawned objects
						spawnedObjects = append(spawnedObjects, math.Vector{X: spawnX, Y: spawnY})

						// TODO: Create actual game object here
						// This would depend on how objects are created in the game
					}
				}
			}
		}
	}
}
