package views

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/ui/components"
	"github.com/user/media-manager/pkg/models"
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
	cmd := []string{
		"ffmpeg", "-y", "-i", videoPath, "-ss", "00:00:01.000", "-vframes", "1",
		"-vf", "scale=200:200:force_original_aspect_ratio=increase,crop=200:200",
		thumbPath,
	}
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
	mediaDir           string
	foldersTree        *widget.Tree
}

func (v *MainView) GetMediaGridContainer() *fyne.Container {
	return v.mediaGridContainer
}

func NewMainView(cfg *config.Config, database *db.Database, win fyne.Window, mediaDir string) *MainView {
	fmt.Printf("[DEBUG] main.go: MainView using MediaDirs: %v\n", cfg.MediaDirs)
	return &MainView{
		config:   cfg,
		database: database,
		window:   win,
		mediaDir: filepath.Clean(mediaDir),
	}
}


func (v *MainView) Build() fyne.CanvasObject {
	toolbar := v.createToolbar()
	sidebar := v.createSidebar()
	statusBar := v.createStatusBar()
	grid := v.createMediaGrid()

	v.mainContentSplit = container.NewHSplit(sidebar, grid)
	v.mainContentSplit.SetOffset(float64(v.config.MainContentSplitOffset))
	// v.mainContentSplit.OnChanged = func(f float32) {
	// 	v.config.MainContentSplitOffset = f
	// }

	content := container.NewBorder(
		toolbar,            // top
		statusBar,          // bottom
		nil,                // left
		nil,                // right
		v.mainContentSplit, // center
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
		fmt.Println("[DEBUG] Add Folder button clicked.")
		picker := dialogs.NewFolderPickerDialog(func(selectedPath string) {
			fmt.Printf("[DEBUG] Selected folder path from custom dialog: %s\n", selectedPath)
			// Add to config.MediaDirs if not already present
			found := false
			for _, dir := range v.config.MediaDirs {
				if dir == selectedPath {
					found = true
					break
				}
			}
			if !found {
				v.config.MediaDirs = append(v.config.MediaDirs, selectedPath)
				fmt.Printf("[DEBUG] Added %s to MediaDirs. New MediaDirs: %v\n", selectedPath, v.config.MediaDirs)
			} else {
				fmt.Printf("[DEBUG] %s already in MediaDirs.\n", selectedPath)
			}
			// Refresh the media grid with the new directory
			fmt.Println("[DEBUG] Refreshing media grid...")
			v.RefreshMediaGrid()

			// Update the sidebar tree to show and select the new folder
			fmt.Printf("[DEBUG] Updating sidebar tree with root: %s\n", selectedPath)
			v.foldersTree.Root = selectedPath
			v.foldersTree.Select(selectedPath)
			v.foldersTree.Refresh()
			fmt.Println("[DEBUG] Sidebar tree refreshed.")
		}, v.window)
		picker.Show()
	})

	searchEntry.SetPlaceHolder("Filter...")
	return container.NewBorder(nil, nil, nil,
		container.NewHBox(refreshBtn, addFolderBtn),
		searchEntry,
	)
}

func (v *MainView) createSidebar() fyne.CanvasObject {
	// Folders section
	foldersLabel := widget.NewRichTextFromMarkdown("**Folders**")

	root := v.mediaDir

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
	func (v *MainView) createNode(branch bool) fyne.CanvasObject {
	return widget.NewLabel("")
}

func (v *MainView) updateNode(uid string, branch bool, obj fyne.CanvasObject) {
	label := obj.(*widget.Label)
	label.SetText(filepath.Base(uid))
	if uid == v.mediaDir {
		label.TextStyle = fyne.TextStyle{Bold: true}
	} else {
		label.TextStyle = fyne.TextStyle{}
	}
}
	foldersTree := widget.NewTree(
		func(uid string) []string { return getChildren(uid) },
		func(uid string) bool { return len(getChildren(uid)) > 0 },
		v.createNode,
		v.updateNode,
	)
	v.foldersTree = foldersTree // Assign to the struct field
	foldersTree.Root = root
	foldersTree.Select(root)
	foldersTree.OnSelected = func(uid string) {
		v.mediaDir = filepath.Clean(uid)
		v.RefreshMediaGrid()
	}
	foldersScroll := container.NewVScroll(foldersTree)
	foldersScroll.SetMinSize(fyne.NewSize(0, 120))
}

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
	v.sidebarSplit = container.NewVSplit(
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
		// First clear all existing cards
		v.mediaGridContainer.Objects = []fyne.CanvasObject{}

		// Then add new cards
		if len(v.config.MediaDirs) > 0 {
			mediaDir := v.config.MediaDirs[0]
			files, err := os.ReadDir(mediaDir)
			if err == nil {
				for _, file := range files {
					if !file.IsDir() {
						filePath := filepath.Join(mediaDir, file.Name())
						mediaType := components.GetMediaType(file.Name())
						card := components.NewMediaCard(filePath, file.Name(), mediaType)
						v.mediaGridContainer.Add(card)
					}
				}
			}
		}

		v.mediaGridContainer.Refresh()
	}
	fmt.Println("Media grid refreshed")
}

func (v *MainView) createMediaGrid() *fyne.Container {
	cardSize := fyne.NewSize(192, 192)
	var cards []fyne.CanvasObject

	// Fetch media files from the database
	mediaFiles, err := v.database.GetMediaFiles(-1, -1) // Fetch all for now
	if err != nil {
		fmt.Printf("Error fetching media files: %v\n", err)
		return container.New(layout.NewGridWrapLayout(cardSize))
	}

	for _, mediaFile := range mediaFiles {
		card := components.NewMediaCard(mediaFile.Path, mediaFile.Filename, components.GetMediaType(mediaFile.Filename))
		cards = append(cards, card)
	}

	return container.New(layout.NewGridWrapLayout(cardSize), cards...)
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
