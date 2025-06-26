package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"runtime/trace"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/go-inovation/ino"
)

var (
	memProfile  = flag.String("memprofile", "", "write memory profile to file")
	traceOut    = flag.String("trace", "", "write trace to file")
	transparent = flag.Bool("transparent", false, "background transparency")
)

func main() {
	flag.Parse()

	if *traceOut != "" {
		f, err := os.Create(*traceOut)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		trace.Start(f)
		defer trace.Stop()
	}

	game, err := ino.NewGame()
	if err != nil {
		panic(err)
	}

	if *transparent {
		ebiten.SetScreenTransparent(true)
		ebiten.SetWindowDecorated(false)
		game.SetTransparent()
	}

	const scale = 2
	ebiten.SetWindowSize(ino.ScreenWidth*scale, ino.ScreenHeight*scale)
	ebiten.SetWindowResizable(true)
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := pprof.WriteHeapProfile(f); err != nil {
			panic(fmt.Sprintf("could not write memory profile: %s", err))
		}
	}
}
