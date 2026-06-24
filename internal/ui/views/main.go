package views

import (
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/preview"
	"github.com/user/media-manager/internal/tagger"
	"github.com/user/media-manager/internal/ui/components"
	"github.com/user/media-manager/pkg/models"
)

// FolderCache holds cached data for a folder to avoid expensive reloads
type FolderCache struct {
	Files         []SortableMediaFile
	ModTime       time.Time
	RecursiveMode bool
}

// SortableMediaFile holds pre-loaded metadata for efficient sorting
type SortableMediaFile struct {
	Entry     os.DirEntry
	Info      os.FileInfo
	MediaFile *models.MediaFile
	SortKey   interface{}
}

type MainView struct {
	config              *config.Config
	database            *db.Database
	mediaGridContainer  *fyne.Container
	mediaGridWrapper    *container.Scroll // Scroll wrapper for media grid
	loadingContainer    *fyne.Container
	loadingLabel        *widget.Label
	window              fyne.Window
	mediaDir            string
	launchFocusFiles    map[string]bool
	launchFocusActive   bool
	launchFocusPrompted bool
	foldersTree         *widget.Tree
	filter              string
	recursiveSearch     bool
	isLoadingMedia      bool
	loadingMessage      string
	debugLog            *widget.TextGrid
	debugWriter         io.Writer
	showDebugLog        bool
	// Sorting state
	sortBy        string
	sortAscending bool
	sortedFiles   []SortableMediaFile
	isSorting     bool
	// UI components for state persistence
	mainSplit    *container.Split  // Reference to the main HSplit container
	sidebarSplit *container.Split  // Reference to the sidebar HSplit container
	treeScroll   *container.Scroll // Reference to the folders tree scroll container
	// Tags filtering
	selectedTags map[string]bool
	tagsTree     *widget.Tree
	tagsScroll   *container.Scroll // Scroll wrapper for tags tree
	allTags      []models.Tag
	tagger       *tagger.FilenameTagger
	// Folder caching to avoid expensive reloads
	folderCache map[string]*FolderCache
	// Downloads tab
	downloadsView   *DownloadsView
	downloadManager DownloadManager
	// Drag-and-drop state
	draggedFilePathVar string
	dragMu             sync.Mutex
}

// DownloadManager is the subset of the service DownloadManager used by the UI.
type DownloadManager interface {
	AddSearchResult(result *models.SearchResult) (*models.DownloadTask, error)
}

func (v *MainView) effectiveFilter() string {
	return v.filter
}

func (v *MainView) isLaunchFocusedFile(fileName string) bool {
	if !v.launchFocusActive || len(v.launchFocusFiles) == 0 {
		return true
	}
	_, ok := v.launchFocusFiles[strings.ToLower(fileName)]
	return ok
}

func (v *MainView) clearLaunchFocus() {
	if !v.launchFocusActive {
		return
	}
	v.launchFocusActive = false
	v.RefreshMediaGrid()
}

func (v *MainView) maybeShowLaunchFocusPrompt() {
	if v.window == nil || !v.launchFocusActive || v.launchFocusPrompted {
		return
	}
	v.launchFocusPrompted = true

	message := "Opened from Explorer. The grid is focused to the selected files."
	if len(v.launchFocusFiles) == 1 {
		for name := range v.launchFocusFiles {
			message = fmt.Sprintf("Opened from Explorer: %s", name)
		}
	}

	prompt := dialog.NewCustomConfirm(
		"Opened from Explorer",
		"Clear Focus",
		"Keep",
		widget.NewLabel(message),
		func(clear bool) {
			if clear {
				v.clearLaunchFocus()
			}
		},
		v.window,
	)
	prompt.Show()
}

func (v *MainView) scrollToCardIndex(index int, cardHeight float32) {
	if v.mediaGridWrapper == nil || v.mediaGridContainer == nil {
		return
	}

	viewHeight := v.mediaGridWrapper.Size().Height
	if viewHeight <= 0 {
		return
	}

	cardWidth := components.CardWidth()
	gridWidth := v.mediaGridContainer.Size().Width
	if gridWidth <= 0 && v.window != nil && v.window.Canvas() != nil {
		gridWidth = v.window.Canvas().Size().Width
	}

	columns := int(gridWidth / cardWidth)
	if columns < 1 {
		columns = 1
	}

	row := index / columns
	targetY := float32(row) * cardHeight
	if viewHeight > cardHeight {
		targetY -= (viewHeight - cardHeight) * 0.5
	}
	if targetY < 0 {
		targetY = 0
	}

	maxY := v.mediaGridContainer.MinSize().Height - viewHeight
	if maxY < 0 {
		maxY = 0
	}
	if targetY > maxY {
		targetY = maxY
	}

	v.mediaGridWrapper.Offset = fyne.NewPos(0, targetY)
	v.mediaGridWrapper.Refresh()
}

func (v *MainView) SaveConfig() {
	if v.database != nil {
		if folders, err := v.database.GetFolders(); err == nil {
			folderPaths := make([]string, len(folders))
			for i, f := range folders {
				folderPaths[i] = normalizeTreePath(f.Path)
			}
			v.config.MediaDirs = folderPaths
		}
	}

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
	v.config.OpenBranches = normalizeOpenBranches(v.config.OpenBranches, v.config.MediaDirs)

	// Also save current window size (frequent saves ensure size persists during development)
	if v.window != nil && v.window.Canvas() != nil {
		size := v.window.Canvas().Size()
		v.config.WindowWidth = size.Width
		v.config.WindowHeight = size.Height
		fmt.Printf("[DEBUG] MainView.SaveConfig: Saved window size: %fx%f\n", size.Width, size.Height)
	}

	// Save to disk
	if err := config.SaveConfig(v.config); err != nil {
		fmt.Printf("[ERROR] Failed to save config: %v\n", err)
	}
}

func normalizeTreePath(path string) string {
	if path == "" {
		return ""
	}

	cleaned := filepath.Clean(filepath.FromSlash(path))
	if volume := filepath.VolumeName(cleaned); volume != "" {
		cleaned = strings.ToUpper(volume) + cleaned[len(volume):]
	}
	return cleaned
}

func treePathKey(path string) string {
	return strings.ToLower(normalizeTreePath(path))
}

func isUnderRoot(path string, rootKeys map[string]string) bool {
	pathKey := treePathKey(path)
	for rootKey := range rootKeys {
		if pathKey == rootKey || strings.HasPrefix(pathKey, rootKey+`\`) {
			return true
		}
	}
	return false
}

func normalizeOpenBranches(openBranches map[string]bool, roots []string) map[string]bool {
	if len(openBranches) == 0 {
		return make(map[string]bool)
	}

	rootKeys := make(map[string]string, len(roots))
	for _, root := range roots {
		normalizedRoot := normalizeTreePath(root)
		if normalizedRoot == "" {
			continue
		}
		rootKeys[treePathKey(normalizedRoot)] = normalizedRoot
	}

	normalized := make(map[string]bool, len(openBranches))
	for branch := range openBranches {
		normalizedBranch := normalizeTreePath(branch)
		if normalizedBranch == "" || !isUnderRoot(normalizedBranch, rootKeys) {
			continue
		}
		if _, err := os.Stat(normalizedBranch); err != nil {
			continue
		}
		if root, ok := rootKeys[treePathKey(normalizedBranch)]; ok {
			normalized[root] = true
			continue
		}
		normalized[normalizedBranch] = true
	}

	return normalized
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

// folderNode is a tree node that can accept dropped media cards.
type folderNode struct {
	widget.Label
	id   string
	view *MainView
}

func newFolderNode(view *MainView) *folderNode {
	n := &folderNode{view: view}
	n.ExtendBaseWidget(n)
	return n
}

var _ desktop.Hoverable = (*folderNode)(nil)

// MouseIn handles the cursor entering the folder node. If a media card is
// currently being dragged, treat this as a drop onto the folder.
func (n *folderNode) MouseIn(e *desktop.MouseEvent) {
	v := n.view
	if v == nil || n.id == "" {
		return
	}

	filePath := v.draggedFilePath()
	if filePath == "" || n.id == filePath {
		return
	}

	log.Printf("[DRAG] Dropped %s onto folder %s", filePath, n.id)
	go v.moveFileToFolder(filePath, n.id)
	v.clearDrag()
}

func (n *folderNode) MouseOut()                      {}
func (n *folderNode) MouseMoved(*desktop.MouseEvent) {}

func (v *MainView) createFoldersTree() *widget.Tree {
	folders, err := v.database.GetFolders()
	if err != nil || len(folders) == 0 {
		fmt.Printf("[ERROR] Failed to load folders from DB: %v\n", err)
		return nil
	}
	folderPaths := make([]string, len(folders))
	for i, f := range folders {
		folderPaths[i] = normalizeTreePath(f.Path)
	}
	v.config.MediaDirs = folderPaths
	v.config.OpenBranches = normalizeOpenBranches(v.config.OpenBranches, folderPaths)
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
			return newFolderNode(v)
		},
		func(id string, branch bool, node fyne.CanvasObject) {
			folder := node.(*folderNode)
			folder.id = id
			if id == "" {
				folder.SetText("/")
			} else {
				folder.SetText(filepath.Base(id))
			}
		},
	)
	tree.OnSelected = func(id string) {
		fmt.Printf("[DEBUG] Selected folder: %s\n", id)
		v.setMediaDirectory(id)
		tree.OpenBranch(id)
	}
	tree.OnBranchOpened = func(id string) {
		v.config.OpenBranches[normalizeTreePath(id)] = true
		v.SaveConfig()
	}
	tree.OnBranchClosed = func(id string) {
		delete(v.config.OpenBranches, normalizeTreePath(id))
		v.SaveConfig()
	}
	// Open branches from saved state
	for branch := range v.config.OpenBranches {
		tree.OpenBranch(branch)
	}
	// Create transparent overlay to catch right-clicks and show context menu
	overlay := &treeContextCatcher{tree: tree, view: v}
	overlay.ExtendBaseWidget(overlay)
	// Return a Max container so overlay sits on top of the tree and receives secondary taps
	// We still want callers to receive the underlying tree object; store reference
	v.foldersTree = tree
	// For callers expecting *widget.Tree, we keep returning the tree while the UI uses the container
	// The caller that builds the layout will use v.createFoldersTree() and place the tree in a scroll.
	// To actually use the overlay we replace the scroll's content where needed in buildMainContent.
	return tree
}

func (v *MainView) setDraggedFilePath(path string) {
	v.dragMu.Lock()
	defer v.dragMu.Unlock()
	v.draggedFilePathVar = path
}

func (v *MainView) draggedFilePath() string {
	v.dragMu.Lock()
	defer v.dragMu.Unlock()
	return v.draggedFilePathVar
}

func (v *MainView) clearDrag() {
	v.dragMu.Lock()
	defer v.dragMu.Unlock()
	v.draggedFilePathVar = ""
}

// moveFileToFolder moves a media file on disk and updates the database/UI.
func (v *MainView) moveFileToFolder(filePath, destDir string) {
	srcPath := normalizeTreePath(filePath)
	destPath := normalizeTreePath(filepath.Join(destDir, filepath.Base(srcPath)))

	if srcPath == destPath {
		return
	}

	// Verify source is a file
	info, err := os.Stat(srcPath)
	if err != nil || info.IsDir() {
		log.Printf("[DRAG] Source file not found or is a directory: %s", srcPath)
		return
	}

	// Verify destination is a directory
	destInfo, err := os.Stat(destDir)
	if err != nil || !destInfo.IsDir() {
		log.Printf("[DRAG] Destination is not a directory: %s", destDir)
		return
	}

	// Move file
	if err := os.Rename(srcPath, destPath); err != nil {
		log.Printf("[DRAG] Failed to move %s to %s: %v", srcPath, destPath, err)
		return
	}

	// Update database
	if err := v.database.UpdateMediaFilePath(srcPath, destPath); err != nil {
		log.Printf("[DRAG] Failed to update database path %s -> %s: %v", srcPath, destPath, err)
	}

	// Remove from current sorted files if visible
	fyne.Do(func() {
		found := false
		for i, sf := range v.sortedFiles {
			sfPath := sf.Entry.Name()
			if v.recursiveSearch {
				sfPath = normalizeTreePath(sfPath)
			} else {
				sfPath = normalizeTreePath(filepath.Join(v.mediaDir, sfPath))
			}
			if sfPath == srcPath {
				v.sortedFiles = append(v.sortedFiles[:i], v.sortedFiles[i+1:]...)
				found = true
				break
			}
		}
		if found {
			v.RefreshMediaGrid()
		}
		// Ensure destination folder is visible/selected
		if v.foldersTree != nil {
			v.foldersTree.OpenBranch(destDir)
		}
	})
}

// treeContextCatcher is a transparent widget over the folders tree that captures
// secondary taps (right-click) to show a context menu for the currently selected folder.
type treeContextCatcher struct {
	widget.BaseWidget
	tree *widget.Tree
	view *MainView
	rect *canvas.Rectangle
}

func (t *treeContextCatcher) CreateRenderer() fyne.WidgetRenderer {
	if t.rect == nil {
		t.rect = canvas.NewRectangle(color.Transparent)
	}
	objs := []fyne.CanvasObject{t.rect}
	return &simpleTreeOverlayRenderer{objs: objs, rect: t.rect}
}

func (t *treeContextCatcher) TappedSecondary(e *fyne.PointEvent) {
	// Use the currently selected folder in the view (kept in sync by setMediaDirectory)
	selected := ""
	if t.view != nil {
		selected = t.view.mediaDir
	}
	if selected == "" {
		return
	}
	// Show context menu with Delete option
	canvas := fyne.CurrentApp().Driver().CanvasForObject(t.tree)
	if canvas == nil {
		return
	}
	menu := fyne.NewMenu("",
		fyne.NewMenuItem("Delete", func() {
			// Confirm deletion
			confirm := dialog.NewConfirm("Delete Folder",
				fmt.Sprintf("Delete folder '%s' and all its media files?", filepath.Base(selected)),
				func(confirmed bool) {
					if !confirmed {
						return
					}
					// Perform deletion: remove DB entries and folder record
					folders, err := t.view.database.GetFolders()
					if err == nil {
						var found *models.Folder
						for _, f := range folders {
							if f.Path == selected {
								found = &f
								break
							}
						}
						if found != nil {
							_ = t.view.database.DeleteMediaFilesByDirectory(found.Path)
							_ = t.view.database.DeleteFolder(found.ID)
						}
					}
					// Clear cache and refresh UI
					delete(t.view.folderCache, selected)
					if strings.HasPrefix(t.view.mediaDir, selected) {
						t.view.mediaDir = ""
						t.view.sortedFiles = nil
					}
					// Rebuild folders tree and refresh grid
					t.view.foldersTree = t.view.createFoldersTree()
					t.view.RefreshMediaGrid()
				}, t.view.window)
			confirm.Show()
		}),
	)
	widget.ShowPopUpMenuAtPosition(menu, canvas, e.AbsolutePosition)
}

// simpleTreeOverlayRenderer implements a minimal renderer for the transparent overlay
type simpleTreeOverlayRenderer struct {
	objs []fyne.CanvasObject
	rect *canvas.Rectangle
}

func (r *simpleTreeOverlayRenderer) Layout(size fyne.Size) {
	if r.rect != nil {
		r.rect.Resize(size)
		r.rect.Move(fyne.NewPos(0, 0))
	}
}

func (r *simpleTreeOverlayRenderer) MinSize() fyne.Size           { return fyne.NewSize(10, 10) }
func (r *simpleTreeOverlayRenderer) Refresh()                     {}
func (r *simpleTreeOverlayRenderer) Objects() []fyne.CanvasObject { return r.objs }
func (r *simpleTreeOverlayRenderer) Destroy()                     {}

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

	if v.isLoadingMedia {
		v.showLoadingPlaceholder()
		return
	}

	// Get current card dimensions (these scale with zoom level)
	cardWidth := components.CardWidth()
	cardHeight := components.CardHeight()

	// Collect all cards
	var cards []fyne.CanvasObject
	focusedCardIndex := -1
	var focusedCard *components.MediaCard
	effectiveFilter := strings.ToLower(v.effectiveFilter())
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
			if effectiveFilter != "" && !strings.Contains(strings.ToLower(fileName), effectiveFilter) {
				continue
			}

			if !v.isLaunchFocusedFile(fileName) {
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
				card = components.NewMediaCardWithForce(mediaFile, v.config.ThumbnailDir, true, func(path, previewPath string) {
					// Store preview path and the source file's modtime used to generate it
					if info, err := os.Stat(path); err == nil {
						v.database.UpdateMediaFilePreviewPath(path, previewPath, info.ModTime())
					} else {
						v.database.UpdateMediaFilePreviewPath(path, previewPath, time.Now())
					}
				})
			} else {
				card = components.NewMediaCard(mediaFile, v.config.ThumbnailDir, func(path, previewPath string) {
					if info, err := os.Stat(path); err == nil {
						v.database.UpdateMediaFilePreviewPath(path, previewPath, info.ModTime())
					} else {
						v.database.UpdateMediaFilePreviewPath(path, previewPath, time.Now())
					}
				})
			}
			// Capture index for closure
			idx := len(cards)
			card.SetOnDelete(func() {
				// Remove the deleted file from v.sortedFiles
				if idx < len(v.sortedFiles) {
					v.sortedFiles = append(v.sortedFiles[:idx], v.sortedFiles[idx+1:]...)
				}
				v.RefreshMediaGrid()
			})

			// Wire drag-and-drop for moving files into folders
			dragPath := filePath
			card.SetOnDragStart(func() {
				v.setDraggedFilePath(dragPath)
			})
			card.SetOnDragEnd(func() {
				v.clearDrag()
			})

			if v.launchFocusActive && focusedCard == nil {
				focusedCard = card
				focusedCardIndex = len(cards)
			}

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

	if focusedCard != nil {
		focusedCard.SetHighlighted(true)
		v.scrollToCardIndex(focusedCardIndex, cardHeight)
	}

	fmt.Printf("Media grid refreshed with card size: %.0fx%.0f, %d files\n", cardWidth, cardHeight, len(cards))
}

func (v *MainView) showLoadingPlaceholder() {
	if v.loadingLabel != nil {
		message := v.loadingMessage
		if message == "" {
			message = "Scanning folders and loading media..."
		}
		v.loadingLabel.SetText(message)
	}

	if v.loadingContainer != nil {
		v.loadingContainer.Show()
		v.loadingContainer.Refresh()
	}

	if v.mediaGridWrapper != nil {
		v.mediaGridWrapper.Content = container.NewCenter(container.NewVBox(
			widget.NewLabel("Please wait while media is being loaded."),
			widget.NewProgressBarInfinite(),
		))
		v.mediaGridWrapper.Refresh()
	}
}

func (v *MainView) hideLoadingPlaceholder() {
	if v.loadingContainer != nil {
		v.loadingContainer.Hide()
		v.loadingContainer.Refresh()
	}
}

func (v *MainView) setMediaLoadingState(loading bool, message string) {
	v.isLoadingMedia = loading
	if message != "" {
		v.loadingMessage = message
	}

	if v.loadingLabel == nil {
		return
	}

	if loading {
		v.showLoadingPlaceholder()
		return
	}

	v.hideLoadingPlaceholder()
}

func (v *MainView) createMediaGrid() fyne.CanvasObject {
	// Use the card dimensions from the components package for consistency
	// These are now functions that return scaled values based on zoom level
	cardWidth := components.CardWidth()
	cardHeight := components.CardHeight()
	var cards []fyne.CanvasObject
	focusedCardIndex := -1
	var focusedCard *components.MediaCard
	effectiveFilter := strings.ToLower(v.effectiveFilter())
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
			if effectiveFilter != "" && !strings.Contains(strings.ToLower(fileName), effectiveFilter) {
				continue
			}

			if !v.isLaunchFocusedFile(fileName) {
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

			card := components.NewMediaCard(mediaFile, v.config.ThumbnailDir, func(path, previewPath string) {
				if info, err := os.Stat(path); err == nil {
					v.database.UpdateMediaFilePreviewPath(path, previewPath, info.ModTime())
				} else {
					v.database.UpdateMediaFilePreviewPath(path, previewPath, time.Now())
				}
			})
			card.SetOnDelete(func() {
				v.RefreshMediaGrid()
			})

			// Wire drag-and-drop for moving files into folders
			dragPath := filePath
			card.SetOnDragStart(func() {
				v.setDraggedFilePath(dragPath)
			})
			card.SetOnDragEnd(func() {
				v.clearDrag()
			})

			if v.launchFocusActive && focusedCard == nil {
				focusedCard = card
				focusedCardIndex = len(cards)
			}

			cards = append(cards, card)
		}
	}
	v.mediaGridContainer = container.NewGridWrap(fyne.NewSize(cardWidth, cardHeight), cards...)

	// Create a scroll container wrapper so the grid can scroll when content exceeds viewport
	// This prevents the window from expanding past screen bounds with many files
	v.mediaGridWrapper = container.NewVScroll(v.mediaGridContainer)
	if focusedCard != nil {
		focusedCard.SetHighlighted(true)
		fyne.Do(func() {
			v.scrollToCardIndex(focusedCardIndex, cardHeight)
		})
	}
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

// loadMediaMetadataAsync loads all file metadata in background to avoid blocking UI.
// Files are shown as soon as filesystem info is available; ffprobe metadata is backfilled
// asynchronously for files that don't already have it in the database.
func (v *MainView) loadMediaMetadataAsync(mediaDir string, callback func([]SortableMediaFile)) {
	go func() {
		var sortableFiles []SortableMediaFile
		var filesNeedingTags []*models.MediaFile

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
				// File doesn't exist in database, create it.
				// Avoid eager ffprobe here to keep large directory loads responsive.
				// Metadata can be filled in later by preview generation/background paths.
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
				}
				err := v.database.CreateMediaFile(newMediaFile)
				if err != nil {
					fmt.Printf("[WARN] Failed to create media file in DB %s: %v\n", filePath, err)
					continue
				}
				mediaFile = newMediaFile
			}

			// Backfill metadata asynchronously if it's missing for a video/image.
			if (mediaFile.Width == 0 && mediaFile.Height == 0 && mediaFile.Duration == 0) &&
				(mediaFile.FileType == "video" || mediaFile.FileType == "image") {
				go v.backfillMetadata(mediaFile)
			}

			// Tag the file if it doesn't have tags yet
			if len(mediaFile.Tags) == 0 {
				// Collect files that need tagging for async processing
				filesNeedingTags = append(filesNeedingTags, mediaFile)
			}

			sortableFiles = append(sortableFiles, SortableMediaFile{
				Entry:     file,
				Info:      info,
				MediaFile: mediaFile,
			})
		}

		// Start async tagging for files that need it
		if len(filesNeedingTags) > 0 {
			go v.tagFilesAsync(filesNeedingTags)
		}

		callback(sortableFiles)
	}()
}

// backfillMetadata extracts width/height/duration asynchronously and updates the DB.
func (v *MainView) backfillMetadata(mediaFile *models.MediaFile) {
	width, height, duration, err := preview.GetMetadata(mediaFile.Path)
	if err != nil {
		fmt.Printf("[DEBUG] Metadata backfill failed for %s: %v\n", mediaFile.Path, err)
		return
	}
	updates := map[string]interface{}{
		"width":    width,
		"height":   height,
		"duration": duration,
	}
	if err := v.database.UpdateMediaFileFields(mediaFile.Path, updates); err != nil {
		fmt.Printf("[WARN] Failed to save backfilled metadata for %s: %v\n", mediaFile.Path, err)
		return
	}
	// Update in-memory model so sorting/filtering can use it without reloading.
	mediaFile.Width = width
	mediaFile.Height = height
	mediaFile.Duration = duration

	// Refresh the grid so duration/dimensions badges update.
	fyne.Do(func() {
		v.RefreshMediaGrid()
	})
}

// tagFilesAsync tags multiple files asynchronously without blocking UI
func (v *MainView) tagFilesAsync(files []*models.MediaFile) {
	for _, mediaFile := range files {
		err := v.tagger.TagMediaFile(mediaFile)
		if err != nil {
			fmt.Printf("[WARN] Failed to tag media file %s: %v\n", mediaFile.Path, err)
		} else {
			fmt.Printf("[DEBUG] Tagged media file: %s\n", mediaFile.Path)
		}
	}

	// After tagging is complete, refresh the tags list and media grid on UI thread
	fyne.Do(func() {
		v.refreshTagsList()
		v.RefreshMediaGrid() // Refresh to show updated titles
	})
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

// checkFolderCache checks if cached data for a folder is still valid
func (v *MainView) checkFolderCache(dir string) (*FolderCache, bool) {
	cache, exists := v.folderCache[dir]
	if !exists {
		return nil, false
	}

	// Check if recursive mode changed
	if cache.RecursiveMode != v.recursiveSearch {
		return nil, false
	}

	// Check if folder modification time changed
	info, err := os.Stat(dir)
	if err != nil {
		return nil, false
	}

	if info.ModTime().After(cache.ModTime) {
		return nil, false
	}

	return cache, true
}

// updateFolderCache updates the cache for a folder
func (v *MainView) updateFolderCache(dir string, files []SortableMediaFile) {
	info, err := os.Stat(dir)
	if err != nil {
		return
	}

	v.folderCache[dir] = &FolderCache{
		Files:         files,
		ModTime:       info.ModTime(),
		RecursiveMode: v.recursiveSearch,
	}
}

// normalizeOpenBranches removes stale OpenBranches entries and normalizes slash
// variants so reopening saved branches doesn't hit nonexistent paths.
// setMediaDirectory changes directory and loads metadata async
func (v *MainView) setMediaDirectory(dir string) {
	v.setMediaDirectoryInternal(dir, false)
}

// SwitchMediaDirectory switches to a directory selected from external UI controls.
func (v *MainView) SwitchMediaDirectory(dir string) {
	v.setMediaDirectory(dir)
}

func (v *MainView) setMediaDirectoryInternal(dir string, forceReload bool) {
	if dir == "" {
		return
	}

	if !forceReload && v.mediaDir == dir {
		return
	}

	fmt.Printf("[DEBUG] Switching to folder: %s\n", dir)

	// Check if we have valid cached data for this folder
	if cache, valid := v.checkFolderCache(dir); valid {
		fmt.Printf("[DEBUG] Using cached data for folder: %s (%d files)\n", dir, len(cache.Files))
		v.mediaDir = dir
		v.sortedFiles = v.sortLoadedFiles(cache.Files)
		v.setMediaLoadingState(false, "")
		v.SaveConfig() // Save selected folder
		v.RefreshMediaGrid()
		v.refreshTagsList()
		return
	}

	// No valid cache, need to load from disk
	fmt.Printf("[DEBUG] Loading folder from disk: %s\n", dir)
	v.mediaDir = dir
	v.sortedFiles = nil // Clear current data while loading
	v.setMediaLoadingState(true, "Scanning folders and loading media...")
	v.SaveConfig() // Save selected folder
	v.RefreshMediaGrid()

	// Load metadata async
	v.loadMediaMetadataAsync(dir, func(files []SortableMediaFile) {
		fyne.Do(func() {
			// Update cache with loaded data
			v.updateFolderCache(dir, files)
			v.sortedFiles = v.sortLoadedFiles(files)
			v.setMediaLoadingState(false, "")
			v.RefreshMediaGrid()
			// Refresh tags list in case new tags were created
			v.refreshTagsList()
		})
	})
}

func (v *MainView) Build() fyne.CanvasObject {
	mainContent := v.buildMainContent()
	v.maybeShowLaunchFocusPrompt()

	if v.downloadsView == nil {
		v.downloadsView = NewDownloadsView(v.database, v.window, v.downloadManager)
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Media", mainContent),
		container.NewTabItem("Downloads", v.downloadsView.Build()),
	)

	// Add debug log panel if enabled
	if v.showDebugLog {
		if v.debugLog == nil {
			v.debugLog = widget.NewTextGridFromString("Debug Log:\n")
			v.debugWriter = components.NewDebugWriter(v.debugLog)
			log.SetOutput(io.MultiWriter(os.Stdout, v.debugWriter))
		}
		debugScroll := container.NewScroll(v.debugLog)
		debugScroll.SetMinSize(fyne.NewSize(0, 200))
		return container.NewBorder(nil, debugScroll, nil, nil, tabs)
	}

	return tabs
}

// updateDebugLogVisibility toggles the debug log panel without rebuilding the entire UI
func (v *MainView) updateDebugLogVisibility() {
	mainContent := v.window.Content()
	if mainContent == nil {
		return
	}

	// If debug log is enabled, wrap the current content with debug panel
	if v.showDebugLog {
		if v.debugLog == nil {
			v.debugLog = widget.NewTextGridFromString("Debug Log:\n")
			v.debugWriter = components.NewDebugWriter(v.debugLog)
			log.SetOutput(io.MultiWriter(os.Stdout, v.debugWriter))
		}
		debugScroll := container.NewScroll(v.debugLog)
		debugScroll.SetMinSize(fyne.NewSize(0, 200))
		v.window.SetContent(container.NewBorder(nil, debugScroll, nil, nil, mainContent))
	} else {
		// If disabling debug log, we need to remove it from the window
		// The structure when debug is on is: Border{top: nil, bottom: debugScroll, left: nil, right: nil, center: mainContent}
		// So we extract the center which is the mainContent
		// Since we can't directly access Border's internal structure, we rebuild but with showDebugLog=false
		// This is safe because toggling off won't trigger infinite recursion
		if v.window.Content() != nil {
			// Just refresh the window content with the current state (showDebugLog is already false)
			mainContent := v.buildMainContent()
			v.window.SetContent(mainContent)
		}
	}
}

// buildMainContent constructs the main content without the debug log wrapper
func (v *MainView) buildMainContent() fyne.CanvasObject {
	v.foldersTree = v.createFoldersTree()
	var treeScroll fyne.CanvasObject
	if v.foldersTree != nil {
		// Use the tree directly in a scroll so primary clicks reach it
		scroll := container.NewScroll(v.foldersTree)
		scroll.SetMinSize(fyne.NewSize(200, 0))
		treeScroll = scroll
		v.treeScroll = scroll // Store reference
	} else {
		treeScroll = container.NewVBox(widget.NewLabel("No folders found"))
		v.treeScroll = nil // No scroll container when no folders
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
		// Refresh grid on UI thread (fyne.Do handles threading)
		v.RefreshMediaGridWithForce(true)
	})

	recursiveCheck := widget.NewCheck("Recursive Search", func(checked bool) {
		v.recursiveSearch = checked
		fmt.Printf("[DEBUG] Recursive Search toggled: %v\n", checked)
		// Clear folder cache since recursive mode affects what files are loaded
		v.folderCache = make(map[string]*FolderCache)
		// Reload metadata when recursive mode changes
		v.setMediaDirectoryInternal(v.mediaDir, true)
		v.SaveConfig() // Save recursive search setting
	})
	recursiveCheck.SetChecked(v.recursiveSearch)

	v.loadingLabel = widget.NewLabel("Scanning folders and loading media...")
	v.loadingContainer = container.NewHBox(widget.NewProgressBarInfinite(), v.loadingLabel)
	if !v.isLoadingMedia {
		v.loadingContainer.Hide()
	}

	debugLogCheck := widget.NewCheck("Show Debug Log", func(checked bool) {
		v.showDebugLog = checked
		v.updateDebugLogVisibility()
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
			v.setMediaDirectory(folderPath)
			v.foldersTree = v.createFoldersTree()
			if v.treeScroll != nil {
				v.treeScroll.Content = v.foldersTree
				v.treeScroll.Refresh()
			}
			v.SaveConfig()
		}, v.window)
		dialog.Resize(fyne.NewSize(800, 600))
		dialog.Show()
	})
	deleteFolderBtn := widget.NewButton("Delete Folder", func() {
		folders, err := v.database.GetFolders()
		if err != nil || len(folders) == 0 {
			dialog := dialog.NewInformation("No Folders", "No folders to delete.", v.window)
			dialog.Show()
			return
		}
		// Create options for select
		options := make([]string, len(folders))
		for i, f := range folders {
			options[i] = f.Name + " (" + f.Path + ")"
		}
		var selectedFolder *models.Folder
		selectWidget := widget.NewSelect(options, func(selected string) {
			// Find the folder
			for _, f := range folders {
				if f.Name+" ("+f.Path+")" == selected {
					selectedFolder = &f
					break
				}
			}
		})
		selectDialog := dialog.NewCustomConfirm("Select Folder to Delete", "Delete", "Cancel",
			container.NewVBox(
				widget.NewLabel("Select a folder to delete:"),
				selectWidget,
			), func(confirmed bool) {
				if !confirmed || selectedFolder == nil {
					return
				}
				// Confirm delete
				confirmDialog := dialog.NewConfirm("Confirm Delete",
					fmt.Sprintf("Delete folder '%s' and all its media files?", selectedFolder.Name),
					func(confirmed bool) {
						if !confirmed {
							return
						}
						// Delete media files
						err := v.database.DeleteMediaFilesByDirectory(selectedFolder.Path)
						if err != nil {
							fmt.Printf("[ERROR] Failed to delete media files: %v\n", err)
							return
						}
						// Delete folder
						err = v.database.DeleteFolder(selectedFolder.ID)
						if err != nil {
							fmt.Printf("[ERROR] Failed to delete folder: %v\n", err)
							return
						}
						// Clear cache
						delete(v.folderCache, selectedFolder.Path)
						// If current dir is under this folder, switch to another
						if strings.HasPrefix(v.mediaDir, selectedFolder.Path) {
							v.mediaDir = ""
							v.sortedFiles = nil
							v.RefreshMediaGrid()
						}
						// Refresh tree
						v.foldersTree = v.createFoldersTree()
						if v.treeScroll != nil {
							v.treeScroll.Content = v.foldersTree
							v.treeScroll.Refresh()
						}
						v.SaveConfig()
					}, v.window)
				confirmDialog.Show()
			}, v.window)
		selectDialog.Show()
	})

	buttonBox := container.NewHBox(
		refreshBtn,
		forceRegenerateBtn,
		addFolderBtn,
		deleteFolderBtn,
		clearTagsBtn,
		recursiveCheck,
		debugLogCheck,
		v.loadingContainer,
	)
	toolbar := container.NewBorder(nil, nil, sortControls, buttonBox, filterEntry)
	if v.mediaDir != "" && v.foldersTree != nil {
		v.foldersTree.Select(v.mediaDir)
	} else if v.foldersTree == nil {
		fmt.Println("[WARN] foldersTree is nil, cannot select root directory")
	}

	return container.NewBorder(toolbar, nil, nil, nil, mainSplit)
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

func NewMainView(cfg *config.Config, db *db.Database, window fyne.Window, mediaDir string, launchFiles []string, downloadManager DownloadManager) *MainView {
	launchFocusFiles := make(map[string]bool)
	for _, filePath := range launchFiles {
		if strings.EqualFold(filepath.Clean(filepath.Dir(filePath)), filepath.Clean(mediaDir)) {
			launchFocusFiles[strings.ToLower(filepath.Base(filePath))] = true
		}
	}
	launchFocusActive := len(launchFocusFiles) > 0

	mv := &MainView{
		config:            cfg,
		database:          db,
		window:            window,
		mediaDir:          mediaDir,
		launchFocusFiles:  launchFocusFiles,
		launchFocusActive: launchFocusActive,
		sortBy:            cfg.SortBy,          // Load from config
		sortAscending:     cfg.SortAscending,   // Load from config
		selectedTags:      cfg.SelectedTags,    // Load from config
		filter:            cfg.FilterText,      // Load from config
		recursiveSearch:   cfg.RecursiveSearch, // Load from config
		showDebugLog:      cfg.ShowDebugLog,    // Load from config
		tagger:            tagger.NewFilenameTagger(db),
		folderCache:       make(map[string]*FolderCache), // Initialize folder cache
		loadingMessage:    "Scanning folders and loading media...",
		downloadManager:   downloadManager,
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
		mv.setMediaLoadingState(true, "Scanning folders and loading media...")
		mv.loadMediaMetadataAsync(mv.mediaDir, func(files []SortableMediaFile) {
			fyne.Do(func() {
				mv.sortedFiles = mv.sortLoadedFiles(files)
				mv.setMediaLoadingState(false, "")
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
