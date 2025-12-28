package views

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	recursiveSearch    bool
	debugLog           *widget.TextGrid
	debugWriter        io.Writer
	showDebugLog       bool
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
		func(id string) []string {
			if id == "" {
				return folderPaths
			}
			return v.getChildDirs(id)
		},
		func(id string) bool {
			if id == "" {
				return true
			}
			return v.hasSubDirs(id)
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id string, branch bool, node fyne.CanvasObject) {
			label := node.(*widget.Label)
			if id == "" {
				label.SetText("/")
			} else {
				label.SetText(filepath.Base(id))
			}
		},
	)
	tree.OnSelected = func(id string) {
		fmt.Printf("[DEBUG] Selected folder: %s\n", id)
		v.mediaDir = id
		v.RefreshMediaGrid()
		tree.OpenBranch(id)
	}
	return tree
}

func (v *MainView) filterMediaFiles(input string) {
	v.filter = input
	v.RefreshMediaGrid()
}

func (v *MainView) RefreshMediaGrid() {
	v.RefreshMediaGridWithForce(false)
}

func (v *MainView) RefreshMediaGridWithForce(forceRegenerate bool) {
	if forceRegenerate {
		log.Println("[INFO] RefreshMediaGrid called with force regenerate")
	} else {
		log.Println("[DEBUG] RefreshMediaGrid called")
	}
	if v.mediaGridContainer != nil {
		v.mediaGridContainer.Objects = []fyne.CanvasObject{}
		if v.mediaDir != "" {
			mediaDir := v.mediaDir
			var files []os.DirEntry
			var err error
			var limitReached bool
			if v.recursiveSearch {
				files, limitReached = getAllMediaFilesRecursive(mediaDir)
				if limitReached {
					v.showRecursiveLimitWarning()
				}
			} else {
				files, err = os.ReadDir(mediaDir)
				if err != nil {
					files = nil
				}
			}
			for _, file := range files {
				var fileName string
				var filePath string
				if v.recursiveSearch {
					filePath = file.Name() // full path
					fileName = filepath.Base(filePath)
				} else {
					if file.IsDir() {
						continue
					}
					fileName = file.Name()
					filePath = filepath.Join(mediaDir, file.Name())
				}
				if v.filter != "" && !strings.Contains(strings.ToLower(fileName), strings.ToLower(v.filter)) {
					continue
				}

				// Look up in database to get existing media file or create a new one
				var mediaFile models.MediaFile
				if dbFile, err := v.database.GetMediaFileByPath(filePath); err == nil {
					mediaFile = *dbFile
				} else {
					// Create a minimal MediaFile struct for files not in DB yet
					mediaFile = models.MediaFile{
						Path:     filePath,
						Filename: fileName,
					}
				}

				card := components.NewMediaCard(mediaFile, v.config.ThumbnailDir)
				card.SetOnDelete(func() {
					v.mediaGridContainer.Remove(card)
					v.mediaGridContainer.Refresh()
				})
				v.mediaGridContainer.Add(card)
			}
		}
		v.mediaGridContainer.Refresh()
	}
	fmt.Println("Media grid refreshed")
}

func (v *MainView) createMediaGrid() *fyne.Container {
	const cardWidth = 216
	const cardHeight = 192
	var cards []fyne.CanvasObject
	if v.mediaDir != "" {
		mediaDir := v.mediaDir
		var files []os.DirEntry
		var err error
		var limitReached bool
		if v.recursiveSearch {
			files, limitReached = getAllMediaFilesRecursive(mediaDir)
			if limitReached {
				v.showRecursiveLimitWarning()
			}
		} else {
			files, err = os.ReadDir(mediaDir)
			if err != nil {
				files = nil
			}
		}
		for _, file := range files {
			var fileName string
			var filePath string
			if v.recursiveSearch {
				filePath = file.Name() // full path
				fileName = filepath.Base(filePath)
			} else {
				if file.IsDir() {
					continue
				}
				fileName = file.Name()
				filePath = filepath.Join(mediaDir, file.Name())
			}
			if v.filter != "" && !strings.Contains(strings.ToLower(fileName), strings.ToLower(v.filter)) {
				continue
			}

			// Look up in database to get existing media file or create a new one
			var mediaFile models.MediaFile
			if dbFile, err := v.database.GetMediaFileByPath(filePath); err == nil {
				mediaFile = *dbFile
			} else {
				// Create a minimal MediaFile struct for files not in DB yet
				mediaFile = models.MediaFile{
					Path:     filePath,
					Filename: fileName,
				}
			}

			card := components.NewMediaCard(mediaFile, v.config.ThumbnailDir)
			card.SetOnDelete(func() {
				v.RefreshMediaGrid()
			})
			cards = append(cards, card)
		}
	}
	v.mediaGridContainer = container.NewGridWrap(fyne.NewSize(cardWidth, cardHeight), cards...)
	return v.mediaGridContainer
}

const maxRecursiveFiles = 2000

func getAllMediaFilesRecursive(root string) ([]os.DirEntry, bool) {
	var files []os.DirEntry
	count := 0
	limitReached := false
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("[WARN] Error reading %s: %v\n", path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		mediaType := components.GetMediaType(path)
		if mediaType == components.MediaTypeFile {
			return nil
		}
		files = append(files, &fullPathDirEntry{DirEntry: d, fullPath: path})
		count++
		if count >= maxRecursiveFiles {
			limitReached = true
			return filepath.SkipDir
		}
		return nil
	})
	if limitReached {
		fmt.Printf("[INFO] Recursive search limit reached (%d files). Showing first %d files only.\n", count, maxRecursiveFiles)
	}
	return files, limitReached
}

type fullPathDirEntry struct {
	os.DirEntry
	fullPath string
}

func (f *fullPathDirEntry) Name() string {
	return f.fullPath
}

func (v *MainView) showRecursiveLimitWarning() {
	dialog := dialog.NewInformation(
		"Recursive Search Limit",
		fmt.Sprintf("Too many media files found (showing first %d). Please refine your search or browse smaller folders.", maxRecursiveFiles),
		v.window,
	)
	dialog.Show()
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
	split.SetOffset(float64(v.config.MainContentSplitOffset))
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filter media...")
	filterEntry.OnChanged = func(input string) {
		v.filterMediaFiles(input)
	}
	refreshBtn := widget.NewButton("Refresh", func() {
		v.RefreshMediaGrid()
	})

	forceRegenerateBtn := widget.NewButton("Force Regenerate Previews", func() {
		log.Println("[INFO] Force regenerating all previews...")
		v.RefreshMediaGridWithForce(true)
	})
	addFolderBtn := widget.NewButton("Add Folder", func() {
		dialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			folderPath := uri.Path()
			folder := &models.Folder{Path: folderPath, Name: filepath.Base(folderPath)}
			err = v.database.CreateFolder(folder)
			if err != nil {
				fmt.Printf("[ERROR] Failed to add folder to DB: %v\n", err)
				return
			}
			v.config.MediaDirs = append(v.config.MediaDirs, folderPath)
			v.mediaDir = folderPath
			v.foldersTree = v.createFoldersTree()
			v.window.SetContent(v.Build())
		}, v.window)
		dialog.Resize(fyne.NewSize(800, 600))
		dialog.Show()
	})
	recursiveCheck := widget.NewCheck("Recursive Search", func(checked bool) {
		v.recursiveSearch = checked
		fmt.Printf("[DEBUG] Recursive Search toggled: %v\n", checked)
		v.RefreshMediaGrid()
	})
	recursiveCheck.SetChecked(v.recursiveSearch)

	debugLogCheck := widget.NewCheck("Show Debug Log", func(checked bool) {
		v.showDebugLog = checked
		v.window.SetContent(v.Build())
	})
	debugLogCheck.SetChecked(v.showDebugLog)

	buttonBox := container.NewHBox(refreshBtn, forceRegenerateBtn, addFolderBtn, recursiveCheck, debugLogCheck)
	toolbar := container.NewBorder(nil, nil, nil, buttonBox, filterEntry)
	if v.mediaDir != "" && v.foldersTree != nil {
		v.foldersTree.Select(v.mediaDir)
	} else if v.foldersTree == nil {
		fmt.Println("[WARN] foldersTree is nil, cannot select root directory")
	}

	mainContent := container.NewBorder(toolbar, nil, nil, nil, split)

	// Add debug log panel if enabled
	if v.showDebugLog {
		if v.debugLog == nil {
			v.debugLog = widget.NewTextGridFromString("Debug Log:\n")
			v.debugWriter = components.NewDebugWriter(v.debugLog)
			log.SetOutput(io.MultiWriter(os.Stdout, v.debugWriter))
		}
		debugScroll := container.NewScroll(v.debugLog)
		debugScroll.SetMinSize(fyne.NewSize(0, 200))
		return container.NewBorder(nil, debugScroll, nil, nil, mainContent)
	}

	return mainContent
}

func NewMainView(cfg *config.Config, db *db.Database, window fyne.Window, mediaDir string) *MainView {
	mv := &MainView{
		config:   cfg,
		database: db,
		window:   window,
		mediaDir: mediaDir,
	}
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
