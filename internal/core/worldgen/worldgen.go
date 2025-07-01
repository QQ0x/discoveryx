package worldgen

import (
	"fmt"
	"path/filepath"
)

// WorldGenerator manages the generation of the game world using world snippets
type WorldGenerator struct {
	Registry *SnippetRegistry
}

// NewWorldGenerator creates a new world generator and initializes the snippet registry
func NewWorldGenerator() (*WorldGenerator, error) {
	registry := NewSnippetRegistry()
	
	// Define paths for metadata and image directories
	metadataDir := filepath.Join("internal", "assets", "images", "gameScene", "World", "metadata")
	imageDir := filepath.Join("internal", "assets", "images", "gameScene", "World")
	
	// Load snippets
	err := registry.LoadSnippets(metadataDir, imageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load world snippets: %w", err)
	}
	
	return &WorldGenerator{
		Registry: registry,
	}, nil
}

// Initialize initializes the world generator with the specified metadata and image directories
func (g *WorldGenerator) Initialize(metadataDir, imageDir string) error {
	// Create a new registry
	g.Registry = NewSnippetRegistry()
	
	// Load snippets
	err := g.Registry.LoadSnippets(metadataDir, imageDir)
	if err != nil {
		return fmt.Errorf("failed to load world snippets: %w", err)
	}
	
	return nil
}

// GetSnippetCount returns the number of snippets in the registry
func (g *WorldGenerator) GetSnippetCount() int {
	return len(g.Registry.Snippets)
}

// GetConnectorCount returns the number of connectors in the registry
func (g *WorldGenerator) GetConnectorCount() int {
	return len(g.Registry.ByConnector)
}