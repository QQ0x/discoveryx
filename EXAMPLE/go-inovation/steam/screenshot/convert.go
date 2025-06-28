package main

import (
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"os"
	"regexp"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var regularTermination = errors.New("regular termiination")

var reFilename = regexp.MustCompile(`^screenshot([^_]+)_([a-z]+)\.png$`)

type Game struct {
	tmpImage *ebiten.Image
}

func (g *Game) getTmpImage(width, height int) *ebiten.Image {
	if g.tmpImage != nil {
		if w, h := g.tmpImage.Size(); w == width && h == height {
			g.tmpImage.Clear()
			return g.tmpImage
		}
	}
	g.tmpImage = ebiten.NewImage(width, height)
	return g.tmpImage
}

func (g *Game) Update() error {
	ents, err := os.ReadDir(".")
	if err != nil {
		return err
	}

	const (
		screenshotWidth  = 1920
		screenshotHeight = 1080
	)

	offscreen := ebiten.NewImage(screenshotWidth, screenshotHeight)
	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		m := reFilename.FindStringSubmatch(name)
		if m == nil {
			continue
		}

		img, _, err := ebitenutil.NewImageFromFile(name)
		if err != nil {
			return err
		}

		w, h := img.Size()
		scale := screenshotHeight / float64(h)
		iscale := math.Ceil(scale)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(iscale, iscale)
		tmpImg := g.getTmpImage(w * int(iscale), h * int(iscale))
		tmpImg.DrawImage(img, op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(scale) / float64(iscale), float64(scale) / float64(iscale))
		op.GeoM.Translate((screenshotWidth - float64(w) * scale) / 2, 0)
		op.Filter = ebiten.FilterLinear
		offscreen.Clear()
		offscreen.Fill(color.Black)
		offscreen.DrawImage(tmpImg, op)

		newName := fmt.Sprintf("screenshot%s_%dx%d_%s.png", m[1], screenshotWidth, screenshotHeight, m[2])
		f, err := os.Create(newName)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := png.Encode(f, offscreen); err != nil {
			return err
		}
	}
	return regularTermination
}

func (*Game) Draw(screen *ebiten.Image) {
}

func (*Game) Layout(width, height int) (int, int) {
	return width, height
}

func main() {
	ebiten.SetWindowDecorated(false)
	ebiten.SetScreenTransparent(true)
	if err := ebiten.RunGame(&Game{}); err != nil && err != regularTermination {
		panic(err)
	}
}
