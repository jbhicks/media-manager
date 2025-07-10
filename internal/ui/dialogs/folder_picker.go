package dialogs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// FolderPickerDialog is a custom dialog for selecting a folder.
type FolderPickerDialog struct {
	customDialog     dialog.Dialog // Add this field
	currentPathEntry *widget.Entry
	folderList       *widget.List
	currentPath      string
	onSelected       func(string)
	window           fyne.Window
}

// NewFolderPickerDialog creates a new custom folder selection dialog.
func NewFolderPickerDialog(onSelected func(string), window fyne.Window) *FolderPickerDialog {
	picker := &FolderPickerDialog{
		onSelected: onSelected,
		window:     window,
	}

	picker.currentPathEntry = widget.NewEntry()
	picker.currentPathEntry.Disable() // User cannot type directly

	upButton := widget.NewButton("Up", func() {
		picker.navigateUp()
	})

	picker.folderList = widget.NewList(
		func() int {
			return len(picker.getFoldersInCurrentPath())
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Folder Name")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			folders := picker.getFoldersInCurrentPath()
			if id < len(folders) {
				label.SetText(filepath.Base(folders[id]))
			}
		},
	)

	picker.folderList.OnSelected = func(id widget.ListItemID) {
		folders := picker.getFoldersInCurrentPath()
		if id < len(folders) {
			picker.navigateTo(folders[id])
		}
	}

	// Set initial path to user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/" // Fallback to root
	}
	picker.navigateTo(homeDir)

	content := container.NewBorder(
		container.NewBorder(nil, nil, upButton, nil, picker.currentPathEntry),
		nil,
		nil,
		nil,
		picker.folderList,
	)

	confirmButton := widget.NewButton("Select", func() {
		if picker.onSelected != nil {
			picker.onSelected(picker.currentPath)
		}
		picker.Hide()
	})

	cancelButton := widget.NewButton("Cancel", func() {
		picker.Hide()
	})

	// Create the custom dialog and assign it to customDialog field
	picker.customDialog = dialog.NewCustom("Select Folder", "Cancel", content, window)
	// Cast to *dialog.CustomDialog to access SetButtons
	if customDialog, ok := picker.customDialog.(*dialog.CustomDialog); ok {
		customDialog.SetButtons([]fyne.CanvasObject{confirmButton, cancelButton})

		// Set dialog size to a percentage of the parent window
		parentSize := window.Canvas().Size()
		dialogWidth := parentSize.Width * 0.8
		dialogHeight := parentSize.Height * 0.8
		customDialog.Resize(fyne.NewSize(dialogWidth, dialogHeight))
		window.CenterOnScreen() // Corrected: call CenterOnScreen on the window
	} else {
		fmt.Println("Error: customDialog is not of type *dialog.CustomDialog")
	}

	return picker
}

// Show displays the folder picker dialog.
func (p *FolderPickerDialog) Show() {
	p.customDialog.Show()
}

// Hide hides the folder picker dialog.
func (p *FolderPickerDialog) Hide() {
	p.customDialog.Hide()
}

func (p *FolderPickerDialog) navigateTo(path string) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error navigating to %s: %v\n", path, err)
		return // Corrected: return on its own line
	}
	p.currentPath = path
	p.currentPathEntry.SetText(path)
	p.folderList.Refresh()
}

func (p *FolderPickerDialog) navigateUp() {
	parent := filepath.Dir(p.currentPath)
	if parent != p.currentPath { // Prevent navigating above root
		p.navigateTo(parent)
	}
}

func (p *FolderPickerDialog) getFoldersInCurrentPath() []string {
	entries, err := os.ReadDir(p.currentPath)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", p.currentPath, err)
		return nil // Corrected: return on its own line
	}

	var folders []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") { // Exclude hidden folders
			folders = append(folders, filepath.Join(p.currentPath, entry.Name()))
		}
	}
	return folders
}
