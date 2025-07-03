package shaders

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// BrightnessShader provides a radial brightness effect around a focus point.
// The brightness is 100% at the player's position and fades to 5% at the
// specified radius.
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

    // Brightness ranges from 1.0 at the center to 0.05 at the edge.
    brightness := mix(1.0, 0.05, t)

    col := imageSrc0At(texCoord)
    col.rgb *= brightness
    return col
}
`
