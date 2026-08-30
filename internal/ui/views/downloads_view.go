package views

import (
	"fmt"
	"image/color"
	"log"
	"sort"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/torrent"
	"github.com/user/media-manager/pkg/models"
)

// DownloadsView provides the "Downloads" tab UI for managing torrent search
// hosts and searching for titles across them.
type DownloadsView struct {
	database        *db.Database
	window          fyne.Window
	downloadManager DownloadManager
	sources         []models.DownloadSource
	selectedSource  *models.DownloadSource
	searchResults   []models.SearchResult
	isSearching     bool

	hostsList      *widget.List
	resultsList    *widget.List
	statusLabel    *widget.Label
	searchEntry    *widget.Entry
	categorySelect *widget.Select
	sortSelect     *widget.Select

	hostStatus map[uint]string // source ID -> reachable status text
	mu         sync.RWMutex
}

// NewDownloadsView creates a new downloads view.
func NewDownloadsView(database *db.Database, window fyne.Window, downloadManager DownloadManager) *DownloadsView {
	v := &DownloadsView{
		database:        database,
		window:          window,
		downloadManager: downloadManager,
		hostStatus:      make(map[uint]string),
	}
	v.loadSources()
	return v
}

func (v *DownloadsView) loadSources() {
	sources, err := v.database.GetDownloadSources()
	if err != nil {
		log.Printf("[DownloadsView] Failed to load sources: %v", err)
		return
	}
	v.sources = sources
}

// Build constructs the downloads tab UI.
func (v *DownloadsView) Build() fyne.CanvasObject {
	left := v.buildHostsPanel()
	center := v.buildSearchPanel()

	split := container.NewHSplit(left, center)
	split.SetOffset(0.28)

	return split
}

func (v *DownloadsView) buildHostsPanel() fyne.CanvasObject {
	header := canvas.NewText("Search Hosts", color.White)
	header.TextSize = 16
	header.TextStyle = fyne.TextStyle{Bold: true}

	addBtn := widget.NewButtonWithIcon("Add", theme.ContentAddIcon(), func() {
		v.showHostDialog(nil)
	})
	discoverBtn := widget.NewButtonWithIcon("Discover", theme.SearchIcon(), func() {
		v.discoverHosts()
	})
	toolbar := container.NewHBox(addBtn, discoverBtn)

	v.hostsList = widget.NewList(
		func() int { return len(v.sources) },
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			status := canvas.NewCircle(color.Gray{Y: 128})
			status.Resize(fyne.NewSize(10, 10))
			name := widget.NewLabel("Host name")
			name.Truncation = fyne.TextTruncateEllipsis
			return container.NewHBox(check, status, name)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(v.sources) {
				return
			}
			source := v.sources[id]
			box := item.(*fyne.Container)
			check := box.Objects[0].(*widget.Check)
			status := box.Objects[1].(*canvas.Circle)
			name := box.Objects[2].(*widget.Label)

			check.SetChecked(source.Enabled)
			check.OnChanged = func(checked bool) {
				source.Enabled = checked
				if err := v.database.UpdateDownloadSource(&source); err != nil {
					log.Printf("[DownloadsView] Failed to update source: %v", err)
				}
			}

			v.mu.RLock()
			statusText := v.hostStatus[source.ID]
			v.mu.RUnlock()
			switch {
			case statusText == "ok":
				status.FillColor = color.NRGBA{G: 200, A: 255}
			case strings.HasPrefix(statusText, "err"):
				status.FillColor = color.NRGBA{R: 200, A: 255}
			default:
				status.FillColor = color.Gray{Y: 128}
			}
			status.Refresh()

			name.SetText(source.Name)
		},
	)

	v.hostsList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(v.sources) {
			s := v.sources[id]
			v.selectedSource = &s
		}
	}

	v.hostsList.OnUnselected = func(id widget.ListItemID) {
		v.selectedSource = nil
	}

	editBtn := widget.NewButtonWithIcon("Edit", theme.DocumentCreateIcon(), func() {
		if v.selectedSource != nil {
			v.showHostDialog(v.selectedSource)
		}
	})
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if v.selectedSource != nil {
			v.confirmDeleteHost(v.selectedSource)
		}
	})
	testBtn := widget.NewButtonWithIcon("Test", theme.ConfirmIcon(), func() {
		v.testSelectedHost()
	})
	btns := container.NewHBox(editBtn, deleteBtn, testBtn)

	return container.NewBorder(
		container.NewVBox(container.NewBorder(nil, nil, header, toolbar), widget.NewSeparator()),
		btns, nil, nil,
		v.hostsList,
	)
}

func (v *DownloadsView) buildSearchPanel() fyne.CanvasObject {
	v.searchEntry = widget.NewEntry()
	v.searchEntry.SetPlaceHolder("Search for a title...")

	v.categorySelect = widget.NewSelect([]string{"all", "movies", "tv", "music", "games", "software", "books", "anime"}, nil)
	v.categorySelect.SetSelected("all")

	v.sortSelect = widget.NewSelect([]string{"Seeders", "Size", "Name"}, func(selected string) {
		v.sortResults(selected)
	})
	v.sortSelect.SetSelected("Seeders")

	searchBtn := widget.NewButtonWithIcon("Search", theme.SearchIcon(), func() {
		v.runSearch()
	})

	v.statusLabel = widget.NewLabel("Ready")

	toolbar := container.NewBorder(
		container.NewHBox(widget.NewLabel("Category:"), v.categorySelect, widget.NewLabel("Sort:"), v.sortSelect),
		nil, nil, container.NewHBox(searchBtn, v.statusLabel),
		v.searchEntry,
	)

	v.resultsList = widget.NewList(
		func() int { return len(v.searchResults) },
		func() fyne.CanvasObject {
			title := widget.NewLabel("Result title")
			title.Truncation = fyne.TextTruncateEllipsis
			meta := widget.NewLabel("Size · S/L · Source")
			meta.TextStyle = fyne.TextStyle{Italic: true}
			meta.Truncation = fyne.TextTruncateEllipsis
			downloadBtn := widget.NewButtonWithIcon("Download", theme.DownloadIcon(), nil)
			return container.NewBorder(nil, nil, nil, downloadBtn, container.NewVBox(title, meta))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(v.searchResults) {
				return
			}
			r := v.searchResults[id]
			box := item.(*fyne.Container)
			leftBox := box.Objects[0].(*fyne.Container)
			title := leftBox.Objects[0].(*widget.Label)
			meta := leftBox.Objects[1].(*widget.Label)
			downloadBtn := box.Objects[1].(*widget.Button)

			title.SetText(r.Title)
			meta.SetText(fmt.Sprintf("%s · %d/%d · %s",
				torrent.FormatBytes(r.Size), r.Seeders, r.Leechers, r.Indexer))

			downloadBtn.OnTapped = func() {
				v.downloadResult(r)
			}
		},
	)

	return container.NewBorder(toolbar, nil, nil, nil, v.resultsList)
}

func (v *DownloadsView) showHostDialog(source *models.DownloadSource) {
	isEdit := source != nil
	nameEntry := widget.NewEntry()
	urlEntry := widget.NewEntry()
	usernameEntry := widget.NewEntry()
	passwordEntry := widget.NewPasswordEntry()
	typeSelect := widget.NewSelect([]string{"qbittorrent", "jackett", "rarbg"}, nil)
	priorityEntry := widget.NewEntry()

	if isEdit {
		nameEntry.SetText(source.Name)
		urlEntry.SetText(source.URL)
		usernameEntry.SetText(source.Username)
		passwordEntry.SetText(source.Password)
		typeSelect.SetSelected(source.Type)
		priorityEntry.SetText(fmt.Sprintf("%d", source.Priority))
	} else {
		typeSelect.SetSelected("qbittorrent")
		priorityEntry.SetText("0")
	}

	items := []*widget.FormItem{
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Type", typeSelect),
		widget.NewFormItem("URL", urlEntry),
		widget.NewFormItem("Username", usernameEntry),
		widget.NewFormItem("Password", passwordEntry),
		widget.NewFormItem("Priority", priorityEntry),
	}

	dialog.ShowForm("Search Host", "Save", "Cancel", items, func(ok bool) {
		if !ok {
			return
		}
		priority := 0
		fmt.Sscanf(priorityEntry.Text, "%d", &priority)

		s := models.DownloadSource{}
		if isEdit {
			s = *source
		}
		s.Name = nameEntry.Text
		s.Type = typeSelect.Selected
		s.URL = urlEntry.Text
		s.Username = usernameEntry.Text
		s.Password = passwordEntry.Text
		s.Priority = priority
		s.Enabled = true
		if isEdit {
			s.Enabled = source.Enabled
		}

		var err error
		if isEdit {
			err = v.database.UpdateDownloadSource(&s)
		} else {
			err = v.database.CreateDownloadSource(&s)
		}
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to save host: %w", err), v.window)
			return
		}
		v.refreshHosts()
	}, v.window)
}

func (v *DownloadsView) confirmDeleteHost(source *models.DownloadSource) {
	dialog.ShowConfirm("Delete Host",
		fmt.Sprintf("Delete search host '%s'?", source.Name),
		func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := v.database.DeleteDownloadSource(source.ID); err != nil {
				dialog.ShowError(fmt.Errorf("failed to delete host: %w", err), v.window)
				return
			}
			v.selectedSource = nil
			v.refreshHosts()
		}, v.window)
}

func (v *DownloadsView) refreshHosts() {
	v.loadSources()
	v.hostsList.Refresh()
	v.hostsList.UnselectAll()
}

func (v *DownloadsView) discoverHosts() {
	go func() {
		sources := torrent.DiscoverQBittorrentEndpoints()
		if len(sources) == 0 {
			fyne.Do(func() {
				dialog.ShowInformation("Discover Hosts", "No qBittorrent endpoints or plugins found on this machine.", v.window)
			})
			return
		}

		added := 0
		for i := range sources {
			existing := false
			for _, s := range v.sources {
				if s.URL == sources[i].URL && s.Type == sources[i].Type {
					existing = true
					break
				}
			}
			if existing {
				continue
			}
			if err := v.database.CreateDownloadSource(&sources[i]); err != nil {
				log.Printf("[DownloadsView] Failed to add discovered source: %v", err)
				continue
			}
			added++
		}

		fyne.Do(func() {
			v.refreshHosts()
			v.statusLabel.SetText(fmt.Sprintf("Discovered %d host(s)", added))
		})
	}()
}

func (v *DownloadsView) testSelectedHost() {
	if v.selectedSource == nil {
		return
	}
	source := *v.selectedSource

	go func() {
		var err error
		switch source.Type {
		case "qbittorrent":
			provider := torrent.NewQBittorrentSearchProvider(&source)
			err = provider.TestConnection()
		default:
			err = fmt.Errorf("test not implemented for type %s", source.Type)
		}

		v.mu.Lock()
		if err != nil {
			v.hostStatus[source.ID] = "err: " + err.Error()
		} else {
			v.hostStatus[source.ID] = "ok"
		}
		v.mu.Unlock()

		fyne.Do(func() {
			v.hostsList.Refresh()
			if err != nil {
				v.statusLabel.SetText(fmt.Sprintf("%s unreachable", source.Name))
			} else {
				v.statusLabel.SetText(fmt.Sprintf("%s reachable", source.Name))
			}
		})
	}()
}

func (v *DownloadsView) runSearch() {
	if v.isSearching {
		return
	}
	query := strings.TrimSpace(v.searchEntry.Text)
	if query == "" {
		return
	}

	v.isSearching = true
	v.statusLabel.SetText("Searching...")
	v.searchResults = nil
	v.resultsList.Refresh()

	category := v.categorySelect.Selected
	if category == "" {
		category = "all"
	}

	go func() {
		searcher := torrent.NewTorrentSearcher()
		for i := range v.sources {
			s := v.sources[i]
			if !s.Enabled {
				continue
			}
			var provider torrent.SearchProvider
			var err error
			switch s.Type {
			case "qbittorrent":
				provider = torrent.NewQBittorrentSearchProvider(&s)
			case "jackett":
				provider = torrent.NewJackettProvider(&s)
			case "rarbg":
				provider, err = torrent.NewRARBGProvider(&s)
			default:
				log.Printf("[DownloadsView] Unknown source type: %s", s.Type)
				continue
			}
			if err != nil {
				log.Printf("[DownloadsView] Failed to create provider %s: %v", s.Name, err)
				continue
			}
			searcher.AddProvider(provider)
		}

		results, timings, err := searcher.Search(query, category, 100, nil)
		if err != nil {
			log.Printf("[DownloadsView] Search error: %v", err)
		}

		for _, t := range timings {
			if t.Error != "" {
				log.Printf("[DownloadsView] %s error: %s", t.Name, t.Error)
			}
		}

		fyne.Do(func() {
			v.searchResults = results
			v.sortResults(v.sortSelect.Selected)
			v.isSearching = false
			v.statusLabel.SetText(fmt.Sprintf("Found %d result(s)", len(results)))
		})
	}()
}

func (v *DownloadsView) sortResults(by string) {
	if v.resultsList == nil {
		return
	}
	switch by {
	case "Seeders":
		sort.Slice(v.searchResults, func(i, j int) bool {
			return v.searchResults[i].Seeders > v.searchResults[j].Seeders
		})
	case "Size":
		sort.Slice(v.searchResults, func(i, j int) bool {
			return v.searchResults[i].Size > v.searchResults[j].Size
		})
	case "Name":
		sort.Slice(v.searchResults, func(i, j int) bool {
			return strings.ToLower(v.searchResults[i].Title) < strings.ToLower(v.searchResults[j].Title)
		})
	}
	v.resultsList.Refresh()
}

func (v *DownloadsView) downloadResult(r models.SearchResult) {
	if r.MagnetLink == "" && r.TorrentURL == "" {
		dialog.ShowInformation("No Download", "This result has no magnet link or torrent URL.", v.window)
		return
	}

	if v.downloadManager == nil {
		dialog.ShowInformation("No Download Manager", "Download manager is not available.", v.window)
		return
	}

	go func() {
		task, err := v.downloadManager.AddSearchResult(&r)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("failed to start download: %w", err), v.window)
				v.statusLabel.SetText("Download failed")
			})
			return
		}

		fyne.Do(func() {
			v.statusLabel.SetText(fmt.Sprintf("Started download: %s", task.Title))
		})
	}()
}
