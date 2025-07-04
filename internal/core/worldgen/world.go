// Package worldgen implements procedural world generation for the game.
// It creates vast, explorable game worlds with varied terrain, paths, and
// environmental features using algorithmic generation rather than manual design.
//
// The world generation system is built on a multi-level architecture:
// - WorldMap: The high-level representation of the entire world
// - WorldCell: Individual cells that make up the world grid
// - WorldChunk: Groups of cells for efficient memory management and rendering
// - WorldSnippet: Visual elements and gameplay objects within cells
//
// Key features of the world generation system:
// - Procedural generation using configurable algorithms
// - Chunking system for performance optimization
// - Dynamic loading/unloading based on player position
// - Multi-level coordinate system (world, cell, chunk, local)
// - Path generation to ensure navigable worlds
//
// The system is designed to create worlds that are:
// - Unique for each playthrough
// - Efficiently rendered on both desktop and mobile devices
// - Interesting to explore with varied environments
// - Balanced for gameplay with appropriate challenge distribution
package worldgen

import (
	"discoveryx/internal/core/ecs"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
)

// CellSize defines the size of a cell in pixels.
// This constant determines the granularity of the world grid.
// Larger values create bigger cells, reducing the number of cells needed
// to represent the world but potentially reducing detail.
const CellSize = 1000

// VisibilityRadius defines how many chunks around the player should be visible.
// This constant controls the draw distance and memory usage:
// - Higher values show more of the world but consume more memory
// - Lower values improve performance but limit visibility
// The value represents the number of chunks in each direction (a radius),
// so the actual visible area is a square with sides of (2*VisibilityRadius+1) chunks.
const VisibilityRadius = 4

// GeneratedWorld implements the ecs.World interface and provides access to the generated world map.
// It serves as the central manager for the procedurally generated game world,
// handling world creation, chunk management, coordinate transformations,
// and dynamic loading/unloading of world sections based on player position.
//
// The GeneratedWorld uses a multi-level coordinate system:
// - World coordinates: Raw pixel positions in the game world
// - Cell coordinates: World coordinates divided by CellSize
// - Chunk coordinates: Cell coordinates divided by ChunkSize
// - Local coordinates: Cell positions within a specific chunk
//
// This chunking approach optimizes memory usage and rendering performance
// by only keeping nearby chunks loaded in memory, while distant chunks
// are unloaded until the player approaches them.
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

// NewGeneratedWorld creates a new generated world with the specified dimensions.
// This factory function initializes the world generation process, creating a
// complete procedural world based on the provided generator and configuration.
//
// The generation process follows these steps:
// 1. Initialize the GeneratedWorld struct with basic properties
// 2. Generate the complete world map using the provided generator
// 3. Place the player at a suitable starting position (middle of main path)
// 4. Organize cells into chunks for efficient rendering
// 5. Load chunks around the player's starting position
//
// Parameters:
// - width, height: The dimensions of the game world in pixels
// - generator: The algorithm to use for world generation
// - config: Configuration parameters for the generation process
//
// Returns:
// - A fully initialized GeneratedWorld ready for gameplay
// - An error if world generation fails
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

	err := world.GenerateNewWorld()
	if err != nil {
		return nil, err
	}

	// Set player position to the middle of the main path
	if len(world.worldMap.MainPathCells) > 0 {
		mainPathCell := world.worldMap.MainPathCells[len(world.worldMap.MainPathCells)/2]
		world.playerX = float64(mainPathCell.X*CellSize + CellSize/2)
		world.playerY = float64(mainPathCell.Y*CellSize + CellSize/2)
	}

	world.organizeChunks()
	world.UpdateVisibleChunks()

	return world, nil
}

// GenerateNewWorld creates a new world map using the configured generator
func (w *GeneratedWorld) GenerateNewWorld() error {
	var err error
	w.worldMap, err = w.generator.GenerateWorld(w.config)
	if err != nil {
		return err
	}
	return nil
}

// organizeChunks groups cells into manageable chunks for efficient rendering
func (w *GeneratedWorld) organizeChunks() {
	w.chunks = make(map[string]*WorldChunk)

	for _, cell := range w.worldMap.Cells {
		chunkX := cell.X / ChunkSize
		chunkY := cell.Y / ChunkSize

		chunkKey := getChunkKey(chunkX, chunkY)
		chunk, exists := w.chunks[chunkKey]
		if !exists {
			chunk = NewWorldChunk(chunkX, chunkY)
			w.chunks[chunkKey] = chunk
		}

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

// UpdateVisibleChunks loads chunks within the visibility radius of the player.
// This method is a core part of the dynamic loading system that optimizes
// memory usage and rendering performance by:
// 1. Unloading all currently loaded chunks
// 2. Calculating which chunks are within the visibility radius of the player
// 3. Loading only those chunks that should be visible
//
// This approach ensures that:
// - Only relevant parts of the world consume memory
// - Rendering is focused on visible areas
// - The game can support very large worlds without performance issues
//
// This method is automatically called when the player's position changes
// and should be called manually if the visibility radius is modified.
func (w *GeneratedWorld) UpdateVisibleChunks() {
	playerCellX := int(w.playerX) / CellSize
	playerCellY := int(w.playerY) / CellSize
	playerChunkX := playerCellX / ChunkSize
	playerChunkY := playerCellY / ChunkSize

	for _, chunk := range w.chunks {
		chunk.Unload()
	}

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

// GetCellAt returns the world cell at the specified world coordinates, or nil if no cell exists there.
// This method handles the complex coordinate transformations required to:
// 1. Convert from world coordinates (pixels) to cell coordinates
// 2. Determine which chunk contains the requested cell
// 3. Calculate local coordinates within that chunk
// 4. Retrieve the specific cell if it exists and is loaded
//
// The method returns nil in three cases:
// - No chunk exists at the calculated chunk coordinates
// - The chunk exists but is not currently loaded
// - The chunk is loaded but contains no cell at the local coordinates
//
// This coordinate transformation system allows the game to efficiently
// locate cells in a potentially infinite world without searching through
// all cells sequentially.
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

// Draw renders all loaded chunks with the given camera offset.
// This method is responsible for visualizing the world by:
// 1. Calculating the screen position for each loaded chunk
// 2. Applying camera offsets to implement scrolling
// 3. Delegating rendering to each chunk's Draw method
//
// The rendering process uses these coordinate transformations:
// - Start with the center of the screen (centerX, centerY)
// - Add the chunk's position in world space (chunk.X*ChunkSize*CellSize)
// - Apply camera offsets for scrolling (offsetX, offsetY)
//
// Only chunks that are currently loaded (within the visibility radius
// of the player) are rendered, improving performance by avoiding
// unnecessary drawing operations.
func (w *GeneratedWorld) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	centerX := float64(w.width) / 2
	centerY := float64(w.height) / 2

	for _, chunk := range w.chunks {
		if chunk.IsLoaded {
			worldX := centerX + float64(chunk.X*ChunkSize*CellSize) + offsetX
			worldY := centerY + float64(chunk.Y*ChunkSize*CellSize) + offsetY

			chunk.Draw(screen, int(worldX), int(worldY), CellSize)
		}
	}
}

// Ensure GeneratedWorld implements the ecs.World interface
var _ ecs.World = (*GeneratedWorld)(nil)
