package worldgen

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
)

// ChunkSize defines the number of cells in a chunk (width and height)
const ChunkSize = 4

// WorldChunk represents a chunk of the world map containing multiple cells
type WorldChunk struct {
	X, Y     int                    // Chunk coordinates
	Cells    map[string]*WorldCell  // Map of cells in this chunk by local coordinates
	IsLoaded bool                   // Whether this chunk is currently loaded
}

// NewWorldChunk creates a new empty world chunk at the specified coordinates
func NewWorldChunk(x, y int) *WorldChunk {
	return &WorldChunk{
		X:        x,
		Y:        y,
		Cells:    make(map[string]*WorldCell),
		IsLoaded: false,
	}
}

// GetKey returns a unique string key for this chunk based on its coordinates
func (c *WorldChunk) GetKey() string {
	return fmt.Sprintf("%d,%d", c.X, c.Y)
}

// AddCell adds a cell to this chunk
func (c *WorldChunk) AddCell(cell *WorldCell) {
	// Calculate local coordinates within the chunk
	localX := cell.X - c.X*ChunkSize
	localY := cell.Y - c.Y*ChunkSize

	// Add to cells map using local coordinates as key
	key := fmt.Sprintf("%d,%d", localX, localY)
	c.Cells[key] = cell
}

// GetCell returns the cell at the specified local coordinates, or nil if no cell exists there
func (c *WorldChunk) GetCell(localX, localY int) *WorldCell {
	key := fmt.Sprintf("%d,%d", localX, localY)
	return c.Cells[key]
}

// GetCellCount returns the number of cells in this chunk
func (c *WorldChunk) GetCellCount() int {
	return len(c.Cells)
}

// Load loads all cells in this chunk
func (c *WorldChunk) Load() {
	c.IsLoaded = true
}

// Unload unloads all cells in this chunk
func (c *WorldChunk) Unload() {
	c.IsLoaded = false
}

// Draw draws all cells in this chunk
func (c *WorldChunk) Draw(screen *ebiten.Image, worldX, worldY int, cellSize int) {
	if !c.IsLoaded {
		return
	}

	// Draw each cell in the chunk
	for _, cell := range c.Cells {
		// Calculate local coordinates within the chunk
		localX := cell.X - c.X*ChunkSize
		localY := cell.Y - c.Y*ChunkSize

		// Calculate screen coordinates for this cell
		screenX := worldX + (localX * cellSize)
		screenY := worldY + (localY * cellSize)

		// Create draw options
		op := &ebiten.DrawImageOptions{}

		// Get the dimensions of the image
		w, h := float64(cell.Snippet.Image.Bounds().Dx()), float64(cell.Snippet.Image.Bounds().Dy())

		// Apply rotation first (if needed)
		if cell.Rotation != 0 {
			// Rotate around the center of the image
			op.GeoM.Translate(-w/2, -h/2)
			op.GeoM.Rotate(float64(cell.Rotation) * (3.14159 / 180.0))
			op.GeoM.Translate(w/2, h/2)
		}

		// Apply scaling after rotation
		op.GeoM.Scale(2.0, 2.0)

		// Apply translation last, adjusting for the scaling factor
		// Since we're scaling by 2.0, we need to position each cell exactly at cellSize intervals
		// to avoid overlaps or gaps
		op.GeoM.Translate(float64(screenX)*2.0, float64(screenY)*2.0)

		// Draw the snippet image
		screen.DrawImage(cell.Snippet.Image, op)
	}
}
