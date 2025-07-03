package shaders

import "github.com/hajimehoshi/ebiten/v2"

var brightnessShader *ebiten.Shader

const brightnessShaderSrc = `//kage:unit pixels
package main

uniform vec2 Center
uniform float Radius
uniform float MinBrightness

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
    clr := imageSrc0At(srcPos)
    dist := distance(dstPos.xy, Center)
    t := clamp(dist/Radius, 0.0, 1.0)
    b := mix(1.0, MinBrightness, t)
    clr.rgb *= b
    return clr * color
}
`

func init() {
	var err error
	brightnessShader, err = ebiten.NewShader([]byte(brightnessShaderSrc))
	if err != nil {
		panic(err)
	}
}

// ApplyBrightness draws src onto dst with a radial brightness falloff.
func ApplyBrightness(dst, src *ebiten.Image, centerX, centerY, radius, minBrightness float64) {
	if dst == nil || src == nil {
		return
	}
	op := &ebiten.DrawRectShaderOptions{}
	op.Images[0] = src
	op.Uniforms = map[string]interface{}{
		"Center":        []float32{float32(centerX), float32(centerY)},
		"Radius":        float32(radius),
		"MinBrightness": float32(minBrightness),
	}
	w, h := src.Size()
	dst.DrawRectShader(w, h, brightnessShader, op)
}
