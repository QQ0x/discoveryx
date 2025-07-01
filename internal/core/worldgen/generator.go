package worldgen

import (
	"fmt"
	"math/rand"
	"time"
)

// WorldGenConfig represents the configuration for world generation
type WorldGenConfig struct {
	MainPathMinLength    int                 // Minimum number of snippets in main path
	MainPathMaxLength    int                 // Maximum number of snippets in main path
	BranchProbability    float64             // Probability of creating a branch (0.0-1.0)
	BranchMaxDepth       int                 // Maximum branch depth from main path
	DeadEndProbability   float64             // Probability of branch ending in dead-end
	SnippetTypeWeights   map[SnippetType]int // Relative weights for snippet types
	SubBranchProbability float64             // Probability of branches from branches
	Seed                 int64               // Random seed for reproducible generation
}

// DefaultWorldGenConfig returns a default configuration for world generation
func DefaultWorldGenConfig() *WorldGenConfig {
	return &WorldGenConfig{
		MainPathMinLength:  20,
		MainPathMaxLength:  40,
		BranchProbability:  0.3,
		BranchMaxDepth:     5,
		DeadEndProbability: 0.5,
		SnippetTypeWeights: map[SnippetType]int{
			SnippetTypePath:     10, // Higher weight for path segments
			SnippetTypeJunction: 5,  // Medium weight for junctions
			SnippetTypeDeadEnd:  2,  // Lower weight for dead-ends
		},
		SubBranchProbability: 0.2,
		Seed:                 time.Now().UnixNano(),
	}
}

// GenerateWorld generates a new world map based on the provided configuration
func (g *WorldGenerator) GenerateWorld(config *WorldGenConfig) (*WorldMap, error) {
	// Initialize random number generator with seed
	rng := rand.New(rand.NewSource(config.Seed))

	// Create a new world map
	worldMap := NewWorldMap()

	// Generate the main path
	err := g.generateMainPath(worldMap, config, rng)
	if err != nil {
		return nil, fmt.Errorf("failed to generate main path: %w", err)
	}

	// Generate branches
	err = g.generateBranches(worldMap, config, rng)
	if err != nil {
		return nil, fmt.Errorf("failed to generate branches: %w", err)
	}

	// Post-process the world map to add borders and fill empty spaces
	err = g.postProcessWorldMap(worldMap, rng)
	if err != nil {
		return nil, fmt.Errorf("failed to post-process world map: %w", err)
	}

	return worldMap, nil
}

// generateMainPath generates the main circular path
func (g *WorldGenerator) generateMainPath(worldMap *WorldMap, config *WorldGenConfig, rng *rand.Rand) error {
	// Determine the length of the main path
	pathLength := config.MainPathMinLength
	if config.MainPathMaxLength > config.MainPathMinLength {
		pathLength += rng.Intn(config.MainPathMaxLength - config.MainPathMinLength + 1)
	}

	// Start with the first cell at (0,0)
	// For simplicity, we'll use a snippet with at least two connectors
	var startSnippet *WorldSnippet
	for _, snippet := range g.Registry.Snippets {
		if len(snippet.Connectors) >= 2 {
			startSnippet = snippet
			break
		}
	}

	if startSnippet == nil {
		return fmt.Errorf("no suitable starting snippet found")
	}

	// Create the first cell
	startCell := &WorldCell{
		X:           0,
		Y:           0,
		Snippet:     startSnippet,
		Rotation:    0, // No rotation for the first cell
		IsMainPath:  true,
		BranchDepth: 0,
	}

	worldMap.AddCell(startCell)

	// Keep track of the current position and direction
	currentX, currentY := 0, 0

	// Direction vectors for the four possible directions
	directions := [][2]int{
		{0, -1}, // Top (0 degrees)
		{1, 0},  // Right (90 degrees)
		{0, 1},  // Bottom (180 degrees)
		{-1, 0}, // Left (270 degrees)
	}

	// Plan a circular path before generating it
	// This ensures we can create a complete circle
	plannedPath := make([][2]int, 0, pathLength)
	plannedPath = append(plannedPath, [2]int{0, 0}) // Start position

	// Generate a spiral-like pattern to ensure we can close the loop
	// We'll start by going right, then down, then left, then up, and repeat
	// This creates a spiral pattern that can be closed easily
	spiralDirections := [][2]int{
		{1, 0},  // Right
		{0, 1},  // Down
		{-1, 0}, // Left
		{0, -1}, // Up
	}

	dirIndex := 0
	stepsInCurrentDir := 1
	stepsTaken := 0

	for i := 1; i < pathLength-1; i++ { // Leave room for the last cell to connect back
		dx, dy := spiralDirections[dirIndex][0], spiralDirections[dirIndex][1]
		nextX, nextY := currentX+dx, currentY+dy

		// Add the position to the planned path
		plannedPath = append(plannedPath, [2]int{nextX, nextY})

		// Update current position
		currentX, currentY = nextX, nextY

		// Update step counter
		stepsTaken++

		// Check if we need to change direction
		if stepsTaken == stepsInCurrentDir {
			dirIndex = (dirIndex + 1) % 4
			stepsTaken = 0

			// Increase steps after completing a half circle
			if dirIndex%2 == 0 {
				stepsInCurrentDir++
			}
		}
	}

	// Now add the last cell that connects back to the start
	// Calculate the direction from the last position to the start
	dx, dy := 0-currentX, 0-currentY

	// Normalize to get a single step
	if dx != 0 {
		dx = dx / abs(dx)
	}
	if dy != 0 {
		dy = dy / abs(dy)
	}

	// If we're more than one step away from the start in both x and y,
	// we need to take multiple steps to get back
	// For simplicity, we'll just add one more step in the x direction
	if abs(dx) + abs(dy) > 1 {
		nextX, nextY := currentX+dx, currentY
		plannedPath = append(plannedPath, [2]int{nextX, nextY})
		currentX, currentY = nextX, nextY

		// Recalculate direction to start
		dx, dy = 0-currentX, 0-currentY
		if dx != 0 {
			dx = dx / abs(dx)
		}
		if dy != 0 {
			dy = dy / abs(dy)
		}
	}

	// Add the final step that should be adjacent to the start
	if abs(dx) + abs(dy) == 1 {
		nextX, nextY := currentX+dx, currentY+dy
		plannedPath = append(plannedPath, [2]int{nextX, nextY})
	} else {
		// If we can't get back to the start with one step, try to find a valid position
		// adjacent to the start
		for _, dir := range directions {
			nx, ny := dir[0], dir[1]
			// Check if this position is already occupied
			positionOccupied := false
			for _, pos := range plannedPath {
				if pos[0] == nx && pos[1] == ny {
					positionOccupied = true
					break
				}
			}

			if !positionOccupied {
				plannedPath = append(plannedPath, [2]int{nx, ny})
				break
			}
		}
	}

	// Now generate the actual path following the planned positions
	currentX, currentY = 0, 0 // Reset to start position

	for i := 1; i < len(plannedPath); i++ {
		nextX, nextY := plannedPath[i][0], plannedPath[i][1]
		dx, dy := nextX-currentX, nextY-currentY

		// Determine the connector direction from the new cell back to the current cell
		var toConnector SnippetConnector
		if dx == 0 && dy == -1 {
			toConnector = ConnectorBottom
		} else if dx == 1 && dy == 0 {
			toConnector = ConnectorLeft
		} else if dx == 0 && dy == 1 {
			toConnector = ConnectorTop
		} else {
			toConnector = ConnectorRight
		}

		// Find a snippet that has the required connector
		snippets := g.Registry.GetSnippetsByConnector(toConnector)
		if len(snippets) == 0 {
			return fmt.Errorf("no snippets found with connector %d", toConnector)
		}

		// Choose a snippet based on weight and type
		snippet := SelectWeightedSnippetWithTypeWeights(snippets, config.SnippetTypeWeights, rng)

		// Determine the rotation needed
		rotation := 0
		for _, conn := range snippet.Connectors {
			if conn == toConnector {
				break
			}
			rotation += 90
			if rotation >= 360 {
				rotation = 0
			}
		}

		// Create the new cell
		newCell := &WorldCell{
			X:           nextX,
			Y:           nextY,
			Snippet:     snippet,
			Rotation:    rotation,
			IsMainPath:  true,
			BranchDepth: 0,
		}

		worldMap.AddCell(newCell)

		// Update the current position
		currentX, currentY = nextX, nextY
	}

	return nil
}

// generateBranches generates branches from the main path
func (g *WorldGenerator) generateBranches(worldMap *WorldMap, config *WorldGenConfig, rng *rand.Rand) error {
	// For each cell in the main path
	for _, cell := range worldMap.MainPathCells {
		// Check each direction from this cell
		for _, dir := range getDirections() {
			// Calculate the adjacent position
			nx, ny := cell.X+dir[0], cell.Y+dir[1]

			// Skip if there's already a cell there
			if worldMap.HasCell(nx, ny) {
				continue
			}

			// Decide whether to create a branch
			if rng.Float64() < config.BranchProbability {
				// Create a branch starting from this cell
				err := g.generateBranch(worldMap, cell, dir, 1, config, rng)
				if err != nil {
					return fmt.Errorf("failed to generate branch from (%d,%d): %w", cell.X, cell.Y, err)
				}
			}
		}
	}

	return nil
}

// generateBranch generates a single branch from the specified cell in the specified direction
func (g *WorldGenerator) generateBranch(worldMap *WorldMap, fromCell *WorldCell, direction [2]int, depth int, config *WorldGenConfig, rng *rand.Rand) error {
	// If we've reached the maximum branch depth, we must place a dead-end
	if depth > config.BranchMaxDepth {
		return nil
	}

	// Calculate the new position
	newX, newY := fromCell.X+direction[0], fromCell.Y+direction[1]

	// Skip if there's already a cell there
	if worldMap.HasCell(newX, newY) {
		return nil
	}

	// Determine if this branch should be a dead-end
	isDeadEnd := rng.Float64() < config.DeadEndProbability

	// We don't need to determine the connector from the current cell
	// as we're only concerned with the connector on the new cell
	// that connects back to the current cell

	// Determine the connector direction from the new cell back to the current cell
	var toConnector SnippetConnector
	if direction[0] == 0 && direction[1] == -1 {
		toConnector = ConnectorBottom
	} else if direction[0] == 1 && direction[1] == 0 {
		toConnector = ConnectorLeft
	} else if direction[0] == 0 && direction[1] == 1 {
		toConnector = ConnectorTop
	} else {
		toConnector = ConnectorRight
	}

	// Find a snippet that has the required connector
	snippets := g.Registry.GetSnippetsByConnector(toConnector)
	if len(snippets) == 0 {
		return fmt.Errorf("no snippets found with connector %d", toConnector)
	}

	// Filter snippets based on whether this should be a dead-end
	var filteredSnippets []*WorldSnippet
	if isDeadEnd || depth == config.BranchMaxDepth {
		// For dead-ends or at max depth, only use snippets with one connector
		for _, s := range snippets {
			if len(s.Connectors) == 1 {
				filteredSnippets = append(filteredSnippets, s)
			}
		}
		// If no single-connector snippets found, we can't create this branch
		if len(filteredSnippets) == 0 {
			// Instead of returning nil, we'll use a snippet with multiple connectors
			// but ensure all its connectors are connected to other snippets
			filteredSnippets = snippets
		}
	} else {
		filteredSnippets = snippets
	}

	// Choose a snippet based on weight and type
	snippet := SelectWeightedSnippetWithTypeWeights(filteredSnippets, config.SnippetTypeWeights, rng)

	// Determine the rotation needed
	rotation := 0
	for _, conn := range snippet.Connectors {
		if conn == toConnector {
			break
		}
		rotation += 90
		if rotation >= 360 {
			rotation = 0
		}
	}

	// Create the new cell
	newCell := &WorldCell{
		X:           newX,
		Y:           newY,
		Snippet:     snippet,
		Rotation:    rotation,
		IsMainPath:  false,
		BranchDepth: depth,
	}

	worldMap.AddCell(newCell)

	// If this is a dead-end or at max depth and has only one connector, don't continue
	if (isDeadEnd || depth == config.BranchMaxDepth) && len(snippet.Connectors) == 1 {
		return nil
	}

	// Check if we can connect back to the main path
	if depth > 1 {
		// Look for adjacent main path cells
		for _, dir := range getDirections() {
			// Skip the direction we came from
			if dir[0] == -direction[0] && dir[1] == -direction[1] {
				continue
			}

			// Calculate the adjacent position
			nx, ny := newX+dir[0], newY+dir[1]

			// Check if there's a main path cell there
			adjacentCell := worldMap.GetCell(nx, ny)
			if adjacentCell != nil && adjacentCell.IsMainPath {
				// We found a main path cell, we can connect to it
				// No need to continue the branch
				return nil
			}
		}
	}

	// Keep track of whether we've created any sub-branches
	createdSubBranch := false

	// Decide whether to continue the branch
	// For each direction from this cell
	for _, dir := range getDirections() {
		// Skip the direction we came from
		if dir[0] == -direction[0] && dir[1] == -direction[1] {
			continue
		}

		// Calculate the adjacent position
		nx, ny := newX+dir[0], newY+dir[1]

		// Skip if there's already a cell there
		if worldMap.HasCell(nx, ny) {
			continue
		}

		// Decide whether to create a sub-branch
		if rng.Float64() < config.SubBranchProbability {
			// Create a branch starting from this cell
			err := g.generateBranch(worldMap, newCell, dir, depth+1, config, rng)
			if err != nil {
				return fmt.Errorf("failed to generate sub-branch from (%d,%d): %w", newX, newY, err)
			}
			createdSubBranch = true
		}
	}

	// If this isn't a dead-end, we need to ensure all connectors are connected to something
	if !isDeadEnd && len(snippet.Connectors) > 1 {
		// Get the rotated connectors for this cell
		rotatedConnectors := newCell.GetRotatedConnectors()

		// For each connector
		for _, conn := range rotatedConnectors {
			// Skip the connector that connects back to the previous cell
			oppositeConnector := (toConnector + 180) % 360
			if conn == oppositeConnector {
				continue
			}

			// Calculate the direction for this connector
			var dir [2]int
			switch conn {
			case ConnectorTop:
				dir = [2]int{0, -1}
			case ConnectorRight:
				dir = [2]int{1, 0}
			case ConnectorBottom:
				dir = [2]int{0, 1}
			case ConnectorLeft:
				dir = [2]int{-1, 0}
			}

			// Calculate the adjacent position
			nx, ny := newX+dir[0], newY+dir[1]

			// Skip if there's already a cell there
			if worldMap.HasCell(nx, ny) {
				continue
			}

			// If we already created a sub-branch in this direction, skip
			if createdSubBranch {
				// Check if we created a sub-branch in this direction
				subBranchExists := false
				for _, subDir := range getDirections() {
					subNx, subNy := newX+subDir[0], newY+subDir[1]
					if subNx == nx && subNy == ny && worldMap.HasCell(subNx, subNy) {
						subBranchExists = true
						break
					}
				}
				if subBranchExists {
					continue
				}
			}

			// Create a dead-end branch in this direction
			// This is mandatory - we must connect all connectors
			err := g.generateDeadEnd(worldMap, newCell, dir, config, rng)
			if err != nil {
				// If we can't create a dead-end, log the error but continue
				// This should never happen with our changes to generateDeadEnd
				fmt.Printf("Warning: failed to generate dead-end from (%d,%d): %v\n", newX, newY, err)
			}
		}
	}

	return nil
}

// generateDeadEnd creates a dead-end snippet in the specified direction
func (g *WorldGenerator) generateDeadEnd(worldMap *WorldMap, fromCell *WorldCell, direction [2]int, config *WorldGenConfig, rng *rand.Rand) error {
	// Calculate the new position
	newX, newY := fromCell.X+direction[0], fromCell.Y+direction[1]

	// Skip if there's already a cell there
	if worldMap.HasCell(newX, newY) {
		return nil
	}

	// Determine the connector direction from the new cell back to the current cell
	var toConnector SnippetConnector
	if direction[0] == 0 && direction[1] == -1 {
		toConnector = ConnectorBottom
	} else if direction[0] == 1 && direction[1] == 0 {
		toConnector = ConnectorLeft
	} else if direction[0] == 0 && direction[1] == 1 {
		toConnector = ConnectorTop
	} else {
		toConnector = ConnectorRight
	}

	// Find a snippet that has the required connector
	snippets := g.Registry.GetSnippetsByConnector(toConnector)
	if len(snippets) == 0 {
		return fmt.Errorf("no snippets found with connector %d", toConnector)
	}

	// Try to find a dead-end snippet (only one connector)
	var deadEndSnippets []*WorldSnippet
	for _, s := range snippets {
		if len(s.Connectors) == 1 {
			deadEndSnippets = append(deadEndSnippets, s)
		}
	}

	var snippet *WorldSnippet
	var isDeadEnd bool

	// Prefer a dead-end snippet, but use any snippet if necessary
	if len(deadEndSnippets) > 0 {
		snippet = SelectWeightedSnippetWithTypeWeights(deadEndSnippets, config.SnippetTypeWeights, rng)
		isDeadEnd = true
	} else {
		// If no dead-end snippets are available, use any snippet with the required connector
		// We'll ensure all other connectors are connected to something
		snippet = SelectWeightedSnippetWithTypeWeights(snippets, config.SnippetTypeWeights, rng)
		isDeadEnd = false
	}

	// Determine the rotation needed
	rotation := 0
	for _, conn := range snippet.Connectors {
		if conn == toConnector {
			break
		}
		rotation += 90
		if rotation >= 360 {
			rotation = 0
		}
	}

	// Create the new cell
	newCell := &WorldCell{
		X:           newX,
		Y:           newY,
		Snippet:     snippet,
		Rotation:    rotation,
		IsMainPath:  false,
		BranchDepth: fromCell.BranchDepth + 1,
	}

	worldMap.AddCell(newCell)

	// If this isn't a dead-end snippet, we need to add dead-ends to all its other connectors
	if !isDeadEnd {
		// Get the rotated connectors for this cell
		rotatedConnectors := newCell.GetRotatedConnectors()

		// For each connector
		for _, conn := range rotatedConnectors {
			// Skip the connector that connects back to the previous cell
			oppositeConnector := (toConnector + 180) % 360
			if conn == oppositeConnector {
				continue
			}

			// Calculate the direction for this connector
			var dir [2]int
			switch conn {
			case ConnectorTop:
				dir = [2]int{0, -1}
			case ConnectorRight:
				dir = [2]int{1, 0}
			case ConnectorBottom:
				dir = [2]int{0, 1}
			case ConnectorLeft:
				dir = [2]int{-1, 0}
			}

			// Calculate the adjacent position
			nx, ny := newX+dir[0], newY+dir[1]

			// Skip if there's already a cell there
			if worldMap.HasCell(nx, ny) {
				continue
			}

			// Create a dead-end in this direction
			// This is mandatory - we must connect all connectors
			// We use a recursive call with a depth limit to prevent infinite recursion
			if newCell.BranchDepth < config.BranchMaxDepth {
				// Try to create a dead-end with a single-connector snippet
				deadEndSnippets := make([]*WorldSnippet, 0)
				for _, s := range g.Registry.GetSnippetsByConnector((conn+180)%360) {
					if len(s.Connectors) == 1 {
						deadEndSnippets = append(deadEndSnippets, s)
					}
				}

				// If we have dead-end snippets, use one of them
				if len(deadEndSnippets) > 0 {
					// Create a temporary cell with a dead-end snippet
					deadEndSnippet := SelectWeightedSnippetWithTypeWeights(deadEndSnippets, config.SnippetTypeWeights, rng)
					rotation := 0
					for _, c := range deadEndSnippet.Connectors {
						if c == (conn+180)%360 {
							break
						}
						rotation += 90
						if rotation >= 360 {
							rotation = 0
						}
					}

					deadEndCell := &WorldCell{
						X:           nx,
						Y:           ny,
						Snippet:     deadEndSnippet,
						Rotation:    rotation,
						IsMainPath:  false,
						BranchDepth: newCell.BranchDepth + 1,
					}

					worldMap.AddCell(deadEndCell)
				} else {
					// If no dead-end snippets are available, use generateDeadEnd
					err := g.generateDeadEnd(worldMap, newCell, dir, config, rng)
					if err != nil {
						// If we can't create a dead-end, log the error but continue
						fmt.Printf("Warning: failed to generate dead-end from (%d,%d): %v\n", newX, newY, err)
					}
				}
			}
		}
	}

	return nil
}

// getDirections returns the four possible directions
func getDirections() [][2]int {
	return [][2]int{
		{0, -1}, // Top
		{1, 0},  // Right
		{0, 1},  // Bottom
		{-1, 0}, // Left
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// postProcessWorldMap performs post-processing on the world map:
// 1. Finds the boundaries of the generated world
// 2. Creates a 2-snippet wide wall of empty snippets around the world
// 3. Fills any empty spaces with the empty snippet
// 4. Ensures no paths/connectors point to the world borders
// 5. Ensures sides without connectors have empty snippets adjacent to them
// 6. Ensures each empty snippet has at least 2 other empty snippets adjacent to it
func (g *WorldGenerator) postProcessWorldMap(worldMap *WorldMap, rng *rand.Rand) error {
	// Find the boundaries of the generated world
	minX, minY, maxX, maxY := findWorldBoundaries(worldMap)

	// Expand boundaries by 2 for the wall
	minX -= 2
	minY -= 2
	maxX += 2
	maxY += 2

	// Get the empty snippet
	emptySnippet := g.Registry.GetSnippet("Worldgen_x.png")
	if emptySnippet == nil {
		return fmt.Errorf("empty snippet 'Worldgen_x.png' not found")
	}

	// Create a 2-snippet wide wall around the world and fill empty spaces
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			// Skip if there's already a cell at this position
			if worldMap.HasCell(x, y) {
				continue
			}

			// Create a new cell with the empty snippet
			newCell := &WorldCell{
				X:           x,
				Y:           y,
				Snippet:     emptySnippet,
				Rotation:    0,
				IsMainPath:  false,
				BranchDepth: 0,
			}

			worldMap.AddCell(newCell)
		}
	}

	// Ensure no paths/connectors point to the world borders
	// This is done by checking all cells near the borders
	for y := minY + 2; y <= maxY - 2; y++ {
		for x := minX + 2; x <= maxX - 2; x++ {
			cell := worldMap.GetCell(x, y)
			if cell == nil || len(cell.Snippet.Connectors) == 0 {
				continue
			}

			// Check if this cell is adjacent to the border
			isBorderCell := x == minX+2 || x == maxX-2 || y == minY+2 || y == maxY-2

			if isBorderCell {
				// Get the rotated connectors for this cell
				rotatedConnectors := cell.GetRotatedConnectors()

				// Check if any connector points to the border
				hasInvalidConnector := false
				for _, conn := range rotatedConnectors {
					// Determine the direction this connector points to
					var dx, dy int
					switch conn {
					case ConnectorTop:
						dx, dy = 0, -1
					case ConnectorRight:
						dx, dy = 1, 0
					case ConnectorBottom:
						dx, dy = 0, 1
					case ConnectorLeft:
						dx, dy = -1, 0
					}

					// Check if this connector points to the border
					nx, ny := x+dx, y+dy
					if nx <= minX+1 || nx >= maxX-1 || ny <= minY+1 || ny >= maxY-1 {
						hasInvalidConnector = true
						break
					}
				}

				// If this cell has an invalid connector, replace it with the empty snippet
				if hasInvalidConnector {
					cell.Snippet = emptySnippet
					cell.Rotation = 0
				}
			}
		}
	}

	// First pass: Ensure sides without connectors have empty snippets adjacent to them
	// This is a strict requirement - all non-connector sides must have empty snippets
	for y := minY + 2; y <= maxY - 2; y++ {
		for x := minX + 2; x <= maxX - 2; x++ {
			cell := worldMap.GetCell(x, y)
			if cell == nil || len(cell.Snippet.Connectors) == 0 {
				continue
			}

			// Get the rotated connectors for this cell
			rotatedConnectors := cell.GetRotatedConnectors()

			// Check all four directions
			directions := [][2]int{
				{0, -1}, // Top
				{1, 0},  // Right
				{0, 1},  // Bottom
				{-1, 0}, // Left
			}
			connectorDirections := []SnippetConnector{
				ConnectorTop,
				ConnectorRight,
				ConnectorBottom,
				ConnectorLeft,
			}

			// For each direction, check if there's a connector pointing that way
			for i, dir := range directions {
				// Skip if there's already a connector in this direction
				hasConnector := false
				for _, conn := range rotatedConnectors {
					if conn == connectorDirections[i] {
						hasConnector = true
						break
					}
				}

				// If there's no connector in this direction, ensure there's an empty snippet adjacent
				if !hasConnector {
					nx, ny := x+dir[0], y+dir[1]

					// Skip if outside the valid world area
					if nx <= minX+1 || nx >= maxX-1 || ny <= minY+1 || ny >= maxY-1 {
						continue
					}

					// If there's no cell there, add an empty snippet
					if !worldMap.HasCell(nx, ny) {
						newCell := &WorldCell{
							X:           nx,
							Y:           ny,
							Snippet:     emptySnippet,
							Rotation:    0,
							IsMainPath:  false,
							BranchDepth: 0,
						}
						worldMap.AddCell(newCell)
					} else {
						// If there's a cell there but it's not an empty snippet, replace it with an empty snippet
						// UNLESS it has a connector pointing back to this cell
						adjacentCell := worldMap.GetCell(nx, ny)

						// Always replace with an empty snippet if the adjacent cell has connectors
						// but none pointing back to this cell
						if len(adjacentCell.Snippet.Connectors) > 0 {
							// Check if the adjacent cell has a connector pointing back to this cell
							adjacentRotatedConnectors := adjacentCell.GetRotatedConnectors()
							oppositeDirection := (i + 2) % 4 // Opposite direction index
							oppositeConnector := connectorDirections[oppositeDirection]

							hasOppositeConnector := false
							for _, conn := range adjacentRotatedConnectors {
								if conn == oppositeConnector {
									hasOppositeConnector = true
									break
								}
							}

							// If the adjacent cell doesn't have a connector pointing back, replace it with an empty snippet
							if !hasOppositeConnector {
								adjacentCell.Snippet = emptySnippet
								adjacentCell.Rotation = 0
							}
						}
					}
				}
			}
		}
	}

	// Second pass: Ensure each empty snippet has at least 2 other empty snippets adjacent to it
	// This is done by checking all empty snippets and adding more empty snippets if needed
	for y := minY + 2; y <= maxY - 2; y++ {
		for x := minX + 2; x <= maxX - 2; x++ {
			cell := worldMap.GetCell(x, y)
			if cell == nil || len(cell.Snippet.Connectors) > 0 {
				continue
			}

			// This is an empty snippet, check how many adjacent empty snippets it has
			directions := [][2]int{
				{0, -1}, // Top
				{1, 0},  // Right
				{0, 1},  // Bottom
				{-1, 0}, // Left
			}

			// Count adjacent empty snippets
			adjacentEmptyCount := 0
			emptyDirections := make([]int, 0)
			nonEmptyDirections := make([]int, 0)

			for i, dir := range directions {
				nx, ny := x+dir[0], y+dir[1]

				// Skip if outside the valid world area
				if nx <= minX+1 || nx >= maxX-1 || ny <= minY+1 || ny >= maxY-1 {
					continue
				}

				adjacentCell := worldMap.GetCell(nx, ny)
				if adjacentCell != nil && len(adjacentCell.Snippet.Connectors) == 0 {
					adjacentEmptyCount++
					emptyDirections = append(emptyDirections, i)
				} else {
					nonEmptyDirections = append(nonEmptyDirections, i)
				}
			}

			// If there are fewer than 2 adjacent empty snippets, add more
			if adjacentEmptyCount < 2 {
				// We need to add (2 - adjacentEmptyCount) empty snippets
				neededCount := 2 - adjacentEmptyCount

				// Try to replace non-empty adjacent cells with empty snippets
				for i := 0; i < len(nonEmptyDirections) && neededCount > 0; i++ {
					dirIndex := nonEmptyDirections[i]
					dir := directions[dirIndex]
					nx, ny := x+dir[0], y+dir[1]

					adjacentCell := worldMap.GetCell(nx, ny)
					if adjacentCell != nil {
						// Check if this cell has a connector pointing to our empty snippet
						hasConnectorToEmpty := false
						if len(adjacentCell.Snippet.Connectors) > 0 {
							rotatedConnectors := adjacentCell.GetRotatedConnectors()
							oppositeDirection := (dirIndex + 2) % 4 // Opposite direction index
							oppositeConnector := []SnippetConnector{
								ConnectorTop,
								ConnectorRight,
								ConnectorBottom,
								ConnectorLeft,
							}[oppositeDirection]

							for _, conn := range rotatedConnectors {
								if conn == oppositeConnector {
									hasConnectorToEmpty = true
									break
								}
							}
						}

						// If it doesn't have a connector pointing to our empty snippet, replace it
						if !hasConnectorToEmpty {
							adjacentCell.Snippet = emptySnippet
							adjacentCell.Rotation = 0
							neededCount--
						}
					}
				}

				// If we still need more empty snippets, create new ones in empty spaces
				if neededCount > 0 {
					// Look for empty spaces around this cell
					for _, dir := range directions {
						nx, ny := x+dir[0], y+dir[1]

						// Skip if outside the valid world area or if there's already a cell
						if nx <= minX+1 || nx >= maxX-1 || ny <= minY+1 || ny >= maxY-1 || worldMap.HasCell(nx, ny) {
							continue
						}

						// Create a new empty snippet
						newCell := &WorldCell{
							X:           nx,
							Y:           ny,
							Snippet:     emptySnippet,
							Rotation:    0,
							IsMainPath:  false,
							BranchDepth: 0,
						}
						worldMap.AddCell(newCell)
						neededCount--

						if neededCount == 0 {
							break
						}
					}
				}
			}
		}
	}

 // Final check to ensure there are no gaps in the world
	// This is a safety measure to guarantee that every position has a snippet
	minX, minY, maxX, maxY = findWorldBoundaries(worldMap)

	// Expand boundaries by 1 to ensure we check the entire area
	minX -= 1
	minY -= 1
	maxX += 1
	maxY += 1

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if !worldMap.HasCell(x, y) {
				// If we find a gap, fill it with the empty snippet
				newCell := &WorldCell{
					X:           x,
					Y:           y,
					Snippet:     emptySnippet,
					Rotation:    0,
					IsMainPath:  false,
					BranchDepth: 0,
				}
				worldMap.AddCell(newCell)
			}
		}
	}

	// Additional check to ensure all connectors are connected to another connector
	// This is a strict requirement - each connector must be adjacent to another connector
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			cell := worldMap.GetCell(x, y)
			if cell == nil || len(cell.Snippet.Connectors) == 0 {
				continue
			}

			// Get the rotated connectors for this cell
			rotatedConnectors := cell.GetRotatedConnectors()

			// For each connector, check if it's connected to another connector
			for _, conn := range rotatedConnectors {
				// Calculate the direction for this connector
				var dx, dy int
				switch conn {
				case ConnectorTop:
					dx, dy = 0, -1
				case ConnectorRight:
					dx, dy = 1, 0
				case ConnectorBottom:
					dx, dy = 0, 1
				case ConnectorLeft:
					dx, dy = -1, 0
				}

				// Calculate the adjacent position
				nx, ny := x+dx, y+dy

				// Get the adjacent cell
				adjacentCell := worldMap.GetCell(nx, ny)
				if adjacentCell == nil {
					// If there's no adjacent cell, create an empty snippet
					adjacentCell = &WorldCell{
						X:           nx,
						Y:           ny,
						Snippet:     emptySnippet,
						Rotation:    0,
						IsMainPath:  false,
						BranchDepth: 0,
					}
					worldMap.AddCell(adjacentCell)
					continue
				}

				// If the adjacent cell has no connectors, it's an empty snippet
				if len(adjacentCell.Snippet.Connectors) == 0 {
					// Find a snippet with the required connector
					var oppositeConnector SnippetConnector
					switch conn {
					case ConnectorTop:
						oppositeConnector = ConnectorBottom
					case ConnectorRight:
						oppositeConnector = ConnectorLeft
					case ConnectorBottom:
						oppositeConnector = ConnectorTop
					case ConnectorLeft:
						oppositeConnector = ConnectorRight
					}

					snippets := g.Registry.GetSnippetsByConnector(oppositeConnector)
					if len(snippets) == 0 {
						continue
					}

					// Choose a snippet with only one connector if possible
					var selectedSnippet *WorldSnippet
					for _, s := range snippets {
						if len(s.Connectors) == 1 {
							selectedSnippet = s
							break
						}
					}

					// If no single-connector snippet found, use any snippet
					if selectedSnippet == nil {
						selectedSnippet = snippets[0]
					}

					// Determine the rotation needed
					rotation := 0
					for _, c := range selectedSnippet.Connectors {
						if c == oppositeConnector {
							break
						}
						rotation += 90
						if rotation >= 360 {
							rotation = 0
						}
					}

					// Replace the empty snippet with the selected snippet
					adjacentCell.Snippet = selectedSnippet
					adjacentCell.Rotation = rotation
				} else {
					// Check if the adjacent cell has a connector pointing back
					adjacentRotatedConnectors := adjacentCell.GetRotatedConnectors()
					var oppositeConnector SnippetConnector
					switch conn {
					case ConnectorTop:
						oppositeConnector = ConnectorBottom
					case ConnectorRight:
						oppositeConnector = ConnectorLeft
					case ConnectorBottom:
						oppositeConnector = ConnectorTop
					case ConnectorLeft:
						oppositeConnector = ConnectorRight
					}

					hasOppositeConnector := false
					for _, c := range adjacentRotatedConnectors {
						if c == oppositeConnector {
							hasOppositeConnector = true
							break
						}
					}

					// If the adjacent cell doesn't have a connector pointing back,
					// replace it with a snippet that does
					if !hasOppositeConnector {
						snippets := g.Registry.GetSnippetsByConnector(oppositeConnector)
						if len(snippets) == 0 {
							continue
						}

						// Choose a snippet with only one connector if possible
						var selectedSnippet *WorldSnippet
						for _, s := range snippets {
							if len(s.Connectors) == 1 {
								selectedSnippet = s
								break
							}
						}

						// If no single-connector snippet found, use any snippet
						if selectedSnippet == nil {
							selectedSnippet = snippets[0]
						}

						// Determine the rotation needed
						rotation := 0
						for _, c := range selectedSnippet.Connectors {
							if c == oppositeConnector {
								break
							}
							rotation += 90
							if rotation >= 360 {
								rotation = 0
							}
						}

						// Replace the adjacent cell with the selected snippet
						adjacentCell.Snippet = selectedSnippet
						adjacentCell.Rotation = rotation
					}
				}
			}
		}
	}

	// One more pass to ensure all snippets are exactly adjacent to each other (no gaps or overlaps)
	// This is done by checking all cells and ensuring they have neighbors in all four directions
	for y := minY + 1; y <= maxY - 1; y++ {
		for x := minX + 1; x <= maxX - 1; x++ {
			cell := worldMap.GetCell(x, y)
			if cell == nil {
				continue
			}

			// Check all four directions
			directions := [][2]int{
				{0, -1}, // Top
				{1, 0},  // Right
				{0, 1},  // Bottom
				{-1, 0}, // Left
			}

			for _, dir := range directions {
				nx, ny := x+dir[0], y+dir[1]

				// If there's no cell in this direction, add an empty snippet
				if !worldMap.HasCell(nx, ny) {
					newCell := &WorldCell{
						X:           nx,
						Y:           ny,
						Snippet:     emptySnippet,
						Rotation:    0,
						IsMainPath:  false,
						BranchDepth: 0,
					}
					worldMap.AddCell(newCell)
				}
			}
		}
	}

	return nil
}

// findWorldBoundaries finds the minimum and maximum x and y coordinates of the world
func findWorldBoundaries(worldMap *WorldMap) (minX, minY, maxX, maxY int) {
	// Initialize with extreme values
	minX, minY = 1000000, 1000000
	maxX, maxY = -1000000, -1000000

	// Find the actual boundaries
	for _, cell := range worldMap.Cells {
		if cell.X < minX {
			minX = cell.X
		}
		if cell.Y < minY {
			minY = cell.Y
		}
		if cell.X > maxX {
			maxX = cell.X
		}
		if cell.Y > maxY {
			maxY = cell.Y
		}
	}

	return minX, minY, maxX, maxY
}
