package views

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDestPathForMove_NoOpSameFolder(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "clip.mp4")
	if _, ok := destPathForMove(src, dir); ok {
		t.Fatalf("expected no-op when dropping onto the current folder")
	}
}

func TestDestPathForMove_EmptyInputs(t *testing.T) {
	if _, ok := destPathForMove("", filepath.Join(t.TempDir(), "media")); ok {
		t.Fatalf("expected no-op for empty source")
	}
	if _, ok := destPathForMove(filepath.Join(t.TempDir(), "clip.mp4"), ""); ok {
		t.Fatalf("expected no-op for empty dest")
	}
}

func TestDestPathForMove_DifferentFolder(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "inbox", "clip.mp4")
	destDir := filepath.Join(root, "shows")
	got, ok := destPathForMove(src, destDir)
	if !ok {
		t.Fatal("expected a destination path")
	}
	want := normalizeTreePath(filepath.Join(destDir, "clip.mp4"))
	if treePathKey(got) != treePathKey(want) {
		t.Fatalf("dest path = %q, want %q", got, want)
	}
}

func TestAllocateUniquePath_NoCollision(t *testing.T) {
	exists := func(string) bool { return false }
	path := filepath.Join("media", "clip.mp4")
	if got := allocateUniquePath(path, exists); got != path {
		t.Fatalf("got %q, want original path", got)
	}
}

func TestAllocateUniquePath_SkipsExisting(t *testing.T) {
	path := filepath.Join("media", "clip.mp4")
	taken := map[string]bool{
		path:                                   true,
		filepath.Join("media", "clip (1).mp4"): true,
	}
	exists := func(p string) bool { return taken[p] }
	got := allocateUniquePath(path, exists)
	want := filepath.Join("media", "clip (2).mp4")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBelongsInCurrentView(t *testing.T) {
	root := t.TempDir()
	inRoot := filepath.Join(root, "clip.mp4")
	inChild := filepath.Join(root, "season1", "clip.mp4")
	elsewhere := filepath.Join(filepath.Dir(root), "inbox", "clip.mp4")

	if !belongsInCurrentView(root, inRoot, false) {
		t.Fatal("file in folder should belong")
	}
	if belongsInCurrentView(root, inChild, false) {
		t.Fatal("nested file should not belong without recursive search")
	}
	if !belongsInCurrentView(root, inChild, true) {
		t.Fatal("nested file should belong with recursive search")
	}
	if belongsInCurrentView(root, elsewhere, true) {
		t.Fatal("file outside root should not belong")
	}
}

func TestMoveMediaOnDisk_MovesFile(t *testing.T) {
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

	got, err := moveMediaOnDisk(src, destDir)
	if err != nil {
		t.Fatalf("moveMediaOnDisk: %v", err)
	}
	want := filepath.Join(destDir, "clip.mp4")
	if treePathKey(got) != treePathKey(want) {
		t.Fatalf("moved to %q, want %q", got, want)
	}
	if pathExists(src) {
		t.Fatal("source file should be gone after move")
	}
	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "video" {
		t.Fatalf("moved content = %q", data)
	}
}

func TestMoveMediaOnDisk_NoOpSameFolder(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "clip.mp4")
	if err := os.WriteFile(src, []byte("video"), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := moveMediaOnDisk(src, root)
	if err != nil {
		t.Fatalf("moveMediaOnDisk: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty dest for no-op, got %q", got)
	}
	if !pathExists(src) {
		t.Fatal("source should still exist after no-op")
	}
}

func TestMoveMediaOnDisk_UniqueNameOnCollision(t *testing.T) {
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
	existing := filepath.Join(destDir, "clip.mp4")
	if err := os.WriteFile(src, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existing, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := moveMediaOnDisk(src, destDir)
	if err != nil {
		t.Fatalf("moveMediaOnDisk: %v", err)
	}
	want := filepath.Join(destDir, "clip (1).mp4")
	if treePathKey(got) != treePathKey(want) {
		t.Fatalf("moved to %q, want unique %q", got, want)
	}
	old, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if string(old) != "old" {
		t.Fatalf("existing file was overwritten: %q", old)
	}
	if pathExists(src) {
		t.Fatal("source file should be gone after move")
	}
}

func TestMoveMediaOnDisk_RejectsNonFolderDest(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "clip.mp4")
	notDir := filepath.Join(root, "file.txt")
	if err := os.WriteFile(src, []byte("video"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(notDir, []byte("nope"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := moveMediaOnDisk(src, notDir); err == nil {
		t.Fatal("expected error when dest is not a folder")
	}
	if !pathExists(src) {
		t.Fatal("source should remain when move fails")
	}
}
func TestMoveMediaOnDisk_CopyFallbackWhenRenameFails(t *testing.T) {
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

	orig := renameFile
	renameFile = func(oldpath, newpath string) error {
		return fmt.Errorf("simulated cross-volume rename")
	}
	t.Cleanup(func() { renameFile = orig })

	got, err := moveMediaOnDisk(src, destDir)
	if err != nil {
		t.Fatalf("moveMediaOnDisk: %v", err)
	}
	want := filepath.Join(destDir, "clip.mp4")
	if treePathKey(got) != treePathKey(want) {
		t.Fatalf("moved to %q, want %q", got, want)
	}
	if pathExists(src) {
		t.Fatal("source file should be gone after copy-fallback move")
	}
	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "video" {
		t.Fatalf("moved content = %q", data)
	}
}
