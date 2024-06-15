package main

import (
	"Cloak/config"
	"Cloak/extension"
	"Cloak/i18n"
	"Cloak/icons"
	"Cloak/models"
	"Cloak/server"
	"database/sql"
	"fyne.io/systray"
	"github.com/pkg/browser"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"

	"github.com/lopezator/migrator"
	_ "github.com/mattn/go-sqlite3"
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
	config      *config.Configurator
	configCh    chan map[string]string
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

func (a *App) loadConfig() {
	var err error
	a.config, err = config.NewConfigurator(filepath.Join(a.configDir, "options.ini"))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize application configurator")
	}

	a.config.SetCallbacks(map[string]config.Callback{
		"locale": func(v string) error {
			if err := i18n.GetLocalizer().SetLocale(v); err != nil {
				logger.Error().Err(err).
					Str("locale", v).
					Msg("Failed to set locale loaded from config file")
				return err
			}
			return nil
		},
		"loglevel": func(v string) error {
			level, err := zerolog.ParseLevel(strings.ToLower(v))
			if err != nil {
				logger.Warn().Err(err).Str("loglevel", v).Msg("Failed to parse log level")
				return err
			}
			zerolog.SetGlobalLevel(level)
			logger.Debug().Interface("Level", level).Msg("Log level changed")
			return nil
		},
	})
	a.config.Load()
}

// NewApp constructs and returns a new App instance
func NewApp() *App {
	app := &App{
		releaseMode: extension.ReleaseMode == "true",
		configCh:    make(chan map[string]string, 10), // TODO How big should the buffer be?
	}

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

	app.apiServer = server.NewApiServer(app.repo, app.releaseMode, app.configCh)

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
			// Someone requested to change config, so we should notify the configurator
			case kv, ok := <-a.configCh:
				if ok {
					for k, v := range kv {
						a.config.Set(k, v)
					}
				}
			// Locale actually changed, so we're okay to load new localed strings from translator
			case locale, ok := <-translator.Ch:
				if ok {
					logger.Debug().Str("locale", locale).Msg("Locale changed")
					openMenu.SetTitle(translator.T("open"))
					quitMenu.SetTitle(translator.T("quit"))
				}
			// Menu item events
			case <-quitMenu.ClickedCh:
				systray.Quit()
			case <-openMenu.ClickedCh:
				browser.OpenURL(a.apiServer.GetAccessUrl())
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
	logger.Info().Msg("App stopped")
}
