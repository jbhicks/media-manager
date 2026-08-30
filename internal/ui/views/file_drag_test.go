package views

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/user/media-manager/internal/ui/components"
	"github.com/user/media-manager/pkg/models"
)

func TestFolderIDAt_HitsVisibleNode(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	n := newFolderNode(v)
	n.id = filepath.Join("C:", "media", "shows")
	n.SetText("shows")
	n.Resize(fyne.NewSize(180, 28))

	w := test.NewWindow(n)
	defer w.Close()
	w.Resize(fyne.NewSize(400, 300))

	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(n)
	got := v.folderIDAt(fyne.NewPos(pos.X+10, pos.Y+10))
	if got != n.id {
		t.Fatalf("folderIDAt = %q, want %q (node pos=%v size=%v)", got, n.id, pos, n.Size())
	}

	miss := v.folderIDAt(fyne.NewPos(pos.X+500, pos.Y+500))
	if miss != "" {
		t.Fatalf("expected miss far from node, got %q", miss)
	}
}

func TestFolderIDAt_IgnoresVirtualRoot(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	n := newFolderNode(v)
	n.id = ""
	n.SetText("/")
	n.Resize(fyne.NewSize(180, 28))

	w := test.NewWindow(n)
	defer w.Close()
	w.Resize(fyne.NewSize(400, 300))

	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(n)
	if got := v.folderIDAt(fyne.NewPos(pos.X+5, pos.Y+5)); got != "" {
		t.Fatalf("virtual root should not be droppable, got %q", got)
	}
}

func TestDragGhost_FollowsPointerAndHides(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	label := widget.NewLabel("background")
	wrapped := v.wrapWithDragGhost(label)
	w := test.NewWindow(wrapped)
	defer w.Close()
	w.Resize(fyne.NewSize(800, 600))

	v.beginCardDrag("clip.mp4")
	if v.dragGhostLabel == nil || v.dragGhostLabel.Text != "clip.mp4" {
		t.Fatalf("ghost label = %v", v.dragGhostLabel)
	}

	v.updateCardDrag(fyne.NewPos(100, 80))
	if !v.dragGhostBg.Visible() {
		t.Fatal("ghost should be visible during drag")
	}
	first := v.dragGhostBg.Position()

	v.updateCardDrag(fyne.NewPos(220, 140))
	second := v.dragGhostBg.Position()
	if second == first {
		t.Fatalf("ghost should follow pointer, stayed at %v", first)
	}
	want := fyne.NewPos(220+dragGhostOffsetX, 140+dragGhostOffsetY)
	if second != want {
		t.Fatalf("ghost pos = %v, want %v", second, want)
	}

	v.endCardDrag()
	if v.dragGhostBg.Visible() {
		t.Fatal("ghost should hide when drag ends")
	}
}

func TestHighlightFolderAt_MarksNodeUnderPointer(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	n := newFolderNode(v)
	n.id = filepath.Join(t.TempDir(), "shows")
	n.SetText("shows")
	n.Resize(fyne.NewSize(200, 28))

	wrapped := v.wrapWithDragGhost(n)
	w := test.NewWindow(wrapped)
	defer w.Close()
	w.Resize(fyne.NewSize(400, 300))

	pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(n)
	v.highlightFolderAt(fyne.NewPos(pos.X+8, pos.Y+8))
	if n.Importance != widget.HighImportance {
		t.Fatalf("folder under pointer Importance = %v, want HighImportance", n.Importance)
	}
	if !n.TextStyle.Bold {
		t.Fatal("folder under pointer should be bold")
	}

	v.clearFolderHighlight()
	if n.Importance != widget.MediumImportance {
		t.Fatalf("cleared highlight Importance = %v, want MediumImportance", n.Importance)
	}
}

func TestWireCardDrag_ShowsGhostHighlightsAndClears(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	n := newFolderNode(v)
	n.id = filepath.Join(t.TempDir(), "shows")
	n.SetText("shows")
	n.Resize(fyne.NewSize(200, 28))

	wrapped := v.wrapWithDragGhost(container.NewWithoutLayout(n))
	w := test.NewWindow(wrapped)
	defer w.Close()
	w.Resize(fyne.NewSize(500, 400))
	n.Move(fyne.NewPos(0, 40))
	n.Refresh()

	card := components.NewMediaCard(models.MediaFile{
		Path:     filepath.Join(t.TempDir(), "clip.mp4"),
		Filename: "clip.mp4",
	}, t.TempDir(), nil)
	v.wireCardDrag(card, card.FilePath())

	card.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{AbsolutePosition: fyne.NewPos(20, 20)},
		Dragged:    fyne.Delta{DX: 10, DY: 0},
	})
	if v.dragGhostBg == nil || !v.dragGhostBg.Visible() {
		t.Fatal("filename ghost should appear once the card is dragged")
	}
	if v.dragGhostLabel.Text != "clip.mp4" {
		t.Fatalf("ghost text = %q, want clip.mp4", v.dragGhostLabel.Text)
	}

	nodePos := fyne.CurrentApp().Driver().AbsolutePositionForObject(n)
	overFolder := fyne.NewPos(nodePos.X+10, nodePos.Y+10)
	card.Dragged(&fyne.DragEvent{
		PointEvent: fyne.PointEvent{AbsolutePosition: overFolder},
		Dragged:    fyne.Delta{DX: 8, DY: 12},
	})
	if n.Importance != widget.HighImportance {
		t.Fatal("folder under pointer should highlight during Dragged")
	}

	card.DragEnd()
	if v.dragGhostBg.Visible() {
		t.Fatal("ghost should hide on DragEnd")
	}
	if n.Importance != widget.MediumImportance {
		t.Fatal("folder highlight should clear on DragEnd")
	}
}

func TestHandleCardDrop_NoOpWhenNotOverFolder(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	v.handleCardDrop(filepath.Join(t.TempDir(), "clip.mp4"), fyne.NewPos(0, 0))
}

func TestMoveFileToFolder_MovesOnDisk(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := test.NewWindow(widget.NewLabel("x"))
	defer w.Close()

	root := t.TempDir()
	srcDir := filepath.Join(root, "inbox")
	destDir := filepath.Join(root, "shows")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(srcDir, "clip.mp4")
	if err := os.WriteFile(src, []byte("video"), 0644); err != nil {
		t.Fatal(err)
	}

	v := &MainView{window: w, folderCache: map[string]*FolderCache{}}
	v.moveFileToFolder(src, destDir)

	if pathExists(src) {
		t.Fatal("source file should be gone after move")
	}
	want := filepath.Join(destDir, "clip.mp4")
	if !pathExists(want) {
		t.Fatalf("expected file at %s", want)
	}
}

func TestDragLayer_MinSizeDoesNotFollowGhost(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	label := widget.NewLabel("background")
	wrapped := v.wrapWithDragGhost(label)
	w := test.NewWindow(wrapped)
	defer w.Close()
	w.Resize(fyne.NewSize(400, 300))

	before := wrapped.MinSize()

	v.beginCardDrag("clip.mp4")
	v.updateCardDrag(fyne.NewPos(800, 700))
	v.endCardDrag()

	after := wrapped.MinSize()
	if after.Width-before.Width > 100 || after.Height-before.Height > 100 {
		t.Fatalf("wrapped MinSize jumped from %v to %v after drag", before, after)
	}
}

func TestSnapshotMediaScroll_IgnoresZeroAfterClamp(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	rect := canvas.NewRectangle(color.White)
	rect.SetMinSize(fyne.NewSize(100, 2000))
	scroll := container.NewVScroll(rect)
	w := test.NewWindow(scroll)
	defer w.Close()
	w.Resize(fyne.NewSize(120, 300))

	v := &MainView{
		mediaGridWrapper: scroll,
		mediaDir:         "/media",
		lastScrollDir:    "/media",
	}
	scroll.Offset = fyne.NewPos(0, 400)
	v.snapshotMediaScroll()
	scroll.Offset = fyne.NewPos(0, 0) // simulate Fyne clamp-to-zero layout
	v.snapshotMediaScroll()
	if v.mediaScrollOffset.Y < 350 || v.mediaScrollOffset.Y > 450 {
		t.Fatalf("mediaScrollOffset.Y = %v, want ~400", v.mediaScrollOffset.Y)
	}
}

func TestRestoreMediaScroll_AppliesAfterZero(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	rect := canvas.NewRectangle(color.White)
	rect.SetMinSize(fyne.NewSize(100, 2000))
	scroll := container.NewVScroll(rect)
	w := test.NewWindow(scroll)
	defer w.Close()
	w.Resize(fyne.NewSize(120, 300))

	v := &MainView{
		mediaGridWrapper: scroll,
		mediaDir:         "/media",
		lastScrollDir:    "/media",
	}
	scroll.Offset = fyne.NewPos(0, 400)
	v.snapshotMediaScroll()
	scroll.Offset = fyne.NewPos(0, 0)
	v.restoreMediaScroll()
	if scroll.Offset.Y < 350 || scroll.Offset.Y > 450 {
		t.Fatalf("Offset.Y = %v, want ~400", scroll.Offset.Y)
	}
}

func TestForgetMediaScroll_OnFolderChange(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	rect := canvas.NewRectangle(color.White)
	rect.SetMinSize(fyne.NewSize(100, 2000))
	scroll := container.NewVScroll(rect)
	w := test.NewWindow(scroll)
	defer w.Close()
	w.Resize(fyne.NewSize(120, 300))

	v := &MainView{
		mediaGridWrapper: scroll,
		mediaDir:         "/old",
		lastScrollDir:    "/old",
	}
	scroll.Offset = fyne.NewPos(0, 400)
	v.snapshotMediaScroll()
	v.mediaDir = "/new"
	v.lastScrollDir = "/old"
	scroll.Offset = fyne.NewPos(0, 0)
	v.restoreMediaScroll()
	if scroll.Offset.Y > 1 {
		t.Fatalf("Offset.Y = %v, want 0 after folder change", scroll.Offset.Y)
	}
}

func TestDragLayer_ImplementsScrollable(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	wrapped := v.wrapWithDragGhost(widget.NewLabel("x"))
	stack, ok := wrapped.(*fyne.Container)
	if !ok {
		t.Fatalf("wrapWithDragGhost should return a container, got %T", wrapped)
	}
	if len(stack.Objects) < 2 {
		t.Fatalf("stack should have content + drag layer, got %d objects", len(stack.Objects))
	}
	last := stack.Objects[len(stack.Objects)-1]
	if _, ok := last.(fyne.Scrollable); !ok {
		t.Fatalf("stack last child %T does not implement fyne.Scrollable", last)
	}
}

func TestStatusBar_BeginUpdateEnd(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	v := &MainView{}
	w := test.NewWindow(v.wrapRoot(widget.NewLabel("content")))
	defer w.Close()
	w.Resize(fyne.NewSize(400, 200))

	if v.statusBar == nil || v.moveProgressBar == nil || v.moveStatusLabel == nil {
		t.Fatal("wrapRoot/ensureStatusBar should create status widgets")
	}
	if v.statusBar.Visible() {
		t.Fatal("status bar should start hidden")
	}

	v.beginMoveProgress("clip.mp4 → shows")
	if v.moveStatusLabel.Text != "clip.mp4 → shows" {
		t.Fatalf("label = %q", v.moveStatusLabel.Text)
	}
	if !v.moveProgressBar.Visible() {
		t.Fatal("progress bar should be visible after begin")
	}
	if !v.statusBar.Visible() {
		t.Fatal("status bar should be visible after begin")
	}

	v.moveProgressBar.SetValue(0.4)
	if got := v.moveProgressBar.Value; got != 0.4 {
		t.Fatalf("progress = %v, want 0.4", got)
	}

	v.endMoveProgress()
	if v.statusBar.Visible() {
		t.Fatal("status bar should hide after end")
	}
	if v.moveProgressBar.Visible() {
		t.Fatal("progress bar should hide after end")
	}
	if v.moveStatusLabel.Text != "" {
		t.Fatalf("label should be cleared, got %q", v.moveStatusLabel.Text)
	}
}
