package app

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/preview"
	"github.com/user/media-manager/internal/scanner"
	"github.com/user/media-manager/internal/ui/views"
	"github.com/user/media-manager/pkg/models"
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
	// Check CLEAR_DB_ON_START env var
	clearDB := os.Getenv("CLEAR_DB_ON_START") == "true"

	cfg, err := config.LoadConfig(mediaDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var database *db.Database
	if clearDB {
		fmt.Println("[DEBUG] CLEAR_DB_ON_START is true: will clear previews after init.")
		// Preview clearing is handled in main.go before app startup.
	}

	// Always re-initialize database for app usage
	database, err = db.NewDatabase(cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	fmt.Printf("[DEBUG] app.go: Received mediaDir: %s\n", mediaDir)

	mediaScanner, err := scanner.NewMediaScanner(database)
	if err != nil {
		return nil, fmt.Errorf("failed to create media scanner: %w", err)
	}

	fyneApp := app.NewWithID("com.mediamanager.app")

	window := fyneApp.NewWindow("Media Manager")

	// Load window size and position from config, or use defaults
	if cfg.WindowWidth > 0 && cfg.WindowHeight > 0 {
		window.Resize(fyne.NewSize(cfg.WindowWidth, cfg.WindowHeight))

		fmt.Printf("[DEBUG] app.go: Loaded window size: %fx%f at %f,%f\n", cfg.WindowWidth, cfg.WindowHeight, cfg.WindowX, cfg.WindowY)
	} else {
		window.Resize(fyne.NewSize(1200, 800))
		window.CenterOnScreen()
		fmt.Println("[DEBUG] app.go: Using default window size and centering on screen.")
	}

	// Ensure the passed-in mediaDir is in the DB as a Folder
	if mediaDir != "" {
		var count int64
		database.GetDB().Model(&models.Folder{}).Where("path = ?", mediaDir).Count(&count)
		if count == 0 {
			folder := &models.Folder{Path: mediaDir, Name: filepath.Base(mediaDir)}
			database.GetDB().Create(folder)
		}
	}

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

	// Rebuild missing animated previews for videos
	app.RebuildMissingPreviews()

	// Start watching the media directory for changes
	fmt.Printf("[DEBUG] app.go: Starting file watcher for %s\n", app.mediaDir)
	err = app.scanner.StartWatching([]string{app.mediaDir})
	if err != nil {
		fmt.Printf("Error starting file watcher: %v\n", err)
	}

	app.window.ShowAndRun()
}

// RebuildMissingPreviews regenerates animated previews for videos with empty PreviewPath
func (app *MediaManagerApp) RebuildMissingPreviews() {
	fmt.Println("[DEBUG] Rebuilding missing animated previews...")
	var videos []models.MediaFile
	db := app.db.GetDB()
	db.Where("file_type = ? AND (preview_path = '' OR preview_path IS NULL)", "video").Find(&videos)
	fmt.Printf("[DEBUG] Found %d videos with missing previews.\n", len(videos))
	if len(videos) == 0 {
		fmt.Println("[DEBUG] No missing previews to rebuild.")
		return
	}
	for _, video := range videos {
		gifPath := filepath.Join(app.config.ThumbnailDir, fmt.Sprintf("%d.gif", video.ID))
		err := preview.GenerateAnimatedPreview(video.Path, gifPath)
		if err != nil {
			fmt.Printf("[ERROR] Failed to generate preview for %s: %v\n", video.Path, err)
			continue
		}
		video.PreviewPath = gifPath
		db.Model(&video).Update("preview_path", gifPath)
		fmt.Printf("[DEBUG] Rebuilt preview for %s -> %s\n", video.Path, gifPath)
	}
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

func (app *MediaManagerApp) SaveConfig() {
	fmt.Println("[DEBUG] app.go: Saving configuration...")
	if app.mainView != nil {
		// app.config.MainContentSplitOffset = app.mainView.GetMainContentSplitOffset() // Disabled: not implemented
		// app.config.SidebarSplitOffset = app.mainView.GetSidebarSplitOffset() // Disabled: not implemented
		fmt.Printf("[DEBUG] app.go: Retrieved MainContentSplitOffset: %f, SidebarSplitOffset: %f\n", app.config.MainContentSplitOffset, app.config.SidebarSplitOffset)
	}

	// Save window size and position
	app.config.WindowWidth = app.window.Canvas().Size().Width
	app.config.WindowHeight = app.window.Canvas().Size().Height

	fmt.Printf("[DEBUG] app.go: Saving window size: %fx%f at %f,%f\n", app.config.WindowWidth, app.config.WindowHeight, app.config.WindowX, app.config.WindowY)

	err := config.SaveConfig(app.config)
	if err != nil {
		fmt.Printf("[DEBUG] app.go: Failed to save config: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] app.go: Config saved: %+v\n", app.config)
	}
}
