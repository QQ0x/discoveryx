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

// ImageCache provides a thread-safe cache for loading and storing images
type ImageCache struct {
	cache map[string]*ebiten.Image
	mutex sync.RWMutex
}

// Global image cache instance
var imageCache = &ImageCache{
	cache: make(map[string]*ebiten.Image),
}

// LoadImage loads an image from the embedded filesystem
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

// GetImage retrieves an image from the cache, loading it if necessary
func GetImage(path string) *ebiten.Image {
	// First check if the image is already in the cache
	imageCache.mutex.RLock()
	img, exists := imageCache.cache[path]
	imageCache.mutex.RUnlock()

	if exists {
		return img
	}

	// If not in cache, load it
	img = LoadImage(path)

	// Store in cache
	imageCache.mutex.Lock()
	imageCache.cache[path] = img
	imageCache.mutex.Unlock()

	return img
}

// Path constants for commonly used assets
const (
	PlayerSpritePath    = "images/gameScene/Ships/spaceShips_001.png"
	StartBackgroundPath = "images/startScene/startpage_background.png"
	PlayButtonPath      = "images/startScene/play_button.png"
	GameBackgroundPath  = "images/gameScene/Background/background.png"
)

// Smaller assets loaded at startup
var (
	PlayerSprite = LoadImage(PlayerSpritePath)
	PlayButton   = LoadImage(PlayButtonPath)
)

// GetStartBackground returns the start background image, loading it if necessary
func GetStartBackground() *ebiten.Image {
	return GetImage(StartBackgroundPath)
}

// GetGameBackground returns the game background image, loading it if necessary
func GetGameBackground() *ebiten.Image {
	return GetImage(GameBackgroundPath)
}
