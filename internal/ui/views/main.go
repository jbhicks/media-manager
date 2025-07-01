package views

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/user/media-manager/internal/config"
	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/preview"
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
	cmd := []string{"ffmpeg", "-y", "-i", videoPath, "-ss", "00:00:01.000", "-vframes", "1", thumbPath}
	fmt.Printf("[DEBUG] Generating video thumbnail: %v\n", cmd)
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
}

func NewMainView(cfg *config.Config, database *db.Database) *MainView {
	fmt.Printf("[DEBUG] main.go: MainView using MediaDirs: %v\n", cfg.MediaDirs)
	return &MainView{
		config:   cfg,
		database: database,
	}
}

func (v *MainView) Build() fyne.CanvasObject {
	// Check ffmpeg availability
	ffmpegMissing := false
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		ffmpegMissing = true
	}

	// Create toolbar
	toolbar := v.createToolbar()

	// Create sidebar
	sidebar := v.createSidebar()

	// Create main content area (media grid)
	v.mediaGridContainer = container.NewMax(v.createMediaGrid())

	// Create status bar
	statusBar := v.createStatusBar()

	// Optional ffmpeg warning banner
	var content fyne.CanvasObject
	if ffmpegMissing {
		banner := widget.NewLabel("Warning: ffmpeg not found. Video previews are disabled.")
		bannerContainer := container.NewVBox(banner, widget.NewSeparator())
		content = container.NewVBox(bannerContainer, container.NewBorder(
			toolbar,   // top
			statusBar, // bottom
			nil,       // left
			nil,       // right
			container.NewHSplit(sidebar, v.mediaGridContainer), // center
		))
	} else {
		mainContent := container.NewHSplit(sidebar, v.mediaGridContainer)
		mainContent.SetOffset(0.25)
		content = container.NewBorder(
			toolbar,     // top
			statusBar,   // bottom
			nil,         // left
			nil,         // right
			mainContent, // center
		)
	}
	return content
}

func (v *MainView) createToolbar() fyne.CanvasObject {
	searchEntry := widget.NewEntry()
	searchEntry.OnChanged = func(input string) {
		v.filterMediaFiles(input)
	}

	refreshBtn := widget.NewButton("Refresh", func() {
		v.refreshMediaGrid()
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
		v.refreshMediaGrid()
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

func (v *MainView) refreshMediaGrid() {
	if v.mediaGridContainer != nil {
		v.mediaGridContainer.Objects = []fyne.CanvasObject{v.createMediaGrid()}
		v.mediaGridContainer.Refresh()
	}
	fmt.Println("Media grid refreshed")
}

func (v *MainView) createMediaGrid() fyne.CanvasObject {
	grid := container.NewGridWithColumns(4)
	if len(v.config.MediaDirs) > 0 {
		mediaDir := v.config.MediaDirs[0]
		fmt.Printf("[DEBUG] main.go: Loading media from: %s\n", mediaDir)
		files, err := os.ReadDir(mediaDir)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() {
					name := file.Name()
					if len(name) > 20 {
						name = name[:17] + "..."
					}
					var previewBox fyne.CanvasObject
					if isImageFile(file.Name()) {
						img := canvas.NewImageFromFile(filepath.Join(mediaDir, file.Name()))
						img.FillMode = canvas.ImageFillContain
						img.SetMinSize(fyne.NewSize(100, 80))
						previewBox = img
					} else if isVideoFile(file.Name()) {
						homeDir, _ := os.UserHomeDir()
						thumbDir := filepath.Join(homeDir, ".media-manager", "thumbnails")
						staticThumbPath := filepath.Join(thumbDir, file.Name()+".jpg")
						animatedGifPath := filepath.Join(thumbDir, file.Name()+".gif")
						videoPath := filepath.Join(mediaDir, file.Name())

						if _, err := os.Stat(staticThumbPath); err == nil {
							// Static thumbnail exists, check for animated GIF
							if _, err := os.Stat(animatedGifPath); err != nil {
								// Generate animated GIF if it doesn't exist
								go preview.GenerateAnimatedPreview(videoPath, animatedGifPath)
							}

							// Create video preview card with GIF animation support
							videoCard := components.NewVideoPreviewCard(staticThumbPath, animatedGifPath, name)
							videoCard.Resize(fyne.NewSize(120, 120))
							grid.Add(videoCard)
							continue // Skip the rest of the card creation logic
						} else {
							// Static thumbnail doesn't exist, generate it first
							go ensureVideoThumbnail(videoPath, staticThumbPath)
							previewBox = widget.NewIcon(theme.FileVideoIcon())
						}
					} else {
						previewBox = widget.NewIcon(theme.FileIcon())
					}
					// Create a regular card for non-video files or videos without thumbnails
					label := widget.NewLabelWithStyle(name, fyne.TextAlignCenter, fyne.TextStyle{})
					card := container.NewVBox(
						previewBox,
						label,
					)
					card.Resize(fyne.NewSize(120, 120))
					grid.Add(card)
				}
			}
		} else {
			fmt.Printf("[DEBUG] main.go: Failed to read media dir: %v\n", err)
		}
	} else {
		fmt.Printf("[DEBUG] main.go: No media directory set\n")
	}
	scroll := container.NewScroll(grid)
	return scroll
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
