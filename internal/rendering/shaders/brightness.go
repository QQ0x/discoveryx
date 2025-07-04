// Package shaders provides GPU-accelerated visual effects for the game.
// It implements various shader programs that can be applied to game scenes
// to create atmospheric lighting, post-processing effects, and visual enhancements.
//
// The shaders in this package are written in the Kage shader language, which is
// Ebiten's shader language based on a subset of Go. Each shader is encapsulated
// in a struct that provides a clean interface for setting parameters and applying
// the effect to game scenes.
//
// Key features of the shaders package:
// - Performance-optimized GPU-based rendering
// - Configurable parameters for customizing effects
// - Frame-rate independent visual effects
// - Support for both desktop and mobile platforms
//
// Common uses include creating lighting effects, visibility systems,
// environmental atmospherics, and special visual effects for gameplay events.
package shaders

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// BrightnessShader provides a radial brightness effect around a focus point.
// The brightness is 100% at the player's position and fades exponentially to 0% at the
// specified radius, creating a very steep falloff effect.
//
// This shader is primarily used to create:
// - A visibility/fog-of-war system where only areas near the player are fully visible
// - Dynamic lighting effects that follow the player
// - Atmospheric darkness that enhances the game's mood
// - Visual indicators of special areas or events
//
// The exponential falloff creates a more realistic lighting effect than linear
// or quadratic falloff, as it better simulates how light intensity diminishes
// with distance in the real world.
//
// Usage:
//
//	bs, _ := shaders.NewBrightnessShader()
//	op := &ebiten.DrawRectShaderOptions{}
//	op.Images[0] = srcImage
//	op.Uniforms = map[string]any{
//	    "PlayerPos": []float32{playerX, playerY},
//	    "Radius":    radius,
//	}
//	screen.DrawRectShader(width, height, bs.Shader(), op)
type BrightnessShader struct {
	shader *ebiten.Shader
}

// NewBrightnessShader compiles and returns a new brightness shader.
func NewBrightnessShader() (*BrightnessShader, error) {
	s, err := ebiten.NewShader([]byte(brightnessShaderSrc))
	if err != nil {
		return nil, err
	}
	return &BrightnessShader{shader: s}, nil
}

// Shader returns the underlying ebiten.Shader.
func (b *BrightnessShader) Shader() *ebiten.Shader {
	return b.shader
}

const brightnessShaderSrc = `//kage:unit pixels
package main

var PlayerPos vec2
var Radius float

func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
    // Calculate distance from the current fragment to the player position.
    dist := distance(position.xy, PlayerPos)

    // Normalize distance to [0,1] based on the radius.
    t := clamp(dist/Radius, 0.0, 1.0)

    // Apply an exponential function to make brightness decrease more rapidly
    // This creates a much steeper falloff than the previous t^2 approach
    t = 1.0 - exp(-5.0 * t * t)

    // Brightness ranges from 1.0 (100%) at the center to 0.0 (0%) at the edge.
    brightness := mix(1.0, 0.0, t)

    // Get the color from the source image
    col := imageSrc0At(texCoord)

    // Apply brightness
    col.rgb *= brightness

    return col
}
`
