package audio

import (
	"bytes"
	"io"
	"path"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	"github.com/hajimehoshi/go-inovation/ino/internal/assets"
)

const sampleRate = 48000

var (
	audioContext = audio.NewContext(sampleRate)
	soundPlayers = map[string]*audio.Player{}
	mute         = false
)

func Mute() {
	mute = true
}

func Load() error {
	const dir = "sound"

	ents, err := assets.Assets.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, ent := range ents {
		name := ent.Name()
		ext := filepath.Ext(name)
		if ext != ".ogg" && ext != ".wav" {
			continue
		}

		f, err := assets.Assets.Open(path.Join(dir, name))
		if err != nil {
			return err
		}
		defer f.Close()

		var s io.ReadSeeker
		switch ext {
		case ".ogg":
			stream, err := vorbis.DecodeWithSampleRate(sampleRate, f)
			if err != nil {
				return err
			}
			bs, err := io.ReadAll(stream)
			if err != nil {
				return err
			}
			s = audio.NewInfiniteLoop(bytes.NewReader(bs), stream.Length())
		case ".wav":
			stream, err := wav.DecodeWithSampleRate(sampleRate, f)
			if err != nil {
				return err
			}
			s = stream
		default:
			panic("invalid file name")
		}

		p, err := audio.NewPlayer(audioContext, s)
		if err != nil {
			return err
		}

		soundPlayers[name] = p
	}
	return nil
}

func Finalize() error {
	for _, p := range soundPlayers {
		if err := p.Close(); err != nil {
			return err
		}
	}
	return nil
}

type BGM string

const (
	BGM0 BGM = "ino1.ogg"
	BGM1 BGM = "ino2.ogg"
)

func SetBGMVolume(volume float64) {
	if mute {
		return
	}
	for _, b := range []BGM{BGM0, BGM1} {
		p := soundPlayers[string(b)]
		if !p.IsPlaying() {
			continue
		}
		p.SetVolume(volume)
		return
	}
}

func PauseBGM() {
	if mute {
		return
	}
	for _, b := range []BGM{BGM0, BGM1} {
		p := soundPlayers[string(b)]
		p.Pause()
	}
}

func ResumeBGM(bgm BGM) {
	if mute {
		return
	}
	PauseBGM()
	p := soundPlayers[string(bgm)]
	p.SetVolume(1)
	p.Play()
}

func PlayBGM(bgm BGM) error {
	if mute {
		return nil
	}
	PauseBGM()
	p := soundPlayers[string(bgm)]
	p.SetVolume(1)
	if err := p.Rewind(); err != nil {
		return err
	}
	p.Play()
	return nil
}

type SE string

const (
	SE_DAMAGE   SE = "damage.wav"
	SE_HEAL     SE = "heal.wav"
	SE_ITEMGET  SE = "itemget.wav"
	SE_ITEMGET2 SE = "itemget2.wav"
	SE_JUMP     SE = "jump.wav"
)

func PlaySE(se SE) {
	if mute {
		return
	}
	p := soundPlayers[string(se)]
	p.Rewind()
	p.Play()
}
