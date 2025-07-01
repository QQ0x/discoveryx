package worldgen

import (
	"discoveryx/internal/assets"
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
	// Define hardcoded metadata for all snippets
	// This is necessary for iOS where file access is restricted
	snippetMetadata := []SnippetMetadata{
		{Filename: "Worldgen_x.png", Connectors: []int{}, Weight: 10}, // Empty snippet with no connectors
		{Filename: "Worldgen_l.png", Connectors: []int{270}, Weight: 10},
		{Filename: "Worldgen_l2.png", Connectors: []int{270}, Weight: 10},
		{Filename: "Worldgen_lo.png", Connectors: []int{270, 0}, Weight: 10},
		{Filename: "Worldgen_lo2.png", Connectors: []int{270, 0}, Weight: 10},
		{Filename: "Worldgen_lou.png", Connectors: []int{270, 0, 180}, Weight: 10},
		{Filename: "Worldgen_lou2.png", Connectors: []int{270, 0, 180}, Weight: 10},
		{Filename: "Worldgen_lou3.png", Connectors: []int{270, 0, 180}, Weight: 10},
		{Filename: "Worldgen_lr.png", Connectors: []int{270, 90}, Weight: 10},
		{Filename: "Worldgen_lr2.png", Connectors: []int{270, 90}, Weight: 10},
		{Filename: "Worldgen_lr3.png", Connectors: []int{270, 90}, Weight: 10},
		{Filename: "Worldgen_lr4.png", Connectors: []int{270, 90}, Weight: 10},
		{Filename: "Worldgen_lr5.png", Connectors: []int{270, 90}, Weight: 10},
		{Filename: "Worldgen_lu.png", Connectors: []int{270, 180}, Weight: 10},
	}

	// Process each metadata entry
	for _, metadata := range snippetMetadata {
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
		imagePath := filepath.Join(imageDir, metadata.Filename)
		img, err := r.loadImage(imagePath)
		if err != nil {
			return fmt.Errorf("failed to load image from %s: %w", imagePath, err)
		}
		snippet.Image = img

		// Add to registry
		r.addSnippet(snippet)
	}

	return nil
}

// Note: loadMetadata method has been removed as we now use hardcoded metadata

// loadImage loads a single image file
func (r *SnippetRegistry) loadImage(path string) (*ebiten.Image, error) {
	// Extract just the filename from the path
	_, filename := filepath.Split(path)

	// Construct the path for the embedded filesystem
	assetPath := filepath.Join("images/gameScene/World", filename)

	// Use the assets package to load the image
	img := assets.GetImage(assetPath)
	return img, nil
}

// addSnippet adds a snippet to the registry and indexes it
func (r *SnippetRegistry) addSnippet(snippet *WorldSnippet) {
	// Add to main map
	r.Snippets[snippet.Filename] = snippet

	// Index by connector
	for _, conn := range snippet.Connectors {
		r.ByConnector[conn] = append(r.ByConnector[conn], snippet)
	}
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
