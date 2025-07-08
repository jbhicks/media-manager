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
	fyneApp  fyne.App
	window   fyne.Window
	config   *config.Config
	db       *db.Database
	mainView *views.MainView
	mediaDir string
	scanner  *scanner.MediaScanner
}

func NewMediaManagerApp(mediaDir string) (*MediaManagerApp, error) {
	fmt.Printf("[DEBUG] app.go: Received mediaDir: %s\n", mediaDir)
	cfg := config.NewConfig(mediaDir)

	database, err := db.NewDatabase(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	mediaScanner, err := scanner.NewMediaScanner(database)
	if err != nil {
		return nil, fmt.Errorf("failed to create media scanner: %w", err)
	}

	fyneApp := app.NewWithID("com.mediamanager.app")

	window := fyneApp.NewWindow("Media Manager")
	window.Resize(fyne.NewSize(1200, 800))
	window.CenterOnScreen()

	return &MediaManagerApp{
		fyneApp:  fyneApp,
		window:   window,
		config:   cfg,
		db:       database,
		mediaDir: mediaDir,
		scanner:  mediaScanner,
	}, nil
}

func (app *MediaManagerApp) Run() {
	app.setupUI()

	// Initial scan of the media directory
	fmt.Printf("[DEBUG] app.go: Starting initial scan of %s\n", app.mediaDir)
	err := app.scanner.ScanDirectory(app.mediaDir)
	if err != nil {
		fmt.Printf("Error during initial scan: %v\n", err)
	}

	// Start watching the media directory for changes
	fmt.Printf("[DEBUG] app.go: Starting file watcher for %s\n", app.mediaDir)
	err = app.scanner.StartWatching([]string{app.mediaDir})
	if err != nil {
		fmt.Printf("Error starting file watcher: %v\n", err)
	}

	// Select the initial media directory in the tree once the window is shown
	app.window.SetOnShowed(func() {
		if app.mainView.foldersTree != nil && app.mediaDir != "" {
			app.mainView.foldersTree.Select(app.mediaDir)
			app.mainView.foldersTree.Refresh()
			// Explicitly update the root node to ensure its style is applied
			app.mainView.foldersTree.UpdateNode(app.mediaDir, app.mainView.createNode, app.mainView.updateNode)
		}
	})

	app.window.ShowAndRun()
}

func (app *MediaManagerApp) setupUI() {
	mainView := views.NewMainView(app.config, app.db, app.window, app.mediaDir)
	app.mainView = mainView

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
			app.RescanMediaDirectory()
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

func (app *MediaManagerApp) RescanMediaDirectory() {
	fmt.Println("[DEBUG] app.go: Rescanning media directory...")
	// Clear existing media files from the database for the current directory
	err := app.db.DeleteMediaFilesByDirectory(app.mediaDir)
	if err != nil {
		fmt.Printf("Error clearing media files for rescan: %v\n", err)
	}

	// Re-scan the directory
	err = app.scanner.ScanDirectory(app.mediaDir)
	if err != nil {
		fmt.Printf("Error during rescan: %v\n", err)
	}
	app.mainView.RefreshMediaGrid()
	fmt.Println("[DEBUG] app.go: RescanMediaDirectory finished.")
}
