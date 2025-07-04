package worldgen

import (
	"fmt"
)

// WorldCell represents a single cell in the world map
type WorldCell struct {
	X, Y        int           // The coordinates of the cell in the world grid
	Snippet     *WorldSnippet // The snippet placed at this cell
	Rotation    int           // The rotation of the snippet in degrees (0, 90, 180, 270)
	IsMainPath  bool          // Whether this cell is part of the main path
	BranchDepth int           // The depth of this cell in a branch (0 for main path)
}

// GetKey returns a unique string key for this cell based on its coordinates
func (c *WorldCell) GetKey() string {
	return fmt.Sprintf("%d,%d", c.X, c.Y)
}

// GetRotatedConnectors returns the snippet connectors adjusted for cell rotation
func (c *WorldCell) GetRotatedConnectors() []SnippetConnector {
	if c.Rotation == 0 {
		return c.Snippet.Connectors
	}

	rotated := make([]SnippetConnector, len(c.Snippet.Connectors))
	for i, conn := range c.Snippet.Connectors {
		rotated[i] = (conn + SnippetConnector(c.Rotation)) % 360
	}
	return rotated
}

// WorldMap represents the generated world map
type WorldMap struct {
	Cells         map[string]*WorldCell // Map of cells by coordinates (key: "x,y")
	MainPathCells []*WorldCell          // List of cells that form the main path
	BranchCells   []*WorldCell          // List of cells that form branches
}

// NewWorldMap creates a new empty world map
func NewWorldMap() *WorldMap {
	return &WorldMap{
		Cells:         make(map[string]*WorldCell),
		MainPathCells: make([]*WorldCell, 0),
		BranchCells:   make([]*WorldCell, 0),
	}
}

// AddCell adds a cell to the world map
func (m *WorldMap) AddCell(cell *WorldCell) {
	key := cell.GetKey()
	m.Cells[key] = cell

	if cell.IsMainPath {
		m.MainPathCells = append(m.MainPathCells, cell)
	} else {
		m.BranchCells = append(m.BranchCells, cell)
	}
}

// GetCell returns the cell at the specified coordinates, or nil if no cell exists there
func (m *WorldMap) GetCell(x, y int) *WorldCell {
	key := fmt.Sprintf("%d,%d", x, y)
	return m.Cells[key]
}

// HasCell returns true if there is a cell at the specified coordinates
func (m *WorldMap) HasCell(x, y int) bool {
	key := fmt.Sprintf("%d,%d", x, y)
	_, exists := m.Cells[key]
	return exists
}

// GetAdjacentCells returns cells in the four cardinal directions (top, right, bottom, left)
func (m *WorldMap) GetAdjacentCells(x, y int) []*WorldCell {
	adjacent := make([]*WorldCell, 0, 4)

	directions := [][2]int{
		{0, -1}, // Top
		{1, 0},  // Right
		{0, 1},  // Bottom
		{-1, 0}, // Left
	}

	for _, dir := range directions {
		nx, ny := x+dir[0], y+dir[1]
		if cell := m.GetCell(nx, ny); cell != nil {
			adjacent = append(adjacent, cell)
		}
	}

	return adjacent
}

// GetCellCount returns the total number of cells in the world map
func (m *WorldMap) GetCellCount() int {
	return len(m.Cells)
}

// GetMainPathLength returns the length of the main path
func (m *WorldMap) GetMainPathLength() int {
	return len(m.MainPathCells)
}

// GetBranchCount returns the number of branches
func (m *WorldMap) GetBranchCount() int {
	return len(m.BranchCells)
}
