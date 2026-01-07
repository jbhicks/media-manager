package views

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/preview"
	"github.com/user/media-manager/internal/tagger"
	"github.com/user/media-manager/internal/ui/components"
	"github.com/user/media-manager/pkg/models"
)

// SortableMediaFile holds pre-loaded metadata for efficient sorting
type SortableMediaFile struct {
	Entry     os.DirEntry
	Info      os.FileInfo
	MediaFile *models.MediaFile
	SortKey   interface{}
}

type MainView struct {
	config             *config.Config
	database           *db.Database
	mediaGridContainer *fyne.Container
	mediaGridWrapper   *container.Scroll // Scroll wrapper for media grid
	window             fyne.Window
	mediaDir           string
	foldersTree        *widget.Tree
	filter             string
	recursiveSearch    bool
	debugLog           *widget.TextGrid
	debugWriter        io.Writer
	showDebugLog       bool
	// Sorting state
	sortBy        string
	sortAscending bool
	sortedFiles   []SortableMediaFile
	isSorting     bool
	// UI components for state persistence
	mainSplit    *container.Split // Reference to the main HSplit container
	sidebarSplit *container.Split // Reference to the sidebar HSplit container
	// Tags filtering
	selectedTags map[string]bool
	tagsTree     *widget.Tree
	tagsScroll   *container.Scroll // Scroll wrapper for tags tree
	allTags      []models.Tag
	tagger       *tagger.FilenameTagger
}

func (v *MainView) SaveConfig() {
	// Update config with current UI state
	v.config.SortBy = v.sortBy
	v.config.SortAscending = v.sortAscending
	v.config.SelectedFolder = v.mediaDir
	v.config.SelectedTags = v.selectedTags
	v.config.FilterText = v.filter
	v.config.RecursiveSearch = v.recursiveSearch
	v.config.ShowDebugLog = v.showDebugLog
	v.config.MainContentSplitOffset = v.GetMainContentSplitOffset()
	v.config.SidebarSplitOffset = v.GetSidebarSplitOffset()

	// Save to disk
	if err := config.SaveConfig(v.config); err != nil {
		fmt.Printf("[ERROR] Failed to save config: %v\n", err)
	}
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
		v.setMediaDirectory(id)
		tree.OpenBranch(id)
	}
	return tree
}

func (v *MainView) refreshTagsList() {
	// Recreate the tags tree with current directory's tags
	newTagsTree := v.createTagsList()

	// Update the scroll container's content if it exists
	if v.tagsScroll != nil {
		v.tagsScroll.Content = newTagsTree
		v.tagsScroll.Refresh()
	}
}

func (v *MainView) createTagsList() fyne.CanvasObject {
	// Collect unique tags from current directory's files
	tagMap := make(map[string]models.Tag)
	for _, sf := range v.sortedFiles {
		if sf.MediaFile != nil {
			for _, tag := range sf.MediaFile.Tags {
				tagMap[tag.Name] = tag
			}
		}
	}

	// Convert map to slice
	var tags []models.Tag
	for _, tag := range tagMap {
		tags = append(tags, tag)
	}

	v.allTags = tags

	// Group tags by category
	tagGroups := make(map[string][]models.Tag)
	for _, tag := range v.allTags {
		category := v.getTagCategory(tag.Name)
		tagGroups[category] = append(tagGroups[category], tag)
	}

	// Get sorted category names
	var categories []string
	for category := range tagGroups {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	// Create tree widget for tags
	tree := widget.NewTree(
		func(id string) []string {
			if id == "" {
				// Root level: return categories
				return categories
			}
			// Category level: return tag names
			if tags, exists := tagGroups[id]; exists {
				var tagNames []string
				for _, tag := range tags {
					tagNames = append(tagNames, tag.Name)
				}
				return tagNames
			}
			return []string{}
		},
		func(id string) bool {
			// Categories are branches, individual tags are leaves
			if id == "" {
				return true // Root is always a branch
			}
			// Check if this is a category (exists in tagGroups)
			_, isCategory := tagGroups[id]
			return isCategory
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return widget.NewLabel("Category")
			}
			return widget.NewCheck("", func(bool) {})
		},
		func(id string, branch bool, node fyne.CanvasObject) {
			if branch {
				// Category branch
				label := node.(*widget.Label)
				label.SetText(id)
			} else {
				// Individual tag leaf
				check := node.(*widget.Check)
				check.SetText(v.getTagDisplayName(id))
				check.SetChecked(v.selectedTags[id])
				check.OnChanged = func(checked bool) {
					if checked {
						v.selectedTags[id] = true
					} else {
						delete(v.selectedTags, id)
					}
					v.RefreshMediaGrid()
					v.SaveConfig() // Save selected tags
				}
			}
		},
	)

	// Open all categories by default
	for _, category := range categories {
		tree.OpenBranch(category)
	}

	v.tagsTree = tree
	return tree
}

func (v *MainView) getTagCategory(tagName string) string {
	if strings.HasPrefix(tagName, "studio:") {
		return "Studio"
	} else if strings.HasPrefix(tagName, "actress:") {
		return "Actress"
	} else if strings.HasPrefix(tagName, "date:") {
		return "Date"
	} else if strings.HasPrefix(tagName, "quality:") {
		return "Quality"
	}
	return "Other"
}

func (v *MainView) getTagDisplayName(tagName string) string {
	// Remove the prefix for display
	if strings.Contains(tagName, ":") {
		parts := strings.SplitN(tagName, ":", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return tagName
}

func (v *MainView) filterMediaFiles(input string) {
	v.filter = input
	v.RefreshMediaGrid() // Filtering is fast, no need for async
	v.SaveConfig()       // Save filter text
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

	// Get current card dimensions (these scale with zoom level)
	cardWidth := components.CardWidth()
	cardHeight := components.CardHeight()

	// Collect all cards
	var cards []fyne.CanvasObject
	if v.mediaDir != "" && len(v.sortedFiles) > 0 {
		for _, sf := range v.sortedFiles {
			var fileName string
			var filePath string
			if v.recursiveSearch {
				filePath = sf.Entry.Name() // full path for recursive
				fileName = filepath.Base(filePath)
			} else {
				fileName = sf.Entry.Name()
				filePath = filepath.Join(v.mediaDir, sf.Entry.Name())
			}

			// Apply filter
			if v.filter != "" && !strings.Contains(strings.ToLower(fileName), strings.ToLower(v.filter)) {
				continue
			}

			// Apply tag filter
			if len(v.selectedTags) > 0 && sf.MediaFile != nil {
				hasMatchingTag := false
				for _, tag := range sf.MediaFile.Tags {
					if v.selectedTags[tag.Name] {
						hasMatchingTag = true
						break
					}
				}
				if !hasMatchingTag {
					continue
				}
			}

			// Create media file struct
			var mediaFile models.MediaFile
			if sf.MediaFile != nil {
				mediaFile = *sf.MediaFile
			} else {
				// Create minimal MediaFile struct for files not in DB
				mediaFile = models.MediaFile{
					Path:     filePath,
					Filename: fileName,
				}
			}

			var card *components.MediaCard
			if forceRegenerate {
				card = components.NewMediaCardWithForce(mediaFile, v.config.ThumbnailDir, true)
			} else {
				card = components.NewMediaCard(mediaFile, v.config.ThumbnailDir)
			}
			card.SetOnDelete(func() {
				v.RefreshMediaGrid()
			})
			cards = append(cards, card)
		}
	}

	// Recreate the GridWrap container with new dimensions (for zoom support)
	v.mediaGridContainer = container.NewGridWrap(fyne.NewSize(cardWidth, cardHeight), cards...)

	// Update the scroll wrapper's content if it exists
	if v.mediaGridWrapper != nil {
		v.mediaGridWrapper.Content = v.mediaGridContainer
		v.mediaGridWrapper.Refresh()
	}

	fmt.Printf("Media grid refreshed with card size: %.0fx%.0f, %d files\n", cardWidth, cardHeight, len(cards))
}

func (v *MainView) createMediaGrid() fyne.CanvasObject {
	// Use the card dimensions from the components package for consistency
	// These are now functions that return scaled values based on zoom level
	cardWidth := components.CardWidth()
	cardHeight := components.CardHeight()
	var cards []fyne.CanvasObject
	if v.mediaDir != "" && len(v.sortedFiles) > 0 {
		for _, sf := range v.sortedFiles {
			var fileName string
			var filePath string
			if v.recursiveSearch {
				filePath = sf.Entry.Name() // full path for recursive
				fileName = filepath.Base(filePath)
			} else {
				fileName = sf.Entry.Name()
				filePath = filepath.Join(v.mediaDir, sf.Entry.Name())
			}

			// Apply filter
			if v.filter != "" && !strings.Contains(strings.ToLower(fileName), strings.ToLower(v.filter)) {
				continue
			}

			// Apply tag filter
			if len(v.selectedTags) > 0 && sf.MediaFile != nil {
				hasMatchingTag := false
				for _, tag := range sf.MediaFile.Tags {
					if v.selectedTags[tag.Name] {
						hasMatchingTag = true
						break
					}
				}
				if !hasMatchingTag {
					continue
				}
			}

			// Create media file struct
			var mediaFile models.MediaFile
			if sf.MediaFile != nil {
				mediaFile = *sf.MediaFile
			} else {
				// Create minimal MediaFile struct for files not in DB
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

	// Create a scroll container wrapper so the grid can scroll when content exceeds viewport
	// This prevents the window from expanding past screen bounds with many files
	v.mediaGridWrapper = container.NewVScroll(v.mediaGridContainer)
	return v.mediaGridWrapper
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

// loadMediaMetadataAsync loads all file metadata in background to avoid blocking UI
func (v *MainView) loadMediaMetadataAsync(mediaDir string, callback func([]SortableMediaFile)) {
	go func() {
		var sortableFiles []SortableMediaFile

		var files []os.DirEntry
		var err error
		if v.recursiveSearch {
			files, _ = getAllMediaFilesRecursive(mediaDir)
		} else {
			files, err = os.ReadDir(mediaDir)
			if err != nil {
				callback([]SortableMediaFile{})
				return
			}
		}

		for _, file := range files {
			var filePath string
			if v.recursiveSearch {
				filePath = file.Name() // full path for recursive
			} else {
				if file.IsDir() {
					continue
				}
				filePath = filepath.Join(mediaDir, file.Name())
			}

			// Load filesystem info
			info, err := file.Info()
			if err != nil {
				continue
			}

			// Load database info or create if doesn't exist
			var mediaFile *models.MediaFile
			if dbFile, err := v.database.GetMediaFileByPath(filePath); err == nil {
				mediaFile = dbFile
			} else {
				// File doesn't exist in database, create it
				width, height, duration, _ := preview.GetMetadata(filePath)
				mediaType := components.GetMediaType(filePath)
				fileTypeStr := "unknown"
				switch mediaType {
				case components.MediaTypeImage:
					fileTypeStr = "image"
				case components.MediaTypeVideo:
					fileTypeStr = "video"
				}
				newMediaFile := &models.MediaFile{
					Path:     filePath,
					Filename: file.Name(),
					Size:     info.Size(),
					ModTime:  info.ModTime(),
					FileType: fileTypeStr,
					MimeType: "", // Will be set by preview.GetMetadata if needed
					Width:    width,
					Height:   height,
					Duration: duration,
				}
				err := v.database.CreateMediaFile(newMediaFile)
				if err != nil {
					fmt.Printf("[WARN] Failed to create media file in DB %s: %v\n", filePath, err)
					continue
				}
				mediaFile = newMediaFile
			}

			// Tag the file if it doesn't have tags yet
			if len(mediaFile.Tags) == 0 {
				err := v.tagger.TagMediaFile(mediaFile)
				if err != nil {
					fmt.Printf("[WARN] Failed to tag media file %s: %v\n", filePath, err)
				} else {
					fmt.Printf("[DEBUG] Tagged media file: %s\n", filePath)
				}
			}

			sortableFiles = append(sortableFiles, SortableMediaFile{
				Entry:     file,
				Info:      info,
				MediaFile: mediaFile,
			})
		}

		callback(sortableFiles)
	}()
}

// computeSortKey calculates the sort value for a file
func (v *MainView) computeSortKey(sf *SortableMediaFile) interface{} {
	switch v.sortBy {
	case "Name":
		return strings.ToLower(sf.Entry.Name())
	case "Size":
		return sf.Info.Size()
	case "Date Modified":
		return sf.Info.ModTime()
	case "Date Created":
		if sf.MediaFile != nil {
			return sf.MediaFile.CreatedAt
		}
		return sf.Info.ModTime() // fallback
	case "Type":
		if sf.MediaFile != nil {
			return sf.MediaFile.FileType
		}
		// fallback to extension
		return strings.ToLower(filepath.Ext(sf.Entry.Name()))
	case "Duration":
		if sf.MediaFile != nil {
			return sf.MediaFile.Duration
		}
		return 0 // fallback
	case "Dimensions":
		if sf.MediaFile != nil {
			return int64(sf.MediaFile.Width) * int64(sf.MediaFile.Height)
		}
		return int64(0) // fallback
	default:
		return strings.ToLower(sf.Entry.Name())
	}
}

// sortLoadedFiles sorts pre-loaded files in memory (fast, no I/O)
func (v *MainView) sortLoadedFiles(files []SortableMediaFile) []SortableMediaFile {
	// Pre-compute sort keys
	for i := range files {
		files[i].SortKey = v.computeSortKey(&files[i])
	}

	// Sort in memory
	sort.Slice(files, func(i, j int) bool {
		less := v.compareSortKeys(files[i].SortKey, files[j].SortKey)
		if !v.sortAscending {
			less = !less
		}
		return less
	})

	return files
}

// compareSortKeys compares two sort keys of potentially different types
func (v *MainView) compareSortKeys(a, b interface{}) bool {
	switch va := a.(type) {
	case string:
		if vb, ok := b.(string); ok {
			return va < vb
		}
	case int64:
		if vb, ok := b.(int64); ok {
			return va < vb
		}
	case int:
		if vb, ok := b.(int); ok {
			return va < vb
		}
	}
	// Fallback: convert to string and compare
	return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
}

// triggerAsyncSort starts background sorting without blocking UI
func (v *MainView) triggerAsyncSort() {
	if v.isSorting || len(v.sortedFiles) == 0 {
		return // Already sorting or no data loaded
	}

	v.isSorting = true

	go func() {
		sortedFiles := v.sortLoadedFiles(v.sortedFiles)

		// Update UI on main thread
		fyne.Do(func() {
			v.sortedFiles = sortedFiles
			v.isSorting = false
			v.RefreshMediaGrid()
		})
	}()
}

// setMediaDirectory changes directory and loads metadata async
func (v *MainView) setMediaDirectory(dir string) {
	if v.mediaDir == dir {
		return
	}

	v.mediaDir = dir
	v.sortedFiles = nil // Invalidate cache
	v.SaveConfig()      // Save selected folder

	// Load metadata async
	v.loadMediaMetadataAsync(dir, func(files []SortableMediaFile) {
		fyne.Do(func() {
			v.sortedFiles = v.sortLoadedFiles(files)
			v.RefreshMediaGrid()
			// Refresh tags list in case new tags were created
			v.refreshTagsList()
		})
	})
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

	// Create tags list on the right
	tagsList := v.createTagsList()
	var tagsScroll fyne.CanvasObject
	if tagsList != nil {
		scroll := container.NewScroll(tagsList)
		scroll.SetMinSize(fyne.NewSize(200, 0))
		tagsScroll = scroll
		v.tagsScroll = scroll // Store reference for updating content
	} else {
		tagsScroll = container.NewVBox(widget.NewLabel("No tags found"))
		v.tagsScroll = nil // No scroll container when no tags
	}

	// Create nested split: folders | (media grid | tags)
	centerSplit := container.NewHSplit(mediaGrid, tagsScroll)
	centerSplit.SetOffset(float64(v.config.SidebarSplitOffset)) // Load saved offset
	v.sidebarSplit = centerSplit                                // Store reference for getting current offset

	mainSplit := container.NewHSplit(treeScroll, centerSplit)
	mainSplit.SetOffset(float64(v.config.MainContentSplitOffset))
	v.mainSplit = mainSplit // Store reference for getting current offset
	filterEntry := widget.NewEntry()
	filterEntry.SetPlaceHolder("Filter media...")
	filterEntry.SetText(v.filter) // Load saved filter text
	filterEntry.OnChanged = func(input string) {
		v.filterMediaFiles(input)
	}
	refreshBtn := widget.NewButton("Refresh", func() {
		v.RefreshMediaGrid()
	})

	forceRegenerateBtn := widget.NewButton("Force Regenerate Previews", func() {
		log.Println("[INFO] Force regenerating all previews...")
		// Run grid refresh in background to avoid blocking UI during I/O
		go func() {
			fyne.Do(func() {
				v.RefreshMediaGridWithForce(true)
			})
		}()
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
			v.setMediaDirectory(folderPath)
			v.foldersTree = v.createFoldersTree()
			v.window.SetContent(v.Build())
		}, v.window)
		dialog.Resize(fyne.NewSize(800, 600))
		dialog.Show()
	})
	recursiveCheck := widget.NewCheck("Recursive Search", func(checked bool) {
		v.recursiveSearch = checked
		fmt.Printf("[DEBUG] Recursive Search toggled: %v\n", checked)
		// Reload metadata when recursive mode changes
		v.setMediaDirectory(v.mediaDir)
		v.SaveConfig() // Save recursive search setting
	})
	recursiveCheck.SetChecked(v.recursiveSearch)

	debugLogCheck := widget.NewCheck("Show Debug Log", func(checked bool) {
		v.showDebugLog = checked
		v.window.SetContent(v.Build())
		v.SaveConfig() // Save debug log visibility
	})
	debugLogCheck.SetChecked(v.showDebugLog)

	// Sorting controls
	sortSelect := widget.NewSelect([]string{
		"Name", "Size", "Date Modified", "Date Created", "Type", "Duration", "Dimensions",
	}, func(selected string) {
		v.sortBy = selected
		v.triggerAsyncSort()
		v.SaveConfig() // Save sort criteria
	})
	sortSelect.SetSelected(v.sortBy)

	sortDirectionBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), nil)
	sortDirectionBtn.OnTapped = func() {
		v.sortAscending = !v.sortAscending
		if v.sortAscending {
			sortDirectionBtn.SetIcon(theme.MoveUpIcon())
		} else {
			sortDirectionBtn.SetIcon(theme.MoveDownIcon())
		}
		v.triggerAsyncSort()
		v.SaveConfig() // Save sort direction
	}

	sortControls := container.NewHBox(
		widget.NewLabel("Sort by:"), sortSelect, sortDirectionBtn,
	)

	clearTagsBtn := widget.NewButton("Clear Tags", func() {
		v.selectedTags = make(map[string]bool)
		if v.tagsTree != nil {
			v.tagsTree.Refresh()
		}
		v.RefreshMediaGrid()
		v.SaveConfig() // Save cleared tags
	})

	buttonBox := container.NewHBox(refreshBtn, forceRegenerateBtn, addFolderBtn, clearTagsBtn, recursiveCheck, debugLogCheck)
	toolbar := container.NewBorder(nil, nil, sortControls, buttonBox, filterEntry)
	if v.mediaDir != "" && v.foldersTree != nil {
		v.foldersTree.Select(v.mediaDir)
	} else if v.foldersTree == nil {
		fmt.Println("[WARN] foldersTree is nil, cannot select root directory")
	}

	mainContent := container.NewBorder(toolbar, nil, nil, nil, mainSplit)

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

// GetMainContentSplitOffset returns the current offset of the main content split
func (v *MainView) GetMainContentSplitOffset() float32 {
	if v.mainSplit != nil {
		return float32(v.mainSplit.Offset)
	}
	return v.config.MainContentSplitOffset // fallback to config value
}

// GetSidebarSplitOffset returns the current offset of the sidebar split (if any)
func (v *MainView) GetSidebarSplitOffset() float32 {
	if v.sidebarSplit != nil {
		return float32(v.sidebarSplit.Offset)
	}
	return v.config.SidebarSplitOffset // fallback to config value
}

// GetSortBy returns the current sort criteria
func (v *MainView) GetSortBy() string {
	return v.sortBy
}

// GetSortAscending returns the current sort direction
func (v *MainView) GetSortAscending() bool {
	return v.sortAscending
}

func NewMainView(cfg *config.Config, db *db.Database, window fyne.Window, mediaDir string) *MainView {
	mv := &MainView{
		config:          cfg,
		database:        db,
		window:          window,
		mediaDir:        mediaDir,
		sortBy:          cfg.SortBy,          // Load from config
		sortAscending:   cfg.SortAscending,   // Load from config
		selectedTags:    cfg.SelectedTags,    // Load from config
		filter:          cfg.FilterText,      // Load from config
		recursiveSearch: cfg.RecursiveSearch, // Load from config
		showDebugLog:    cfg.ShowDebugLog,    // Load from config
		tagger:          tagger.NewFilenameTagger(db),
	}
	// Initialize selectedTags map if nil
	if mv.selectedTags == nil {
		mv.selectedTags = make(map[string]bool)
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
	// Load initial tags
	mv.refreshTagsList()
	// Load initial metadata async
	if mv.mediaDir != "" {
		mv.loadMediaMetadataAsync(mv.mediaDir, func(files []SortableMediaFile) {
			fyne.Do(func() {
				mv.sortedFiles = mv.sortLoadedFiles(files)
				// Don't call RefreshMediaGrid here as Build() will be called next
				// Refresh tags list in case new tags were created
				mv.refreshTagsList()
			})
		})
	}
	if window.Content() == nil {
		window.Resize(fyne.NewSize(1200, 800))
	}
	return mv
}
