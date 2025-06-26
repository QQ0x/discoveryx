package assets

import (
	"embed"
)

//go:embed images/* sound/*
var Assets embed.FS
