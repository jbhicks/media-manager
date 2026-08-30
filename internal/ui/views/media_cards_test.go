package views

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/user/media-manager/internal/ui/components"
	"github.com/user/media-manager/pkg/models"
)

func TestGetOrCreateMediaCard_ReusesSamePointer(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	mf := models.MediaFile{Path: filepath.Join("C:", "media", "clip.mp4"), Filename: "clip.mp4"}
	first := v.getOrCreateMediaCard(mf.Path, mf, false)
	second := v.getOrCreateMediaCard(mf.Path, mf, false)
	if first == nil || second == nil {
		t.Fatal("expected non-nil cards")
	}
	if first != second {
		t.Fatal("getOrCreateMediaCard should reuse the same card pointer for the same path")
	}
}

func TestGetOrCreateMediaCard_ForceMakesNew(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	mf := models.MediaFile{Path: filepath.Join("C:", "media", "clip.mp4"), Filename: "clip.mp4"}
	first := v.getOrCreateMediaCard(mf.Path, mf, false)
	forced := v.getOrCreateMediaCard(mf.Path, mf, true)
	if first == nil || forced == nil {
		t.Fatal("expected non-nil cards")
	}
	if first == forced {
		t.Fatal("forceRegenerate should create a new MediaCard pointer")
	}
}

func TestRemoveSortedFile_DropsCardAndEntry(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	dir := t.TempDir()
	path := filepath.Join(dir, "clip.mp4")
	if err := os.WriteFile(path, []byte("video"), 0644); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 dir entry, got %d", len(entries))
	}

	v := &MainView{
		mediaDir:    dir,
		sortedFiles: []SortableMediaFile{{Entry: entries[0]}},
	}
	mf := models.MediaFile{Path: path, Filename: "clip.mp4"}
	if card := v.getOrCreateMediaCard(path, mf, false); card == nil {
		t.Fatal("expected card")
	}
	key := treePathKey(path)
	if _, ok := v.mediaCards[key]; !ok {
		t.Fatalf("mediaCards missing key %q", key)
	}

	v.removeSortedFile(path)
	if len(v.sortedFiles) != 0 {
		t.Fatalf("sortedFiles len = %d, want 0", len(v.sortedFiles))
	}
	if _, ok := v.mediaCards[key]; ok {
		t.Fatalf("mediaCards still contains key %q after removeSortedFile", key)
	}
}

func TestRefreshMediaGrid_ReusesCards(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	dir := t.TempDir()
	for _, name := range []string{"a.mp4", "b.mp4"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("video"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var sorted []SortableMediaFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		sorted = append(sorted, SortableMediaFile{Entry: e})
	}
	if len(sorted) != 2 {
		t.Fatalf("expected 2 files, got %d", len(sorted))
	}

	w := test.NewWindow(widget.NewLabel("x"))
	defer w.Close()

	v := &MainView{
		window:      w,
		mediaDir:    dir,
		sortedFiles: sorted,
	}
	v.RefreshMediaGrid()
	if len(v.mediaCards) != 2 {
		t.Fatalf("after first refresh mediaCards len = %d, want 2", len(v.mediaCards))
	}

	snapshot := make(map[string]*components.MediaCard, len(v.mediaCards))
	for key, card := range v.mediaCards {
		if card == nil {
			t.Fatalf("nil card for %q", key)
		}
		snapshot[key] = card
	}

	v.RefreshMediaGrid()
	if len(v.mediaCards) != 2 {
		t.Fatalf("after second refresh mediaCards len = %d, want 2", len(v.mediaCards))
	}
	for key, before := range snapshot {
		after, ok := v.mediaCards[key]
		if !ok {
			t.Fatalf("missing card for %q after second refresh", key)
		}
		if after != before {
			t.Fatalf("RefreshMediaGrid rebuilt card for %q instead of reusing it", key)
		}
	}
}
