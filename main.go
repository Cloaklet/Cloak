package main

import (
	"Cloak/extension"
	"github.com/getlantern/systray"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	logger = extension.GetLogger("main")
}

func main() {
	app := NewApp()
	systray.Run(app.Start, app.Stop)
}
