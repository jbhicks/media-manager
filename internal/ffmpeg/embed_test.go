package ffmpeg

import (
	"testing"
)

func TestEmbeddedFiles(t *testing.T) {
	entries, err := ffmpegBinaries.ReadDir("bin")
	if err != nil {
		t.Fatalf("failed to read bin dir: %v", err)
	}
	if len(entries) == 0 {
		t.Errorf("no files found in embedded bin dir")
	}
	for _, entry := range entries {
		t.Logf("found embedded file: %s", entry.Name())
	}
}
