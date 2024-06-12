package main

import (
	"Cloak/extension"
	"fyne.io/systray"
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
