// Package assets provides functionality for loading, caching, and managing game assets.
// It handles the efficient loading and storage of images, sounds, and other resources
// needed by the game. The package uses Go's embed system to include assets directly
// in the binary, ensuring they're always available regardless of the deployment platform.
//
// Key features of the assets package:
// - Embedded resources for cross-platform compatibility
// - Thread-safe caching to improve performance
// - Lazy loading to reduce memory usage
// - Centralized access to common game assets
// - Path constants for consistent asset referencing
package assets

import (
	"embed"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	_ "image/png"
	"sync"
)

//go:embed images
var Assets embed.FS

// ImageCache provides a thread-safe cache for loading and storing images.
// This struct implements a simple but effective caching system that:
// - Stores loaded images in memory to avoid repeated disk/embed access
// - Uses read-write mutexes to ensure thread safety in concurrent contexts
// - Provides fast lookups for frequently used assets
//
// The cache is particularly important for mobile platforms where
// asset loading can be more expensive and memory management is critical.
type ImageCache struct {
	cache map[string]*ebiten.Image // Map of image paths to loaded images
	mutex sync.RWMutex             // Mutex to protect concurrent access to the cache
}

// Global image cache instance used throughout the game
// This singleton approach provides a centralized cache that can be
// accessed from anywhere in the codebase without passing references.
var imageCache = &ImageCache{
	cache: make(map[string]*ebiten.Image),
}

// LoadImage loads an image from the embedded filesystem.
// This function handles the direct loading of images from the embedded
// filesystem without caching. It's used internally by GetImage and for
// preloading essential assets at startup.
//
// The function performs several steps:
// 1. Opens the file from the embedded filesystem
// 2. Decodes the image data into a Go image.Image
// 3. Converts it to an Ebiten image for rendering
//
// Parameters:
// - name: The path to the image within the embedded filesystem
//
// Returns:
// - An Ebiten image ready for rendering
//
// Note: This function panics if the image cannot be found or decoded,
// as missing assets are considered a critical error that prevents proper
// game operation.
func LoadImage(name string) *ebiten.Image {
	f, err := Assets.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

// GetImage retrieves an image from the cache, loading it if necessary.
// This is the primary method for accessing images throughout the game.
// It implements a lazy loading strategy that:
// - Returns cached images immediately if they've been loaded before
// - Loads and caches images on first access
// - Ensures thread safety for concurrent access
//
// This approach optimizes both memory usage and performance by:
// - Only loading assets when they're actually needed
// - Avoiding redundant loading of the same asset
// - Providing fast access to frequently used assets
//
// Parameters:
// - path: The path to the image within the embedded filesystem
//
// Returns:
// - An Ebiten image ready for rendering
func GetImage(path string) *ebiten.Image {
	// First check if the image is already in the cache using a read lock
	// This allows multiple goroutines to read from the cache simultaneously
	imageCache.mutex.RLock()
	img, exists := imageCache.cache[path]
	imageCache.mutex.RUnlock()

	if exists {
		return img
	}

	// If not in cache, load it from the embedded filesystem
	img = LoadImage(path)

	// Store in cache using a write lock to ensure thread safety
	// This prevents race conditions when multiple goroutines try to
	// load the same image simultaneously
	imageCache.mutex.Lock()
	imageCache.cache[path] = img
	imageCache.mutex.Unlock()

	return img
}

// Path constants for commonly used assets.
// These constants provide a centralized place to define asset paths,
// making it easier to:
// - Maintain consistent naming across the codebase
// - Update paths if the asset structure changes
// - Prevent typos and errors in path strings
//
// Using constants instead of hardcoded strings also enables IDE features
// like auto-completion and refactoring support.
const (
	PlayerSpritePath    = "images/gameScene/Ships/spaceShips_001.png"
	StartBackgroundPath = "images/startScene/startpage_background.png"
	PlayButtonPath      = "images/startScene/play_button.png"
	GameBackgroundPath  = "images/gameScene/Background/background.png"
	PlayerBulletPath    = "images/gameScene/Bullets/PlayerBullets/PlayerBullet_Single.png"
	EnemyBulletPath     = "images/gameScene/Bullets/EnemyBullets/EnemyBullet_Single.png"
)

// Preloaded assets that are used frequently throughout the game.
// These assets are loaded at startup rather than on-demand because:
// - They're used immediately when the game starts
// - They're used frequently during gameplay
// - Preloading reduces hitches and stutters during gameplay
//
// This approach balances memory usage with performance by only
// preloading the most essential assets.
var (
	PlayerSprite = LoadImage(PlayerSpritePath) // The player's ship sprite
	PlayButton   = LoadImage(PlayButtonPath)   // The play button on the start screen
	PlayerBullet = LoadImage(PlayerBulletPath) // The player's bullet sprite
	EnemyBullet  = LoadImage(EnemyBulletPath)  // The enemy's bullet sprite
)

// GetStartBackground returns the start background image, loading it if necessary.
// This function provides a convenient way to access the start screen background,
// which is a larger asset that's only loaded when needed (typically once per game session).
//
// The background is loaded on first access and then cached for subsequent calls,
// balancing memory usage with loading performance.
//
// Returns:
// - An Ebiten image of the start screen background
func GetStartBackground() *ebiten.Image {
	return GetImage(StartBackgroundPath)
}

// GetGameBackground returns the game background image, loading it if necessary.
// This function provides a convenient way to access the main game background,
// which is a larger asset that's only loaded when needed (when entering gameplay).
//
// The background is loaded on first access and then cached for subsequent calls,
// balancing memory usage with loading performance.
//
// Returns:
// - An Ebiten image of the main game background
func GetGameBackground() *ebiten.Image {
	return GetImage(GameBackgroundPath)
}
