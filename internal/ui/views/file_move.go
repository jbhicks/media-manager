package views

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const uniqueNameAttempts = 1000

// destPathForMove returns the intended destination path for moving srcPath
// into destDir. ok is false when the drop is a no-op (empty inputs or the
// file is already in destDir).
func destPathForMove(srcPath, destDir string) (string, bool) {
	srcPath = normalizeTreePath(srcPath)
	destDir = normalizeTreePath(destDir)
	if srcPath == "" || destDir == "" {
		return "", false
	}
	if treePathKey(filepath.Dir(srcPath)) == treePathKey(destDir) {
		return "", false
	}
	destPath := normalizeTreePath(filepath.Join(destDir, filepath.Base(srcPath)))
	if treePathKey(srcPath) == treePathKey(destPath) {
		return "", false
	}
	return destPath, true
}

// allocateUniquePath returns path if it does not exist, otherwise
// "name (N).ext" so a move never silently overwrites.
func allocateUniquePath(path string, exists func(string) bool) string {
	if exists == nil {
		exists = pathExists
	}
	if !exists(path) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; i < uniqueNameAttempts; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if !exists(candidate) {
			return candidate
		}
	}
	return path
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// belongsInCurrentView reports whether filePath should appear when browsing mediaDir.
func belongsInCurrentView(mediaDir, filePath string, recursive bool) bool {
	mediaKey := treePathKey(mediaDir)
	fileKey := treePathKey(filePath)
	if mediaKey == "" || fileKey == "" {
		return false
	}
	if treePathKey(filepath.Dir(filePath)) == mediaKey {
		return true
	}
	if !recursive {
		return false
	}
	return strings.HasPrefix(fileKey, mediaKey+string(filepath.Separator))
}

// renameFile is os.Rename, overridable in tests to simulate cross-volume failure.
var renameFile = os.Rename

type moveProgress func(fraction float64)

// moveMediaOnDisk moves srcPath into destDir. It never overwrites: name
// collisions get a "name (N).ext" suffix. Cross-volume drops fall back to
// copy + delete. Returns ("", nil) for a no-op drop onto the file's current folder.
func moveMediaOnDisk(srcPath, destDir string) (string, error) {
	return moveMediaOnDiskProgress(srcPath, destDir, nil)
}

func moveMediaOnDiskProgress(srcPath, destDir string, progress moveProgress) (string, error) {
	srcPath = normalizeTreePath(srcPath)
	destDir = normalizeTreePath(destDir)

	destPath, ok := destPathForMove(srcPath, destDir)
	if !ok {
		return "", nil
	}

	info, err := os.Stat(srcPath)
	if err != nil {
		return "", fmt.Errorf("source file not found: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("source is a directory: %s", srcPath)
	}

	destInfo, err := os.Stat(destDir)
	if err != nil {
		return "", fmt.Errorf("destination folder not found: %w", err)
	}
	if !destInfo.IsDir() {
		return "", fmt.Errorf("destination is not a folder: %s", destDir)
	}

	destPath = allocateUniquePath(destPath, pathExists)
	if err := renameFile(srcPath, destPath); err != nil {
		if copyErr := copyThenRemove(srcPath, destPath, info.Mode(), info.Size(), progress); copyErr != nil {
			return "", fmt.Errorf("failed to move %s to %s: %w", srcPath, destPath, copyErr)
		}
	} else if progress != nil {
		progress(1)
	}
	return destPath, nil
}

func copyThenRemove(src, dst string, mode os.FileMode, size int64, progress moveProgress) error {
	if err := copyFileExclusive(src, dst, mode, size, progress); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("copied to %s but could not remove source: %w", dst, err)
	}
	if progress != nil {
		progress(1)
	}
	return nil
}

func copyFileExclusive(src, dst string, mode os.FileMode, size int64, progress moveProgress) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	pw := &progressWriter{w: out, total: size, cb: progress}
	_, copyErr := io.Copy(pw, in)
	closeErr := out.Close()
	if copyErr != nil {
		os.Remove(dst)
		return copyErr
	}
	if closeErr != nil {
		os.Remove(dst)
		return closeErr
	}
	return nil
}

type progressWriter struct {
	w     io.Writer
	n     int64
	total int64
	cb    moveProgress
	last  float64
}

func (p *progressWriter) Write(b []byte) (int, error) {
	n, err := p.w.Write(b)
	p.n += int64(n)
	if p.cb != nil && p.total > 0 {
		frac := float64(p.n) / float64(p.total)
		if frac-p.last >= 0.01 || frac >= 1 {
			p.last = frac
			p.cb(frac)
		}
	}
	return n, err
}