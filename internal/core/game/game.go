package game

import (
	// "discoveryx/internal/core/ecs"
	"discoveryx/internal/core/gameplay/player"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
)

// Game is the main game struct
type Game struct {
	player *player.Player
	width  int
	height int
}

// GetWidth returns the width of the game world
func (g *Game) GetWidth() int {
	return g.width
}

// GetHeight returns the height of the game world
func (g *Game) GetHeight() int {
	return g.height
}

// New creates a new game instance
func New() *Game {
	g := &Game{
		width:  640,
		height: 480,
	}

	// Load player sprite and create player
	g.player = player.NewPlayer(g)

	return g
}

// Update updates the game state
func (g *Game) Update() error {
	return g.player.Update()
}

// Draw draws the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen with a dark color
	screen.Fill(color.RGBA{20, 20, 20, 255})

	// Draw the player
	g.player.Draw(screen)
}

// Layout implements ebiten.Game's Layout method
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	// Update game dimensions
	g.width = outsideWidth
	g.height = outsideHeight

	return outsideWidth, outsideHeight
}
