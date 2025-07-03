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
// This satisfies requirement 3: The main path runs in a circular shape and is thus endless - follows the main path, flies in a huge circle
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
	// We need to ensure the last cell is adjacent to the start cell (0,0)
	// Find a position adjacent to (0,0) that isn't already in the path
	lastPosition := [2]int{-1, -1}

	// Check all four directions from the start position
	for _, dir := range directions {
		nx, ny := dir[0], dir[1]

		// Check if this position is already in the planned path
		positionOccupied := false
		for _, pos := range plannedPath[1:] { // Skip the start position
			if pos[0] == nx && pos[1] == ny {
				positionOccupied = true
				break
			}
		}

		// If this position isn't occupied, we can use it for the last cell
		if !positionOccupied {
			lastPosition = [2]int{nx, ny}
			break
		}
	}

	// If we couldn't find an empty position adjacent to the start,
	// we'll need to modify the path to make room
	if lastPosition[0] == -1 && lastPosition[1] == -1 {
		// Choose a direction (right for simplicity)
		lastPosition = [2]int{1, 0}

		// Remove any cell at this position from the planned path
		for i := 1; i < len(plannedPath); i++ {
			if plannedPath[i][0] == lastPosition[0] && plannedPath[i][1] == lastPosition[1] {
				// Remove this position from the path
				plannedPath = append(plannedPath[:i], plannedPath[i+1:]...)
				break
			}
		}
	}

	// Now we need to connect the current position to the last position
	// First, add a path from the current position to a position adjacent to the last position
	for currentX != lastPosition[0] || currentY != lastPosition[1] {
		// Determine the direction to move
		dx, dy := 0, 0
		if currentX < lastPosition[0] {
			dx = 1
		} else if currentX > lastPosition[0] {
			dx = -1
		} else if currentY < lastPosition[1] {
			dy = 1
		} else if currentY > lastPosition[1] {
			dy = -1
		}

		// Move one step in that direction
		currentX += dx
		currentY += dy

		// Add this position to the planned path
		plannedPath = append(plannedPath, [2]int{currentX, currentY})
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

		// For the last cell, we need a snippet with exactly two connectors:
		// one connecting to the previous cell and one connecting to the start cell
		if i == len(plannedPath)-1 {
			// Determine the connector direction from the last cell to the start cell
			var toStartConnector SnippetConnector
			if nextX == 0 && nextY-1 == 0 {
				toStartConnector = ConnectorBottom
			} else if nextX+1 == 0 && nextY == 0 {
				toStartConnector = ConnectorLeft
			} else if nextX == 0 && nextY+1 == 0 {
				toStartConnector = ConnectorTop
			} else {
				toStartConnector = ConnectorRight
			}

			// Filter snippets to only include those with both required connectors
			// Consider rotation when checking for connectors
			var filteredSnippets []*WorldSnippet

			// Create a map to store the correct rotation for each snippet
			snippetRotations := make(map[*WorldSnippet]int)

			for _, s := range snippets {
				// Check if this snippet can have both connectors after rotation
				canHaveBothConnectors := false

				// Try all possible rotations (0, 90, 180, 270)
				for rotation := 0; rotation < 360; rotation += 90 {
					// Create a temporary cell with this snippet and rotation
					tempCell := &WorldCell{
						Snippet:  s,
						Rotation: rotation,
					}

					// Get the rotated connectors
					rotatedConnectors := tempCell.GetRotatedConnectors()

					// Check if the rotated connectors include both required connectors
					hasToConnector := false
					hasToStartConnector := false

					for _, conn := range rotatedConnectors {
						if conn == toConnector {
							hasToConnector = true
						}
						if conn == toStartConnector {
							hasToStartConnector = true
						}
					}

					if hasToConnector && hasToStartConnector {
						canHaveBothConnectors = true
						snippetRotations[s] = rotation // Store the correct rotation
						break
					}
				}

				if canHaveBothConnectors {
					filteredSnippets = append(filteredSnippets, s)
				}
			}

			if len(filteredSnippets) > 0 {
				snippets = filteredSnippets

				// Choose a snippet based on weight and type
				snippet := SelectWeightedSnippetWithTypeWeights(snippets, config.SnippetTypeWeights, rng)

				// Use the stored rotation for this snippet
				rotation := snippetRotations[snippet]

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

				// Skip the rest of the loop for the last cell
				continue
			} else {
				return fmt.Errorf("no snippets found with connectors %d and %d for the last cell (even after considering rotation)", toConnector, toStartConnector)
			}
		}

		// Choose a snippet based on weight and type
		snippet := SelectWeightedSnippetWithTypeWeights(snippets, config.SnippetTypeWeights, rng)

		// Determine the rotation needed
		rotation := 0

		// Find a rotation that satisfies the toConnector
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

	// Verify that the main path forms a complete circle
	// The last cell should be adjacent to the start cell and have a connector pointing to it
	lastCell := worldMap.GetCell(currentX, currentY)
	if lastCell == nil {
		return fmt.Errorf("last cell of main path not found")
	}

	// Check if the last cell is adjacent to the start cell
	dx, dy := 0-currentX, 0-currentY
	if abs(dx)+abs(dy) != 1 {
		return fmt.Errorf("last cell of main path is not adjacent to the start cell")
	}

	// Check if the last cell has a connector pointing to the start cell
	var toStartConnector SnippetConnector
	if dx == 0 && dy == -1 {
		toStartConnector = ConnectorBottom
	} else if dx == 1 && dy == 0 {
		toStartConnector = ConnectorLeft
	} else if dx == 0 && dy == 1 {
		toStartConnector = ConnectorTop
	} else {
		toStartConnector = ConnectorRight
	}

	rotatedConnectors := lastCell.GetRotatedConnectors()
	hasConnectorToStart := false
	for _, conn := range rotatedConnectors {
		if conn == toStartConnector {
			hasConnectorToStart = true
			break
		}
	}

	if !hasConnectorToStart {
		return fmt.Errorf("last cell of main path does not have a connector pointing to the start cell")
	}

	// Also check if the start cell has a connector pointing to the last cell
	var toLastConnector SnippetConnector
	if dx == 0 && dy == -1 {
		toLastConnector = ConnectorTop
	} else if dx == 1 && dy == 0 {
		toLastConnector = ConnectorRight
	} else if dx == 0 && dy == 1 {
		toLastConnector = ConnectorBottom
	} else {
		toLastConnector = ConnectorLeft
	}

	startRotatedConnectors := startCell.GetRotatedConnectors()
	hasConnectorToLast := false
	for _, conn := range startRotatedConnectors {
		if conn == toLastConnector {
			hasConnectorToLast = true
			break
		}
	}

	if !hasConnectorToLast {
		return fmt.Errorf("start cell of main path does not have a connector pointing to the last cell")
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
// This satisfies requirement 6: Branches from the main path either lead back to the main path or end with a snippet with only 1 connector as a dead-end
func (g *WorldGenerator) generateBranch(worldMap *WorldMap, fromCell *WorldCell, direction [2]int, depth int, config *WorldGenConfig, rng *rand.Rand) error {
	// Verify that the branch starts from the main path if this is the first level of the branch
	// This satisfies requirement 5: Every branch must necessarily start from the main path
	if depth == 1 && !fromCell.IsMainPath {
		return fmt.Errorf("branch must start from the main path")
	}

	// If we've reached the maximum branch depth, we must place a dead-end
	if depth > config.BranchMaxDepth {
		return g.generateDeadEnd(worldMap, fromCell, direction, config, rng)
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
		// If no single-connector snippets found, we'll use a snippet with multiple connectors
		// but ensure all its connectors are connected to other snippets
		if len(filteredSnippets) == 0 {
			return g.generateDeadEnd(worldMap, fromCell, direction, config, rng)
		}
	} else {
		// For non-dead-ends, use snippets with at least two connectors
		for _, s := range snippets {
			if len(s.Connectors) >= 2 {
				filteredSnippets = append(filteredSnippets, s)
			}
		}

		// If no suitable snippets found, use any snippet with the required connector
		if len(filteredSnippets) == 0 {
			filteredSnippets = snippets
		}
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
				for _, s := range g.Registry.GetSnippetsByConnector((conn + 180) % 360) {
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
// 7. Ensures all paths are connected to the main path
// 8. Ensures all connectors are connected to another connector
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

	// Ensure no paths/connectors point to the world borders or to cells without matching connectors
	// This is done by checking all cells in the world, not just those adjacent to borders
	// This satisfies requirement 2: Each connector must necessarily be adjacent to (connected with) another connector
	for y := minY + 2; y <= maxY-2; y++ {
		for x := minX + 2; x <= maxX-2; x++ {
			cell := worldMap.GetCell(x, y)
			if cell == nil || len(cell.Snippet.Connectors) == 0 {
				continue
			}

			// Get the rotated connectors for this cell
			rotatedConnectors := cell.GetRotatedConnectors()

			// Check if any connector points to the border or to an empty space
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

				// Calculate the position this connector points to
				nx, ny := x+dx, y+dy

				// Check if this connector points to the border
				if nx <= minX+1 || nx >= maxX-1 || ny <= minY+1 || ny >= maxY-1 {
					hasInvalidConnector = true
					break
				}

				// Check if this connector points to a cell with a connector pointing back
				adjacentCell := worldMap.GetCell(nx, ny)
				if adjacentCell == nil || len(adjacentCell.Snippet.Connectors) == 0 {
					// Points to an empty space or empty snippet
					hasInvalidConnector = true
					break
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
					for _, adjConn := range adjacentRotatedConnectors {
						if adjConn == oppositeConnector {
							hasOppositeConnector = true
							break
						}
					}

					if !hasOppositeConnector {
						hasInvalidConnector = true
						break
					}
				}
			}

			// If this cell has an invalid connector, replace it with the empty snippet
			// But never replace main path cells
			if hasInvalidConnector && !cell.IsMainPath {
				cell.Snippet = emptySnippet
				cell.Rotation = 0
			}
		}
	}

	// First pass: Ensure sides without connectors have empty snippets adjacent to them
	// This satisfies requirement 7: On each side of a snippet where there is no connector, an empty snippet must be adjacent (Worldgen_x.png)
	for y := minY + 2; y <= maxY-2; y++ {
		for x := minX + 2; x <= maxX-2; x++ {
			cell := worldMap.GetCell(x, y)
			// Skip empty cells or cells with no snippet
			if cell == nil {
				continue
			}

			// Process all cells, not just those with connectors
			// For empty snippets, we'll handle them differently

			// Define directions and connector types
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

			// If this is a cell with connectors
			if len(cell.Snippet.Connectors) > 0 {
				// Get the rotated connectors for this cell
				rotatedConnectors := cell.GetRotatedConnectors()

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
									// First, check if replacing the adjacent cell would break other connections
									// Count how many connectors the adjacent cell has
									adjacentConnectorCount := len(adjacentCell.Snippet.Connectors)

									// If the adjacent cell has multiple connectors, we need to be careful
									if adjacentConnectorCount > 1 {
										// Check if any of its connectors are connected to other cells
										isConnectedToOthers := false

										// For each connector in the adjacent cell
										for _, adjConn := range adjacentRotatedConnectors {
											// Skip the connector that should point to our cell
											if adjConn == oppositeConnector {
												continue
											}

											// Calculate the direction for this connector
											var adjDx, adjDy int
											switch adjConn {
											case ConnectorTop:
												adjDx, adjDy = 0, -1
											case ConnectorRight:
												adjDx, adjDy = 1, 0
											case ConnectorBottom:
												adjDx, adjDy = 0, 1
											case ConnectorLeft:
												adjDx, adjDy = -1, 0
											}

											// Calculate the position this connector points to
											adjNx, adjNy := nx+adjDx, ny+adjDy

											// Check if there's a cell there with a connector pointing back
											otherCell := worldMap.GetCell(adjNx, adjNy)
											if otherCell != nil && len(otherCell.Snippet.Connectors) > 0 {
												otherRotatedConnectors := otherCell.GetRotatedConnectors()

												// Determine the opposite connector
												var otherOppositeConnector SnippetConnector
												switch adjConn {
												case ConnectorTop:
													otherOppositeConnector = ConnectorBottom
												case ConnectorRight:
													otherOppositeConnector = ConnectorLeft
												case ConnectorBottom:
													otherOppositeConnector = ConnectorTop
												case ConnectorLeft:
													otherOppositeConnector = ConnectorRight
												}

												// Check if the other cell has a connector pointing back
												for _, otherConn := range otherRotatedConnectors {
													if otherConn == otherOppositeConnector {
														isConnectedToOthers = true
														break
													}
												}

												if isConnectedToOthers {
													break
												}
											}
										}

										// If the adjacent cell is connected to other cells, don't replace it
										if isConnectedToOthers {
											continue
										}
									}

									// If we get here, it's safe to replace the adjacent cell with an empty snippet
									// But never replace main path cells
									if !adjacentCell.IsMainPath {
										adjacentCell.Snippet = emptySnippet
										adjacentCell.Rotation = 0
									}
								}
							}
						}
					}
				}
			} else {
				// This is an empty snippet
				// For empty snippets, we don't need to check for connectors
				// but we still need to ensure all sides have valid neighbors

				// For each direction from this empty snippet
				for _, dir := range directions {
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
					}
				}
			}
		}
	}

	// Second pass: Ensure each empty snippet has at least 2 other empty snippets adjacent to it
	// This satisfies requirement 8: At least 2 other empty snippets must be adjacent to each empty snippet

	// First, create a map to track which empty snippets need additional adjacent empty snippets
	emptySnippetsNeedingAdjacent := make(map[string]bool)

	// Identify all empty snippets that need additional adjacent empty snippets
	for y := minY + 2; y <= maxY-2; y++ {
		for x := minX + 2; x <= maxX-2; x++ {
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

			for _, dir := range directions {
				nx, ny := x+dir[0], y+dir[1]

				// Skip if outside the valid world area
				if nx <= minX+1 || nx >= maxX-1 || ny <= minY+1 || ny >= maxY-1 {
					continue
				}

				adjacentCell := worldMap.GetCell(nx, ny)
				if adjacentCell != nil && len(adjacentCell.Snippet.Connectors) == 0 {
					adjacentEmptyCount++
				}
			}

			// If this empty snippet has fewer than 2 adjacent empty snippets, mark it
			if adjacentEmptyCount < 2 {
				key := fmt.Sprintf("%d,%d", x, y)
				emptySnippetsNeedingAdjacent[key] = true
			}
		}
	}

	// Now process each empty snippet that needs additional adjacent empty snippets
	// We do this in a separate pass to avoid conflicts between different empty snippets
	for y := minY + 2; y <= maxY-2; y++ {
		for x := minX + 2; x <= maxX-2; x++ {
			key := fmt.Sprintf("%d,%d", x, y)
			if !emptySnippetsNeedingAdjacent[key] {
				continue
			}

			cell := worldMap.GetCell(x, y)
			if cell == nil || len(cell.Snippet.Connectors) > 0 {
				continue
			}

			// This is an empty snippet that needs more adjacent empty snippets
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

				// Strategy 1: First, try to create new empty snippets in empty spaces
				// This is the safest approach as it doesn't modify existing cells
				for _, dir := range directions {
					if neededCount <= 0 {
						break
					}

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
				}

				// Strategy 2: If we still need more, try diagonal directions
				if neededCount > 0 {
					diagonalDirections := [][2]int{
						{-1, -1}, // Top-left
						{1, -1},  // Top-right
						{1, 1},   // Bottom-right
						{-1, 1},  // Bottom-left
					}

					for _, dir := range diagonalDirections {
						if neededCount <= 0 {
							break
						}

						nx, ny := x+dir[0], y+dir[1]

						// Skip if outside the valid world area
						if nx <= minX+1 || nx >= maxX-1 || ny <= minY+1 || ny >= maxY-1 {
							continue
						}

						// Create a new empty snippet in this direction if there's no cell
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
							neededCount--
						} else {
							// If there's already a cell here, check if it's an empty snippet
							adjacentCell := worldMap.GetCell(nx, ny)
							if adjacentCell != nil && len(adjacentCell.Snippet.Connectors) == 0 {
								// It's already an empty snippet, count it
								neededCount--
							}
						}
					}
				}

				// Strategy 3: If we still need more empty snippets, try to replace non-empty adjacent cells
				// that don't have important connections
				if neededCount > 0 {
					// Try to replace non-empty adjacent cells with empty snippets
					for i := 0; i < len(nonEmptyDirections) && neededCount > 0; i++ {
						dirIndex := nonEmptyDirections[i]
						dir := directions[dirIndex]
						nx, ny := x+dir[0], y+dir[1]

						adjacentCell := worldMap.GetCell(nx, ny)
						if adjacentCell == nil {
							continue
						}

						// Skip if this is a cell with no connectors (already an empty snippet)
						if len(adjacentCell.Snippet.Connectors) == 0 {
							continue
						}

						// Skip if this is a main path cell - we never replace main path cells
						if adjacentCell.IsMainPath {
							continue
						}

						// Check if this cell has a connector pointing to our empty snippet
						rotatedConnectors := adjacentCell.GetRotatedConnectors()
						oppositeDirection := (dirIndex + 2) % 4 // Opposite direction index
						oppositeConnector := []SnippetConnector{
							ConnectorTop,
							ConnectorRight,
							ConnectorBottom,
							ConnectorLeft,
						}[oppositeDirection]

						hasConnectorToEmpty := false
						for _, conn := range rotatedConnectors {
							if conn == oppositeConnector {
								hasConnectorToEmpty = true
								break
							}
						}

						// If it has a connector pointing to our empty snippet, we shouldn't replace it
						// This would violate requirement 7 (sides without connectors must have empty snippets)
						if hasConnectorToEmpty {
							continue
						}

						// Check if any of its connectors are connected to other cells with connectors
						isConnectedToOthers := false

						// For each connector in the adjacent cell
						for _, adjConn := range rotatedConnectors {
							// Calculate the direction for this connector
							var adjDx, adjDy int
							switch adjConn {
							case ConnectorTop:
								adjDx, adjDy = 0, -1
							case ConnectorRight:
								adjDx, adjDy = 1, 0
							case ConnectorBottom:
								adjDx, adjDy = 0, 1
							case ConnectorLeft:
								adjDx, adjDy = -1, 0
							}

							// Calculate the position this connector points to
							adjNx, adjNy := nx+adjDx, ny+adjDy

							// Skip if this connector points to our empty snippet
							if adjNx == x && adjNy == y {
								continue
							}

							// Check if there's a cell there with a connector pointing back
							otherCell := worldMap.GetCell(adjNx, adjNy)
							if otherCell != nil && len(otherCell.Snippet.Connectors) > 0 {
								otherRotatedConnectors := otherCell.GetRotatedConnectors()

								// Determine the opposite connector
								var otherOppositeConnector SnippetConnector
								switch adjConn {
								case ConnectorTop:
									otherOppositeConnector = ConnectorBottom
								case ConnectorRight:
									otherOppositeConnector = ConnectorLeft
								case ConnectorBottom:
									otherOppositeConnector = ConnectorTop
								case ConnectorLeft:
									otherOppositeConnector = ConnectorRight
								}

								// Check if the other cell has a connector pointing back
								for _, otherConn := range otherRotatedConnectors {
									if otherConn == otherOppositeConnector {
										isConnectedToOthers = true
										break
									}
								}

								if isConnectedToOthers {
									break
								}
							}
						}

						// If the adjacent cell is not connected to other cells, we can replace it
						if !isConnectedToOthers {
							// Before replacing, check if this would violate requirement 7
							// For each direction from this cell
							canReplace := true
							for j, checkDir := range directions {
								// Skip the direction pointing to our empty snippet
								if j == oppositeDirection {
									continue
								}

								checkNx, checkNy := nx+checkDir[0], ny+checkDir[1]

								// Skip if outside the valid world area
								if checkNx <= minX+1 || checkNx >= maxX-1 || checkNy <= minY+1 || checkNy >= maxY-1 {
									continue
								}

								checkCell := worldMap.GetCell(checkNx, checkNy)
								if checkCell != nil && len(checkCell.Snippet.Connectors) > 0 {
									// Check if this cell has a connector pointing to our cell
									checkRotatedConnectors := checkCell.GetRotatedConnectors()
									checkOppositeDirection := (j + 2) % 4 // Opposite direction index
									checkOppositeConnector := []SnippetConnector{
										ConnectorTop,
										ConnectorRight,
										ConnectorBottom,
										ConnectorLeft,
									}[checkOppositeDirection]

									for _, conn := range checkRotatedConnectors {
										if conn == checkOppositeConnector {
											// This cell has a connector pointing to our cell
											// If we replace our cell with an empty snippet, this would violate requirement 7
											canReplace = false
											break
										}
									}

									if !canReplace {
										break
									}
								}
							}

							if canReplace {
								adjacentCell.Snippet = emptySnippet
								adjacentCell.Rotation = 0
								neededCount--
							}
						}
					}
				}

				// If we still need more empty snippets, look for cells further away
				if neededCount > 0 {
					// Look for empty spaces two cells away
					for _, dir1 := range directions {
						if neededCount <= 0 {
							break
						}

						nx1, ny1 := x+dir1[0], y+dir1[1]

						// Skip if outside the valid world area
						if nx1 <= minX+1 || nx1 >= maxX-1 || ny1 <= minY+1 || ny1 >= maxY-1 {
							continue
						}

						// Check cells in directions perpendicular to dir1
						for _, dir2 := range directions {
							// Skip if dir2 is parallel to dir1
							if (dir1[0] == dir2[0] && dir1[0] != 0) || (dir1[1] == dir2[1] && dir1[1] != 0) {
								continue
							}

							nx2, ny2 := nx1+dir2[0], ny1+dir2[1]

							// Skip if outside the valid world area
							if nx2 <= minX+1 || nx2 >= maxX-1 || ny2 <= minY+1 || ny2 >= maxY-1 {
								continue
							}

							// If there's no cell there, add an empty snippet
							if !worldMap.HasCell(nx2, ny2) {
								newCell := &WorldCell{
									X:           nx2,
									Y:           ny2,
									Snippet:     emptySnippet,
									Rotation:    0,
									IsMainPath:  false,
									BranchDepth: 0,
								}
								worldMap.AddCell(newCell)
								neededCount--

								if neededCount <= 0 {
									break
								}
							}
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
	// This satisfies requirement 2: Each connector must necessarily be adjacent to (connected with) another connector
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
					// If there's no adjacent cell, create a dead-end snippet
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
						// If no snippets found with the required connector, we need to replace the current cell
						// with a snippet that doesn't have this connector
						// First, find all snippets that don't have this connector
						var alternativeSnippets []*WorldSnippet
						for _, s := range g.Registry.Snippets {
							hasConnector := false
							for _, c := range s.Connectors {
								if c == conn {
									hasConnector = true
									break
								}
							}
							if !hasConnector {
								alternativeSnippets = append(alternativeSnippets, s)
							}
						}

						if len(alternativeSnippets) > 0 {
							// Choose a snippet without the problematic connector
							cell.Snippet = alternativeSnippets[0]
							cell.Rotation = 0
							// Skip this connector since we've replaced the cell
							continue
						} else {
							// If no alternative snippets found, create an empty snippet adjacent
							// This is a fallback, but should never happen with our snippet set
							adjacentCell = &WorldCell{
								X:           nx,
								Y:           ny,
								Snippet:     emptySnippet,
								Rotation:    0,
								IsMainPath:  false,
								BranchDepth: 0,
							}
						}
					} else {
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

						// Create a new cell with the selected snippet
						adjacentCell = &WorldCell{
							X:           nx,
							Y:           ny,
							Snippet:     selectedSnippet,
							Rotation:    rotation,
							IsMainPath:  false,
							BranchDepth: cell.BranchDepth + 1,
						}
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
						// If no snippets found with the required connector, we need to replace the current cell
						// with a snippet that doesn't have this connector
						// First, find all snippets that don't have this connector
						var alternativeSnippets []*WorldSnippet
						for _, s := range g.Registry.Snippets {
							hasConnector := false
							for _, c := range s.Connectors {
								if c == conn {
									hasConnector = true
									break
								}
							}
							if !hasConnector {
								alternativeSnippets = append(alternativeSnippets, s)
							}
						}

						if len(alternativeSnippets) > 0 {
							// Choose a snippet without the problematic connector
							cell.Snippet = alternativeSnippets[0]
							cell.Rotation = 0
							// Skip this connector since we've replaced the cell
							continue
						}
						// If no alternative snippets found, continue with the next connector
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
					adjacentCell.BranchDepth = cell.BranchDepth + 1
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
					// we need to fix this situation
					if !hasOppositeConnector {
						// First, check if replacing the adjacent cell would break other connections
						// Count how many connectors the adjacent cell has
						adjacentConnectorCount := len(adjacentCell.Snippet.Connectors)

						// If the adjacent cell has multiple connectors, we need to be careful
						if adjacentConnectorCount > 1 {
							// Check if any of its connectors are connected to other cells
							isConnectedToOthers := false

							// For each connector in the adjacent cell
							for _, adjConn := range adjacentRotatedConnectors {
								// Skip the connector that should point to our cell
								if adjConn == oppositeConnector {
									continue
								}

								// Calculate the direction for this connector
								var adjDx, adjDy int
								switch adjConn {
								case ConnectorTop:
									adjDx, adjDy = 0, -1
								case ConnectorRight:
									adjDx, adjDy = 1, 0
								case ConnectorBottom:
									adjDx, adjDy = 0, 1
								case ConnectorLeft:
									adjDx, adjDy = -1, 0
								}

								// Calculate the position this connector points to
								adjNx, adjNy := nx+adjDx, ny+adjDy

								// Check if there's a cell there with a connector pointing back
								otherCell := worldMap.GetCell(adjNx, adjNy)
								if otherCell != nil && len(otherCell.Snippet.Connectors) > 0 {
									otherRotatedConnectors := otherCell.GetRotatedConnectors()

									// Determine the opposite connector
									var otherOppositeConnector SnippetConnector
									switch adjConn {
									case ConnectorTop:
										otherOppositeConnector = ConnectorBottom
									case ConnectorRight:
										otherOppositeConnector = ConnectorLeft
									case ConnectorBottom:
										otherOppositeConnector = ConnectorTop
									case ConnectorLeft:
										otherOppositeConnector = ConnectorRight
									}

									// Check if the other cell has a connector pointing back
									for _, otherConn := range otherRotatedConnectors {
										if otherConn == otherOppositeConnector {
											isConnectedToOthers = true
											break
										}
									}

									if isConnectedToOthers {
										break
									}
								}
							}

							// If the adjacent cell is connected to other cells, it's better to modify our cell
							if isConnectedToOthers {
								// Find all snippets that don't have this connector
								var alternativeSnippets []*WorldSnippet
								for _, s := range g.Registry.Snippets {
									hasConnector := false
									for _, c := range s.Connectors {
										if c == conn {
											hasConnector = true
											break
										}
									}
									if !hasConnector {
										alternativeSnippets = append(alternativeSnippets, s)
									}
								}

								if len(alternativeSnippets) > 0 {
									// Choose a snippet without the problematic connector
									cell.Snippet = alternativeSnippets[0]
									cell.Rotation = 0
									// Skip this connector since we've replaced the cell
									continue
								}
								// If no alternative snippets found, continue with the next connector
								continue
							}
						}

						// If we get here, we should replace the adjacent cell
						snippets := g.Registry.GetSnippetsByConnector(oppositeConnector)
						if len(snippets) == 0 {
							// If no snippets found with the required connector, we need to replace the current cell
							// with a snippet that doesn't have this connector
							// First, find all snippets that don't have this connector
							var alternativeSnippets []*WorldSnippet
							for _, s := range g.Registry.Snippets {
								hasConnector := false
								for _, c := range s.Connectors {
									if c == conn {
										hasConnector = true
										break
									}
								}
								if !hasConnector {
									alternativeSnippets = append(alternativeSnippets, s)
								}
							}

							if len(alternativeSnippets) > 0 {
								// Choose a snippet without the problematic connector
								cell.Snippet = alternativeSnippets[0]
								cell.Rotation = 0
								// Skip this connector since we've replaced the cell
								continue
							}
							// If no alternative snippets found, continue with the next connector
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
						adjacentCell.BranchDepth = cell.BranchDepth + 1
					}
				}
			}
		}
	}

	// One more pass to ensure all snippets are exactly adjacent to each other (no gaps or overlaps)
	// This is done by checking all cells and ensuring they have neighbors in all four directions
	// This satisfies requirement 1: All snippets must be exactly next to each other, no free spaces or overlaps are allowed
	for y := minY + 1; y <= maxY-1; y++ {
		for x := minX + 1; x <= maxX-1; x++ {
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

	// Verify that all paths are connected to the main path
	// This satisfies requirement 5: Everything must start from the main path. Therefore, all paths are always connected to it!
	err := g.verifyAllPathsConnectedToMainPath(worldMap, emptySnippet)
	if err != nil {
		return fmt.Errorf("failed to verify all paths are connected to main path: %w", err)
	}

	return nil
}

// verifyAllPathsConnectedToMainPath verifies that all paths are connected to the main path
// If any paths are not connected, they are either connected to the main path or replaced with empty snippets
func (g *WorldGenerator) verifyAllPathsConnectedToMainPath(worldMap *WorldMap, emptySnippet *WorldSnippet) error {
	// If there are no main path cells, we can't verify connectivity
	if len(worldMap.MainPathCells) == 0 {
		return fmt.Errorf("no main path cells found")
	}

	// Create a map to track visited cells
	visited := make(map[string]bool)

	// Start with all main path cells
	queue := make([]*WorldCell, 0, len(worldMap.MainPathCells))
	for _, cell := range worldMap.MainPathCells {
		key := cell.GetKey()
		visited[key] = true
		queue = append(queue, cell)
	}

	// Perform breadth-first search to find all cells reachable from the main path
	for len(queue) > 0 {
		// Dequeue a cell
		cell := queue[0]
		queue = queue[1:]

		// Check all four directions
		directions := [][2]int{
			{0, -1}, // Top
			{1, 0},  // Right
			{0, 1},  // Bottom
			{-1, 0}, // Left
		}

		// Get the rotated connectors for this cell
		rotatedConnectors := cell.GetRotatedConnectors()

		// For each direction, check if there's a connector pointing that way
		for i, dir := range directions {
			// Calculate the adjacent position
			nx, ny := cell.X+dir[0], cell.Y+dir[1]

			// Skip if we've already visited this cell
			adjacentKey := fmt.Sprintf("%d,%d", nx, ny)
			if visited[adjacentKey] {
				continue
			}

			// Get the adjacent cell
			adjacentCell := worldMap.GetCell(nx, ny)
			if adjacentCell == nil {
				continue
			}

			// Check if this cell has a connector pointing to the adjacent cell
			hasConnector := false
			connectorDirections := []SnippetConnector{
				ConnectorTop,
				ConnectorRight,
				ConnectorBottom,
				ConnectorLeft,
			}
			for _, conn := range rotatedConnectors {
				if conn == connectorDirections[i] {
					hasConnector = true
					break
				}
			}

			// If this cell has a connector pointing to the adjacent cell,
			// and the adjacent cell has a connector pointing back,
			// mark the adjacent cell as visited and add it to the queue
			if hasConnector {
				// Check if the adjacent cell has a connector pointing back
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

				if hasOppositeConnector {
					visited[adjacentKey] = true
					queue = append(queue, adjacentCell)
				}
			}
		}
	}

	// Check if all cells with connectors are reachable from the main path
	for _, cell := range worldMap.Cells {
		// Skip cells with no connectors (empty snippets)
		if len(cell.Snippet.Connectors) == 0 {
			continue
		}

		// Skip cells that are already reachable from the main path
		key := cell.GetKey()
		if visited[key] {
			continue
		}

		// This cell has connectors but is not reachable from the main path
		// Replace it with an empty snippet
		cell.Snippet = emptySnippet
		cell.Rotation = 0
		cell.IsMainPath = false
		cell.BranchDepth = 0
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
