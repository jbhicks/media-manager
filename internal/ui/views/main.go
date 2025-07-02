package views

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/ui/components"
	"github.com/user/media-manager/pkg/models"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return true
	default:
		return false
	}
}

func isVideoFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".mp4", ".webm", ".ogv", ".flv", ".mov", ".avi", ".mkv":
		return true
	default:
		return false
	}
}

func ensureVideoThumbnail(videoPath, thumbPath string) {
	fmt.Printf("[DEBUG] Checking for video thumbnail: %s\n", thumbPath)
	if _, err := os.Stat(thumbPath); err == nil {
		fmt.Printf("[DEBUG] Thumbnail already exists: %s\n", thumbPath)
		return // thumbnail exists
	}
	fmt.Printf("[DEBUG] Thumbnail not found, generating for: %s\n", videoPath)
	_ = os.MkdirAll(filepath.Dir(thumbPath), 0755)

	// Use 200px wide thumbnail generation (no padding)
	cmd := []string{"ffmpeg", "-y", "-i", videoPath, "-ss", "00:00:01.000", "-vframes", "1",
		"-vf", "scale=200:200:force_original_aspect_ratio=increase,crop=200:200",
		thumbPath}
	fmt.Printf("[DEBUG] Generating 200px wide video thumbnail: %v\n", cmd)
	err := runCommand(cmd)
	if err != nil {
		fmt.Printf("[DEBUG] ffmpeg error: %v\n", err)
	} else {
		fmt.Printf("[DEBUG] ffmpeg thumbnail generated: %s\n", thumbPath)
	}
}

func runCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}
	return exec.Command(args[0], args[1:]...).Run()
}

type MainView struct {
	config             *config.Config
	database           *db.Database
	mediaGridContainer *fyne.Container
	window             fyne.Window
}

func (v *MainView) GetMediaGridContainer() *fyne.Container {
	return v.mediaGridContainer
}

func NewMainView(cfg *config.Config, database *db.Database, win fyne.Window) *MainView {
	fmt.Printf("[DEBUG] main.go: MainView using MediaDirs: %v\n", cfg.MediaDirs)
	return &MainView{
		config:   cfg,
		database: database,
		window:   win,
	}
}

func (v *MainView) Build() fyne.CanvasObject {
	toolbar := v.createToolbar()
	sidebar := v.createSidebar()
	statusBar := v.createStatusBar()
	grid := v.createMediaGrid()

	mainContent := container.NewHSplit(sidebar, grid)
	mainContent.SetOffset(0.25)

	content := container.NewBorder(
		toolbar,     // top
		statusBar,   // bottom
		nil,         // left
		nil,         // right
		mainContent, // center
	)
	return content
}
func (v *MainView) createToolbar() fyne.CanvasObject {
	searchEntry := widget.NewEntry()
	searchEntry.OnChanged = func(input string) {
		v.filterMediaFiles(input)
	}

	refreshBtn := widget.NewButton("Refresh", func() {
		v.RefreshMediaGrid()
	})

	addFolderBtn := widget.NewButton("Add Folder", func() {
		v.database.CreateFolder(&models.Folder{Path: "/path/to/new/folder", Name: "New Folder"})
	})

	searchEntry.SetPlaceHolder("Filter...")
	searchEntry.Wrapping = fyne.TextTruncate
	searchEntry.SetPlaceHolder("Filter...")
	searchEntry.Wrapping = fyne.TextTruncate
	return container.NewBorder(nil, nil, nil,
		container.NewHBox(refreshBtn, addFolderBtn),
		searchEntry,
	)
}

func (v *MainView) createSidebar() fyne.CanvasObject {
	// Folders section
	foldersLabel := widget.NewRichTextFromMarkdown("**Folders**")

	current := ""
	root := ""
	if len(v.config.MediaDirs) > 0 {
		root = v.config.MediaDirs[0]
		current = root
	}
	// Build folder tree
	getChildren := func(uid string) []string {
		entries, err := os.ReadDir(uid)
		if err != nil {
			return nil
		}
		var children []string
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				children = append(children, filepath.Join(uid, entry.Name()))
			}
		}
		return children
	}
	createNode := func(branch bool) fyne.CanvasObject {
		return widget.NewLabel("")
	}
	updateNode := func(uid string, branch bool, obj fyne.CanvasObject) {
		label := obj.(*widget.Label)
		label.SetText(filepath.Base(uid))
		if uid == current {
			label.TextStyle = fyne.TextStyle{Bold: true}
		} else {
			label.TextStyle = fyne.TextStyle{}
		}
	}
	foldersTree := widget.NewTree(
		func(uid string) []string { return getChildren(uid) },
		func(uid string) bool { return len(getChildren(uid)) > 0 },
		createNode,
		updateNode,
	)
	foldersTree.Root = root
	foldersTree.OnSelected = func(uid string) {
		v.config.MediaDirs = []string{uid}
		v.RefreshMediaGrid()
	}
	foldersScroll := container.NewVScroll(foldersTree)
	foldersScroll.SetMinSize(fyne.NewSize(0, 120))

	// Tags section
	tagsLabel := widget.NewRichTextFromMarkdown("**Tags**")
	tagsList := widget.NewList(
		func() int { return 0 }, // TODO: return actual tag count
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			// TODO: Update tag item
		},
	)

	tagsSection := container.NewVBox(tagsLabel, tagsList)
	split := container.NewVSplit(
		container.NewBorder(foldersLabel, nil, nil, nil, container.NewVScroll(foldersTree)),
		tagsSection,
	)
	split.SetOffset(0.95)
	return split

}

func (v *MainView) filterMediaFiles(input string) {
	// Logic to filter media files
	fmt.Printf("Filtering media files with input: %s\n", input)
}

func (v *MainView) RefreshMediaGrid() {
	fmt.Println("[DEBUG] views/main.go: RefreshMediaGrid called.")
	if v.mediaGridContainer != nil {
		newGrid := v.createMediaGrid()
		v.mediaGridContainer.Objects = newGrid.Objects
		v.mediaGridContainer.Refresh()
	}
	fmt.Println("Media grid refreshed")
}

func (v *MainView) createMediaGrid() *fyne.Container {
	cardSize := fyne.NewSize(192, 192)
	var cards []fyne.CanvasObject
	if len(v.config.MediaDirs) > 0 {
		mediaDir := v.config.MediaDirs[0]
		files, err := os.ReadDir(mediaDir)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(mediaDir, file.Name())
					mediaType := components.GetMediaType(file.Name())
					card := components.NewMediaCard(filePath, file.Name(), mediaType)
					cards = append(cards, card)
				}
			}
		}
	}
	return container.New(layout.NewGridWrapLayout(cardSize), cards...)
}
			}
		}
	}
	return container.New(layout.NewGridWrapLayout(cardSize), cards...)

	cardSize := fyne.NewSize(192, 192)
	grid := container.New(layout.NewGridWrapLayout(cardSize))
	if len(v.config.MediaDirs) > 0 {
		mediaDir := v.config.MediaDirs[0]
		files, err := os.ReadDir(mediaDir)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(mediaDir, file.Name())
					mediaType := components.GetMediaType(file.Name())
					card := components.NewMediaCard(filePath, file.Name(), mediaType)
					grid.Add(card)
				}
			}
		}
	}
	return grid

	// Use GridWrap with fixed cell size instead of GridWithColumns
	cardSize := fyne.NewSize(192, 192) // Modern card size
grid := container.New(layout.NewGridWrapLayout(cardSize))
gridBox := container.New(layout.NewCenterLayout(), grid)
padded := container.NewPadded(gridBox)
return padded	if len(v.config.MediaDirs) > 0 {
		mediaDir := v.config.MediaDirs[0]
		fmt.Printf("[DEBUG] main.go: Loading media from: %s\n", mediaDir)
		files, err := os.ReadDir(mediaDir)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(mediaDir, file.Name())
					mediaType := components.GetMediaType(file.Name())

					// Create uniform media card
					card := components.NewMediaCard(filePath, file.Name(), mediaType)
					grid.Add(card)
				}
			}
		} else {
			fmt.Printf("[DEBUG] main.go: Failed to read media dir: %v\n", err)
		}
	} else {
		fmt.Printf("[DEBUG] main.go: No media directory set\n")
	}

	return grid
}
func (v *MainView) createStatusBar() fyne.CanvasObject {
	statusLabel := widget.NewLabel("Ready")
	fileCountLabel := widget.NewLabel("0 files")

	return container.NewHBox(
		statusLabel,
		widget.NewSeparator(),
		fileCountLabel,
	)
}
