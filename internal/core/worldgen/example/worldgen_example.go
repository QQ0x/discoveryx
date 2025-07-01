package main

import (
	"discoveryx/internal/core/worldgen"
	"fmt"
	"os"
)

func main() {
	// Create a new world generator
	generator, err := worldgen.NewWorldGenerator()
	if err != nil {
		fmt.Printf("Error creating world generator: %v\n", err)
		os.Exit(1)
	}

	// Use default configuration
	config := worldgen.DefaultWorldGenConfig()

	// Create a new generated world
	world, err := worldgen.NewGeneratedWorld(1000, 1000, generator, config)
	if err != nil {
		fmt.Printf("Error creating generated world: %v\n", err)
		os.Exit(1)
	}

	// Get the world map
	worldMap := world.GetWorldMap()

	// Print some statistics about the generated world
	fmt.Printf("Generated world with %d cells\n", worldMap.GetCellCount())
	fmt.Printf("Main path length: %d\n", worldMap.GetMainPathLength())
	fmt.Printf("Branch count: %d\n", worldMap.GetBranchCount())

	// Print the layout of the world map
	printWorldMap(worldMap)
}

// printWorldMap prints a simple ASCII representation of the world map
func printWorldMap(worldMap *worldgen.WorldMap) {
	// Find the bounds of the world map
	minX, minY, maxX, maxY := findWorldMapBounds(worldMap)

	// Print the world map
	fmt.Println("World Map:")
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			cell := worldMap.GetCell(x, y)
			if cell == nil {
				fmt.Print(" ")
			} else if cell.IsMainPath {
				fmt.Print("M")
			} else {
				fmt.Print("B")
			}
		}
		fmt.Println()
	}
}

// findWorldMapBounds finds the minimum and maximum coordinates of the world map
func findWorldMapBounds(worldMap *worldgen.WorldMap) (minX, minY, maxX, maxY int) {
	minX, minY = 1000000, 1000000
	maxX, maxY = -1000000, -1000000

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

	return
}
