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
	var app *App
	systray.Run(func() {
		app = NewApp()
		logger.Info().Msg("Ready")
	}, func() {
		app.stop()
		logger.Info().Msg("Quit")
	})
}
