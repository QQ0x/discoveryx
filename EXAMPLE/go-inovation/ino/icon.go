package ino

import (
	"image"
	_ "image/png"
	"path"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/go-inovation/ino/internal/assets"
)

func setIcons() error {
	const dir = "images/icons"

	ents, err := assets.Assets.ReadDir(dir)
	if err != nil {
		return err
	}

	var icons []image.Image
	for _, ent := range ents {
		name := ent.Name()
		ext := filepath.Ext(name)
		if ext != ".png" {
			continue
		}

		f, err := assets.Assets.Open(path.Join(dir, name))
		if err != nil {
			return err
		}
		defer f.Close()

		img, _, err := image.Decode(f)
		if err != nil {
			return err
		}
		icons = append(icons, img)
	}

	ebiten.SetWindowIcon(icons)
	return nil
}
