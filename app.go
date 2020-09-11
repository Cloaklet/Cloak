package main

import (
	"Cloak/icons"
	"Cloak/models"
	"database/sql"
	"github.com/getlantern/systray"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/browser"
)

type App struct {
	apiServer *ApiServer
	repo      *models.VaultRepo
	db        *sql.DB
}

func (a *App) stop() {
	a.db.Close()
	a.apiServer.Stop()
}

func (a *App) Migrate() {
	driver, err := sqlite3.WithInstance(a.db, &sqlite3.Config{})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to wrap migration instance")
	}
	migration, err := migrate.NewWithDatabaseInstance(
		"file://./migrations", "sqlite", driver,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to init migration")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal().Err(err).Msg("Failed to run migrations")
	}
	logger.Info().Msg("Migration OK")
}

// NewApp constructs and returns a new App instance
func NewApp() *App {
	app := &App{}

	var err error
	app.db, err = sql.Open("sqlite3", "./test.db")
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

	// FIXME
	app.apiServer = NewApiServer("/bin/gocryptfs", app.repo)
	go func() {
		app.apiServer.Start("127.0.0.1:9763")
	}()
	return app
}
