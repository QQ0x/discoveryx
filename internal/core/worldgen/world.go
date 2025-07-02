package worldgen

import (
	"discoveryx/internal/core/ecs"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
)

// CellSize defines the size of a cell in pixels
const CellSize = 1000

// VisibilityRadius defines how many chunks around the player should be visible
const VisibilityRadius = 4

// GeneratedWorld implements the ecs.World interface and provides access to the generated world map
type GeneratedWorld struct {
	width       int
	height      int
	matchScreen bool
	worldMap    *WorldMap
	chunks      map[string]*WorldChunk
	generator   *WorldGenerator
	config      *WorldGenConfig
	playerX     float64
	playerY     float64
}

// NewGeneratedWorld creates a new generated world with the specified dimensions
func NewGeneratedWorld(width, height int, generator *WorldGenerator, config *WorldGenConfig) (*GeneratedWorld, error) {
	world := &GeneratedWorld{
		width:       width,
		height:      height,
		matchScreen: false,
		generator:   generator,
		config:      config,
		chunks:      make(map[string]*WorldChunk),
		playerX:     0,
		playerY:     0,
	}

	// Generate the initial world map
	err := world.GenerateNewWorld()
	if err != nil {
		return nil, err
	}

	// Set player position to a cell on the main path
	// This satisfies requirement 8: The player must spawn on the main path
	if len(world.worldMap.MainPathCells) > 0 {
		// Choose a cell from the main path
		mainPathCell := world.worldMap.MainPathCells[len(world.worldMap.MainPathCells)/2] // Use middle of the path for better gameplay

		// Convert cell coordinates to world coordinates
		world.playerX = float64(mainPathCell.X * CellSize + CellSize/2)
		world.playerY = float64(mainPathCell.Y * CellSize + CellSize/2)
	}

	// Organize cells into chunks
	world.organizeChunks()

	// Load initial chunks around the player
	world.UpdateVisibleChunks()

	return world, nil
}

// GenerateNewWorld generates a new world map using the configured generator and config
func (w *GeneratedWorld) GenerateNewWorld() error {
	var err error
	w.worldMap, err = w.generator.GenerateWorld(w.config)
	if err != nil {
		return err
	}

	// Spawn objects on walls
	objectTypes := []string{"Pilz", "Kristall", "Flechte", "Spinnennetz"}
	w.SpawnObjectsOnWalls(objectTypes, 0.05, 10.0) // 5% chance per wall, minimum distance 10 units

	return nil
}

// organizeChunks organizes the cells in the world map into chunks
func (w *GeneratedWorld) organizeChunks() {
	// Clear existing chunks
	w.chunks = make(map[string]*WorldChunk)

	// Organize cells into chunks
	for _, cell := range w.worldMap.Cells {
		// Calculate chunk coordinates
		chunkX := cell.X / ChunkSize
		chunkY := cell.Y / ChunkSize

		// Get or create chunk
		chunkKey := getChunkKey(chunkX, chunkY)
		chunk, exists := w.chunks[chunkKey]
		if !exists {
			chunk = NewWorldChunk(chunkX, chunkY)
			w.chunks[chunkKey] = chunk
		}

		// Add cell to chunk
		chunk.AddCell(cell)
	}
}

// getChunkKey returns a unique string key for a chunk based on its coordinates
func getChunkKey(x, y int) string {
	return fmt.Sprintf("%d,%d", x, y)
}

// GetChunk returns the chunk at the specified chunk coordinates, or nil if no chunk exists there
func (w *GeneratedWorld) GetChunk(chunkX, chunkY int) *WorldChunk {
	key := getChunkKey(chunkX, chunkY)
	return w.chunks[key]
}

// GetChunkAt returns the chunk that contains the specified world coordinates, or nil if no chunk exists there
func (w *GeneratedWorld) GetChunkAt(worldX, worldY int) *WorldChunk {
	// Convert world coordinates to cell coordinates
	cellX := worldX / CellSize
	cellY := worldY / CellSize

	// Convert cell coordinates to chunk coordinates
	chunkX := cellX / ChunkSize
	chunkY := cellY / ChunkSize

	return w.GetChunk(chunkX, chunkY)
}

// UpdateVisibleChunks updates which chunks are loaded based on the player's position
func (w *GeneratedWorld) UpdateVisibleChunks() {
	// Convert player position to chunk coordinates
	playerCellX := int(w.playerX) / CellSize
	playerCellY := int(w.playerY) / CellSize
	playerChunkX := playerCellX / ChunkSize
	playerChunkY := playerCellY / ChunkSize

	// Unload all chunks first
	for _, chunk := range w.chunks {
		chunk.Unload()
	}

	// Load chunks within visibility radius
	for y := playerChunkY - VisibilityRadius; y <= playerChunkY+VisibilityRadius; y++ {
		for x := playerChunkX - VisibilityRadius; x <= playerChunkX+VisibilityRadius; x++ {
			chunk := w.GetChunk(x, y)
			if chunk != nil {
				chunk.Load()
			}
		}
	}
}

// SetPlayerPosition sets the player's position in the world
func (w *GeneratedWorld) SetPlayerPosition(x, y float64) {
	w.playerX = x
	w.playerY = y
	w.UpdateVisibleChunks()
}

// GetPlayerPosition returns the player's position in the world
func (w *GeneratedWorld) GetPlayerPosition() (float64, float64) {
	return w.playerX, w.playerY
}

// GetWorldMap returns the generated world map
func (w *GeneratedWorld) GetWorldMap() *WorldMap {
	return w.worldMap
}

// GetWidth returns the width of the game world
func (w *GeneratedWorld) GetWidth() int {
	return w.width
}

// GetHeight returns the height of the game world
func (w *GeneratedWorld) GetHeight() int {
	return w.height
}

// SetWidth sets the width of the game world
func (w *GeneratedWorld) SetWidth(width int) {
	w.width = width
}

// SetHeight sets the height of the game world
func (w *GeneratedWorld) SetHeight(height int) {
	w.height = height
}

// ShouldMatchScreen returns true if the world dimensions should
// automatically match the screen dimensions
func (w *GeneratedWorld) ShouldMatchScreen() bool {
	return w.matchScreen
}

// SetMatchScreen sets whether the world dimensions should
// automatically match the screen dimensions
func (w *GeneratedWorld) SetMatchScreen(match bool) {
	w.matchScreen = match
}

// GetCellAt returns the world cell at the specified world coordinates, or nil if no cell exists there
func (w *GeneratedWorld) GetCellAt(worldX, worldY int) *WorldCell {
	// Convert world coordinates to cell coordinates
	cellX := worldX / CellSize
	cellY := worldY / CellSize

	// Get the chunk that contains this cell
	chunkX := cellX / ChunkSize
	chunkY := cellY / ChunkSize
	chunk := w.GetChunk(chunkX, chunkY)

	if chunk == nil || !chunk.IsLoaded {
		return nil
	}

	// Calculate local coordinates within the chunk
	localX := cellX - chunkX*ChunkSize
	localY := cellY - chunkY*ChunkSize

	return chunk.GetCell(localX, localY)
}

// GetSnippetAt returns the world snippet at the specified world coordinates, or nil if no snippet exists there
func (w *GeneratedWorld) GetSnippetAt(worldX, worldY int) *WorldSnippet {
	cell := w.GetCellAt(worldX, worldY)
	if cell == nil {
		return nil
	}
	return cell.Snippet
}

// Draw draws the visible chunks of the world
func (w *GeneratedWorld) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	// Calculate the center of the screen
	centerX := float64(w.width) / 2
	centerY := float64(w.height) / 2

	// Calculate the camera center position in world coordinates
	cameraCenterX := -offsetX
	cameraCenterY := -offsetY

	// Draw each loaded chunk
	for _, chunk := range w.chunks {
		if chunk.IsLoaded {
			// Calculate the world coordinates for this chunk
			// Use camera position instead of player position for smoother movement
			worldX := centerX + offsetX + float64(chunk.X * ChunkSize * CellSize) - cameraCenterX
			worldY := centerY + offsetY + float64(chunk.Y * ChunkSize * CellSize) - cameraCenterY

			// Draw the chunk - convert final position to int to avoid subpixel rendering issues
			chunk.Draw(screen, int(worldX), int(worldY), CellSize)
		}
	}
}

// Ensure GeneratedWorld implements the ecs.World interface
var _ ecs.World = (*GeneratedWorld)(nil)
