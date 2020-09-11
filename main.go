package main

import (
	"github.com/getlantern/systray"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"os"
	"time"
)

var logger zerolog.Logger

func init() {
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	logger = zlog.Logger.With().Str("module", "main").Logger()
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
