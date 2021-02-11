package main

import (
	"Cloak/extension"
	"Cloak/i18n"
	"Cloak/icons"
	"Cloak/models"
	"Cloak/server"
	"database/sql"
	"github.com/getlantern/systray"
	"github.com/lopezator/migrator"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/browser"
	"gopkg.in/ini.v1"
	"os"
	"path/filepath"
)

// A dumb debug logger for migrator
type logPrinter struct{}

func (l *logPrinter) Printf(f string, v ...interface{}) {
	logger.Debug().Msgf(f, v...)
}

// App is the main type to control lifecycle of the whole application
type App struct {
	dataDir     string
	configDir   string
	apiServer   *server.ApiServer
	repo        *models.VaultRepo
	db          *sql.DB
	releaseMode bool
}

// migrate runs database migrations
func (a *App) migrate() {
	m, err := migrator.New(
		migrator.WithLogger(&logPrinter{}),
		migrator.Migrations(migrations...),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to init migrations")
	}

	if err = m.Migrate(a.db); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run migrations")
	} else {
		logger.Debug().Msg("Database migration ok")
	}
}

// Options represents INI-based application settings
type Options struct {
	Locale string `ini:"locale"`
}

func (a *App) loadConfig() {
	appOptions := new(Options)
	if err := ini.MapTo(
		appOptions, filepath.Join(a.configDir, "options.ini"),
	); err != nil && !os.IsNotExist(err) {
		logger.Error().Err(err).Msg("Failed to map app options from config file")
		return
	}
	if appOptions.Locale == "" {
		return
	}

	if err := i18n.GetLocalizer().SetLocale(appOptions.Locale); err != nil {
		logger.Error().Err(err).
			Str("locale", appOptions.Locale).
			Msg("Failed to set locale loaded from config file")
	}
}

// NewApp constructs and returns a new App instance
func NewApp() *App {
	app := &App{releaseMode: extension.ReleaseMode == "true"}

	// Locate data directories
	for dirName, dirPathFunc := range map[string]func() string{
		"appDataDir": extension.GetAppDataDirectory,
		"configDir":  extension.GetConfigDirectory,
	} {
		dirPath, err := extension.EnsureDirectoryExists(dirPathFunc())
		if err != nil {
			logger.Fatal().Err(err).
				Str("name", dirName).
				Str("path", dirPath).
				Msg("Failed to get directory")
		}
		logger.Info().Str("name", dirName).Str("path", dirPath).Msg("Determined directory")
	}
	app.dataDir = extension.GetAppDataDirectory()
	app.configDir = extension.GetConfigDirectory()

	// Load database
	var err error
	app.db, err = sql.Open("sqlite3", filepath.Join(app.dataDir, "vaults.db"))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to open vault database")
	}

	// Migrate database
	app.migrate()
	app.repo = models.NewVaultRepo(app.db)

	// Load app config
	app.loadConfig()

	app.apiServer = server.NewApiServer(app.repo, app.releaseMode)

	logger.Debug().Msg("App created")
	return app
}

// Start starts the app.
// It does not wait for the app to exit, so an external event loop must be maintained elsewhere.
// Systray must be ready before this method call be called.
func (a *App) Start() {
	// Setup menu icon
	systray.SetTemplateIcon(icons.TrayTemplate, icons.Tray)
	systray.SetTooltip("Cloak")

	// i18n
	translator := i18n.GetLocalizer()

	// Setup menu items
	openMenu := systray.AddMenuItem(translator.T("open"), "")
	quitMenu := systray.AddMenuItem(translator.T("quit"), "")

	go func() {
		for {
			select {
			// Realtime i18n changing
			case locale, ok := <-translator.Ch:
				if ok {
					logger.Debug().Str("locale", locale).Msg("Locale changed")
					openMenu.SetTitle(translator.T("open"))
					quitMenu.SetTitle(translator.T("quit"))

					// Persistent locale to config file
					appOptions := &Options{Locale: locale}
					cfg := ini.Empty()
					if err := cfg.ReflectFrom(appOptions); err != nil {
						logger.Error().Err(err).Msg("Failed to update app options")
						continue
					}
					if err := cfg.SaveTo(filepath.Join(a.configDir, "options.ini")); err != nil {
						logger.Error().Err(err).Msg("Failed to save app options to config file")
					}
				}
			// Menu item events
			case <-quitMenu.ClickedCh:
				systray.Quit()
			case <-openMenu.ClickedCh:
				browser.OpenURL("http://127.0.0.1:9763")
			}
		}
	}()

	// Run API server in the background
	go a.apiServer.Start("127.0.0.1:9763")

	logger.Info().Msg("App started")
}

// Stop stops the app
func (a *App) Stop() {
	a.db.Close()
	a.apiServer.Stop()
	logger.Info().Msg("App stopped")
}
