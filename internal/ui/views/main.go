package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/ui/components"
	"github.com/user/media-manager/pkg/models"
)

type MainView struct {
	widget.BaseWidget
	config             *config.Config
	database           *db.Database
	mediaGridContainer *fyne.Container
	window             fyne.Window
	mediaDir           string
	foldersTree        *widget.Tree
	filter             string
}

func (v *MainView) getChildDirs(path string) []string {
	var children []string
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Printf("[ERROR] Failed to read directory %s: %v\n", path, err)
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			children = append(children, filepath.Join(path, entry.Name()))
		}
	}
	return children
}

func (v *MainView) hasSubDirs(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return true
		}
	}
	return false
}

func (v *MainView) createFoldersTree() *widget.Tree {
	// Load all folders from DB
	folders, err := v.database.GetFolders()
	if err != nil || len(folders) == 0 {
		fmt.Printf("[ERROR] Failed to load folders from DB: %v\n", err)
		return nil
	}
	folderPaths := make([]string, len(folders))
	for i, f := range folders {
		folderPaths[i] = f.Path
	}
	v.config.MediaDirs = folderPaths
	if v.mediaDir == "" && len(folderPaths) > 0 {
		v.mediaDir = folderPaths[0]
	}

	tree := widget.NewTree(
		// Get child IDs for a given node
		func(id string) []string {
			if id == "" {
				return folderPaths
			}
			return v.getChildDirs(id)
		},
		// Check if a node has children
		func(id string) bool {
			if id == "" {
				return true
			}
			return v.hasSubDirs(id)
		},
		// Create the UI for a node
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		// Update the UI for a node
		func(id string, branch bool, node fyne.CanvasObject) {
			label := node.(*widget.Label)
			if id == "" {
				label.SetText("/")
			} else {
				label.SetText(filepath.Base(id))
			}
		},
	)

	// Handle selection
	tree.OnSelected = func(id string) {
		fmt.Printf("[DEBUG] Selected folder: %s\n", id)
		v.mediaDir = id
		v.RefreshMediaGrid()
	}

	return tree
}

func (v *MainView) filterMediaFiles(input string) {
	v.filter = input
	v.RefreshMediaGrid()
}

func (v *MainView) RefreshMediaGrid() {
	fmt.Println("[DEBUG] views/main.go: RefreshMediaGrid called.")
	if v.mediaGridContainer != nil {
		v.mediaGridContainer.Objects = []fyne.CanvasObject{}

		if len(v.config.MediaDirs) > 0 {
			mediaDir := v.config.MediaDirs[0]
			files, err := os.ReadDir(mediaDir)
			if err == nil {
				for _, file := range files {
					if !file.IsDir() {
						if v.filter != "" && !strings.Contains(strings.ToLower(file.Name()), strings.ToLower(v.filter)) {
							continue
						}
						filePath := filepath.Join(mediaDir, file.Name())
						mediaType := components.GetMediaType(file.Name())
						var thumbPath string
						// Only use thumbPath for images if needed
						card := components.NewMediaCard(filePath, file.Name(), mediaType, thumbPath)
						card.SetOnDelete(func() {
							v.mediaGridContainer.Remove(card)
							v.mediaGridContainer.Refresh()
						})
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
	cardSize := fyne.NewSize(180, 180)
	v.mediaGridContainer = container.New(layout.NewGridWrapLayout(cardSize))

	if len(v.config.MediaDirs) > 0 {
		mediaDir := v.config.MediaDirs[0]
		files, err := os.ReadDir(mediaDir)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() {
					if v.filter != "" && !strings.Contains(strings.ToLower(file.Name()), strings.ToLower(v.filter)) {
						continue
					}
					filePath := filepath.Join(mediaDir, file.Name())
					mediaType := components.GetMediaType(file.Name())
					var thumbPath string
					// Only use thumbPath for images if needed
					card := components.NewMediaCard(filePath, file.Name(), mediaType, thumbPath)
					card.SetOnDelete(func() {
						v.mediaGridContainer.Remove(card)
						v.mediaGridContainer.Refresh()
					})
					v.mediaGridContainer.Add(card)
				}
			}
		}
	}

	return v.mediaGridContainer
}

func (v *MainView) Build() fyne.CanvasObject {
	v.foldersTree = v.createFoldersTree()
	var treeScroll fyne.CanvasObject
	if v.foldersTree != nil {
		scroll := container.NewScroll(v.foldersTree)
		scroll.SetMinSize(fyne.NewSize(200, 0))
		treeScroll = scroll
	} else {
		treeScroll = container.NewVBox(widget.NewLabel("No folders found"))
	}

	mediaGrid := v.createMediaGrid()
	split := container.NewHSplit(treeScroll, mediaGrid)
	// Set offset from config
	split.SetOffset(float64(v.config.MainContentSplitOffset))
	// Note: Fyne v2 does not support OnChanged for Split. Offset persistence not supported here.

	// Toolbar: filter entry, refresh button, add folder button
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filter media...")
	filterEntry.OnChanged = func(input string) {
		v.filterMediaFiles(input)
	}
	refreshBtn := widget.NewButton("Refresh", func() {
		v.RefreshMediaGrid()
	})
	addFolderBtn := widget.NewButton("Add Folder", func() {
		dialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			folderPath := uri.Path()
			// Add to DB
			folder := &models.Folder{Path: folderPath, Name: filepath.Base(folderPath)}
			err = v.database.CreateFolder(folder)
			if err != nil {
				fmt.Printf("[ERROR] Failed to add folder to DB: %v\n", err)
				return
			}
			// Add to config and refresh sidebar
			v.config.MediaDirs = append(v.config.MediaDirs, folderPath)
			v.mediaDir = folderPath
			v.foldersTree = v.createFoldersTree()
			v.window.SetContent(v.Build())
		}, v.window)
		dialog.Show()
	})
	buttonBox := container.NewHBox(refreshBtn, addFolderBtn)
	toolbar := container.NewBorder(nil, nil, nil, buttonBox, filterEntry)

	// Pre-select the root media directory
	if v.mediaDir != "" && v.foldersTree != nil {
		v.foldersTree.Select(v.mediaDir)
	} else if v.foldersTree == nil {
		fmt.Println("[WARN] foldersTree is nil, cannot select root directory")
	}

	return container.NewVBox(toolbar, split)
}

func NewMainView(cfg *config.Config, db *db.Database, window fyne.Window, mediaDir string) *MainView {
	mv := &MainView{
		config:   cfg,
		database: db,
		window:   window,
		mediaDir: mediaDir,
	}

	// Load all folders from DB on startup
	folders, err := db.GetFolders()
	if err == nil && len(folders) > 0 {
		folderPaths := make([]string, len(folders))
		for i, f := range folders {
			folderPaths[i] = f.Path
		}
		mv.config.MediaDirs = folderPaths
		if mv.mediaDir == "" {
			mv.mediaDir = folderPaths[0]
		}
	}

	if window.Content() == nil {
		window.Resize(fyne.NewSize(1200, 800))
	}

	return mv
}
