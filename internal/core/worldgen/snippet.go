package worldgen

import (
	"discoveryx/internal/assets"
	"encoding/json"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"math/rand"
	"path/filepath"
)

// SnippetConnector represents the position of a connector on a world snippet
// 0: Top connector
// 90: Right connector
// 180: Bottom connector
// 270: Left connector
type SnippetConnector int

const (
	ConnectorTop    SnippetConnector = 0
	ConnectorRight  SnippetConnector = 90
	ConnectorBottom SnippetConnector = 180
	ConnectorLeft   SnippetConnector = 270
)

// SnippetType represents the type of a world snippet based on its function
type SnippetType string

const (
	SnippetTypePath     SnippetType = "path"     // A path segment with two connectors
	SnippetTypeJunction SnippetType = "junction" // A junction with three or more connectors
	SnippetTypeDeadEnd  SnippetType = "dead-end" // A dead-end with only one connector
)

// WorldSnippet represents a single world snippet with its metadata and image
type WorldSnippet struct {
	Filename   string             // The filename of the snippet image
	Connectors []SnippetConnector // The connectors this snippet has
	Weight     int                // The relative probability weight for selection
	Image      *ebiten.Image      // The loaded image
	Walls      []WallPoint        // The wall points detected in this snippet
}

// GetType returns the type of the snippet based on the number of connectors
func (s *WorldSnippet) GetType() SnippetType {
	switch len(s.Connectors) {
	case 1:
		return SnippetTypeDeadEnd
	case 2:
		return SnippetTypePath
	default:
		return SnippetTypeJunction
	}
}

// SnippetMetadata represents the JSON structure of a snippet metadata file
type SnippetMetadata struct {
	Filename   string `json:"filename"`
	Connectors []int  `json:"connectors"`
	Weight     int    `json:"weight"`
}

// SnippetRegistry manages all available world snippets and provides methods to access them
type SnippetRegistry struct {
	Snippets    map[string]*WorldSnippet                // Map of snippets by filename
	ByConnector map[SnippetConnector][]*WorldSnippet    // Map of snippets by connector
}

// NewSnippetRegistry creates a new empty snippet registry
func NewSnippetRegistry() *SnippetRegistry {
	return &SnippetRegistry{
		Snippets:    make(map[string]*WorldSnippet),
		ByConnector: make(map[SnippetConnector][]*WorldSnippet),
	}
}

// LoadSnippets loads all snippet metadata and images from the specified directory
func (r *SnippetRegistry) LoadSnippets(metadataDir, imageDir string) error {
	// Load metadata from JSON files
	metadataFiles, err := assets.Assets.ReadDir(metadataDir)
	if err != nil {
		return fmt.Errorf("failed to read metadata directory: %w", err)
	}

	// Process each metadata file
	for _, metadataFile := range metadataFiles {
		if filepath.Ext(metadataFile.Name()) != ".json" {
			continue
		}

		// Load metadata from JSON file
		metadataPath := filepath.Join(metadataDir, metadataFile.Name())
		metadata, err := r.loadMetadata(metadataPath)
		if err != nil {
			return fmt.Errorf("failed to load metadata from %s: %w", metadataPath, err)
		}

		// Create WorldSnippet from metadata
		snippet := &WorldSnippet{
			Filename:   metadata.Filename,
			Weight:     metadata.Weight,
			Connectors: make([]SnippetConnector, len(metadata.Connectors)),
		}

		// Convert connector integers to SnippetConnector type
		for i, conn := range metadata.Connectors {
			snippet.Connectors[i] = SnippetConnector(conn)
		}

		// Load the image
		imagePath := filepath.Join("images/gameScene/World", metadata.Filename)
		img := assets.GetImage(imagePath)
		snippet.Image = img

		// Add to registry
		r.addSnippet(snippet)
	}

	return nil
}

// loadMetadata loads a single metadata file
func (r *SnippetRegistry) loadMetadata(path string) (*SnippetMetadata, error) {
	// Read the metadata file
	data, err := assets.Assets.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	// Parse the JSON data
	var metadata SnippetMetadata
	err = json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata JSON: %w", err)
	}

	return &metadata, nil
}

// addSnippet adds a snippet to the registry and indexes it
func (r *SnippetRegistry) addSnippet(snippet *WorldSnippet) {
	// Add to main map
	r.Snippets[snippet.Filename] = snippet

	// Index by connector
	for _, conn := range snippet.Connectors {
		r.ByConnector[conn] = append(r.ByConnector[conn], snippet)
	}

	// Detect walls in the snippet
	snippet.Walls = DetectWallsInSnippet(snippet)
}

// GetSnippetsByConnector returns all snippets that have the specified connector
func (r *SnippetRegistry) GetSnippetsByConnector(connector SnippetConnector) []*WorldSnippet {
	return r.ByConnector[connector]
}

// GetSnippet returns a snippet by its filename
func (r *SnippetRegistry) GetSnippet(filename string) *WorldSnippet {
	return r.Snippets[filename]
}

// SelectWeightedSnippet selects a snippet from the provided list based on their weights
func SelectWeightedSnippet(snippets []*WorldSnippet, rng *rand.Rand) *WorldSnippet {
	return SelectWeightedSnippetWithTypeWeights(snippets, nil, rng)
}

// SelectWeightedSnippetWithTypeWeights selects a snippet from the provided list based on their weights and type weights
func SelectWeightedSnippetWithTypeWeights(snippets []*WorldSnippet, typeWeights map[SnippetType]int, rng *rand.Rand) *WorldSnippet {
	if len(snippets) == 0 {
		return nil
	}

	// Calculate total weight
	totalWeight := 0
	effectiveWeights := make([]int, len(snippets))

	for i, snippet := range snippets {
		effectiveWeight := snippet.Weight

		// Apply type weight if available
		if typeWeights != nil {
			if typeWeight, ok := typeWeights[snippet.GetType()]; ok {
				effectiveWeight *= typeWeight
			}
		}

		effectiveWeights[i] = effectiveWeight
		totalWeight += effectiveWeight
	}

	// If all weights are 0, select randomly
	if totalWeight == 0 {
		return snippets[rng.Intn(len(snippets))]
	}

	// Select based on weight
	randomWeight := rng.Intn(totalWeight)
	currentWeight := 0

	for i, snippet := range snippets {
		currentWeight += effectiveWeights[i]
		if randomWeight < currentWeight {
			return snippet
		}
	}

	// Fallback (should never happen)
	return snippets[len(snippets)-1]
}
