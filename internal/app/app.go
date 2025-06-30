package app

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/ui/views"
)

type MediaManagerApp struct {
	fyneApp fyne.App
	window  fyne.Window
	config  *config.Config
	db      *db.Database
}

func NewMediaManagerApp(mediaDir string) (*MediaManagerApp, error) {
	fmt.Printf("[DEBUG] app.go: Received mediaDir: %s\n", mediaDir)
	cfg := config.NewConfig(mediaDir)

	database, err := db.NewDatabase(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	fyneApp := app.NewWithID("com.mediamanager.app")

	window := fyneApp.NewWindow("Media Manager")
	window.Resize(fyne.NewSize(1200, 800))
	window.CenterOnScreen()

	return &MediaManagerApp{
		fyneApp: fyneApp,
		window:  window,
		config:  cfg,
		db:      database,
	}, nil
}

func (app *MediaManagerApp) Run() {
	app.setupUI()
	app.window.ShowAndRun()
}

func (app *MediaManagerApp) setupUI() {
	// Create main views
	mainView := views.NewMainView(app.config, app.db)

	// Create menu bar
	app.setupMenuBar()

	// Set window content
	app.window.SetContent(mainView.Build())
}

func (app *MediaManagerApp) setupMenuBar() {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Add Folder...", func() {
			// TODO: Implement folder selection dialog
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Settings...", func() {
			// TODO: Implement settings dialog
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			app.fyneApp.Quit()
		}),
	)

	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Refresh", func() {
			// TODO: Implement refresh functionality
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Small Thumbnails", nil),
		fyne.NewMenuItem("Medium Thumbnails", nil),
		fyne.NewMenuItem("Large Thumbnails", nil),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			// TODO: Implement about dialog
		}),
	)

	mainMenu := fyne.NewMainMenu(fileMenu, viewMenu, helpMenu)
	app.window.SetMainMenu(mainMenu)
}
