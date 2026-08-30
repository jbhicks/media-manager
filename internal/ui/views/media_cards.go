package views

import (
	"os"
	"path/filepath"
	"time"

	"github.com/user/media-manager/internal/ui/components"
	"github.com/user/media-manager/pkg/models"
)

func (v *MainView) ensureMediaCards() {
	if v.mediaCards == nil {
		v.mediaCards = make(map[string]*components.MediaCard)
	}
}

func (v *MainView) forgetMediaCards() {
	v.mediaCards = make(map[string]*components.MediaCard)
}

func (v *MainView) previewSaver() func(path, previewPath string) {
	return func(path, previewPath string) {
		if v.database == nil {
			return
		}
		if info, err := os.Stat(path); err == nil {
			v.database.UpdateMediaFilePreviewPath(path, previewPath, info.ModTime())
		} else {
			v.database.UpdateMediaFilePreviewPath(path, previewPath, time.Now())
		}
	}
}

func (v *MainView) getOrCreateMediaCard(filePath string, mediaFile models.MediaFile, forceRegenerate bool) *components.MediaCard {
	v.ensureMediaCards()
	key := treePathKey(filePath)
	if !forceRegenerate {
		if card, ok := v.mediaCards[key]; ok && card != nil {
			v.attachCardHandlers(card, filePath)
			return card
		}
	}
	var card *components.MediaCard
	thumbDir := ""
	if v.config != nil {
		thumbDir = v.config.ThumbnailDir
	}
	if forceRegenerate {
		card = components.NewMediaCardWithForce(mediaFile, thumbDir, true, v.previewSaver())
	} else {
		card = components.NewMediaCard(mediaFile, thumbDir, v.previewSaver())
	}
	v.mediaCards[key] = card
	v.attachCardHandlers(card, filePath)
	return card
}

func (v *MainView) attachCardHandlers(card *components.MediaCard, filePath string) {
	if card == nil {
		return
	}
	card.SetOnDelete(func() {
		v.snapshotMediaScroll()
		v.removeSortedFile(filePath)
		v.RefreshMediaGrid()
	})
	v.wireCardDrag(card, filePath)
}

func (v *MainView) removeSortedFile(filePath string) {
	key := treePathKey(filePath)
	for i, sf := range v.sortedFiles {
		var sfPath string
		if v.recursiveSearch {
			sfPath = sf.Entry.Name()
		} else if v.mediaDir != "" {
			sfPath = filepath.Join(v.mediaDir, sf.Entry.Name())
		} else {
			sfPath = sf.Entry.Name()
		}
		if treePathKey(sfPath) == key {
			v.sortedFiles = append(v.sortedFiles[:i], v.sortedFiles[i+1:]...)
			break
		}
	}
	if v.mediaCards != nil {
		delete(v.mediaCards, key)
	}
}

func (v *MainView) pruneMediaCards(keep map[string]struct{}) {
	v.ensureMediaCards()
	for key := range v.mediaCards {
		if _, ok := keep[key]; !ok {
			delete(v.mediaCards, key)
		}
	}
}

func (v *MainView) mediaCardKeepKeys() map[string]struct{} {
	keep := make(map[string]struct{}, len(v.sortedFiles))
	for _, sf := range v.sortedFiles {
		if sf.Entry == nil {
			continue
		}
		var filePath string
		if v.recursiveSearch {
			filePath = sf.Entry.Name()
		} else if v.mediaDir != "" {
			filePath = filepath.Join(v.mediaDir, sf.Entry.Name())
		} else {
			filePath = sf.Entry.Name()
		}
		keep[treePathKey(filePath)] = struct{}{}
	}
	return keep
}
