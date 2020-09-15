package main

import (
	"Cloak/extension"
	"Cloak/icons"
	"Cloak/models"
	"Cloak/server"
	"database/sql"
	"github.com/getlantern/systray"
	"github.com/lopezator/migrator"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/browser"
	"path/filepath"
)

var ReleaseMode string

// A dumb debug logger for migrator
type logPrinter struct{}

func (l *logPrinter) Printf(f string, v ...interface{}) {
	logger.Debug().Msgf(f, v...)
}

type App struct {
	apiServer   *server.ApiServer
	repo        *models.VaultRepo
	db          *sql.DB
	releaseMode bool
}

func (a *App) stop() {
	a.db.Close()
	a.apiServer.Stop()
}

func (a *App) Migrate() {
	m, err := migrator.New(
		migrator.WithLogger(&logPrinter{}),
		migrator.Migrations(
			&migrator.Migration{
				Name: "Create vaults table",
				Func: func(tx *sql.Tx) error {
					_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS vaults (
    id INTEGER PRIMARY KEY,
    path TEXT NOT NULL UNIQUE,
    mountpoint TEXT UNIQUE
);`)
					return err
				},
			},
		),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to init migrations")
	}

	if err = m.Migrate(a.db); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run migrations")
	} else {
		logger.Debug().Msg("Migration ok")
	}
}

// NewApp constructs and returns a new App instance
func NewApp() *App {
	app := &App{releaseMode: ReleaseMode == "true"}
	appDataDir, err := extension.GetAppDataDirectory()
	if err != nil {
		logger.Fatal().Err(err).
			Str("appDataDir", appDataDir).
			Msg("Failed to get application data directory")
	} else {
		logger.Info().Str("appDataDir", appDataDir).Msg("Determined application data directory")
	}

	app.db, err = sql.Open("sqlite3", filepath.Join(appDataDir, "vaults.db"))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to open vault database")
	}

	app.Migrate()
	app.repo = models.NewVaultRepo(app.db)

	systray.SetTemplateIcon(icons.TRAY_TPL, icons.TRAY)
	systray.SetTooltip("Cloak - a gocryptfs GUI")
	openBrowser := systray.AddMenuItem("Open", "")
	quit := systray.AddMenuItem("Quit", "")
	go func() {
		for {
			select {
			case <-quit.ClickedCh:
				systray.Quit()
			case <-openBrowser.ClickedCh:
				browser.OpenURL("http://127.0.0.1:9763")
			}
		}
	}()

	app.apiServer = server.NewApiServer(app.repo, app.releaseMode)
	go func() {
		app.apiServer.Start("127.0.0.1:9763")
	}()
	return app
}
