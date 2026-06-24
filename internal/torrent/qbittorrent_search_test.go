package torrent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanQBittorrentPlugins(t *testing.T) {
	dir := t.TempDir()

	plugin1 := `#VERSION: 1.0
class testone(object):
    url = 'https://example.com/one'
    name = 'Test One'
    supported_categories = {'all': '0'}
`
	plugin2 := `#VERSION: 2.0
class testtwo(object):
    url = 'https://example.com/two'
    name = 'Test Two'
    supported_categories = {'all': '0'}
`
	if err := os.WriteFile(filepath.Join(dir, "testone.py"), []byte(plugin1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "testtwo.py"), []byte(plugin2), 0644); err != nil {
		t.Fatal(err)
	}
	// Non-python file should be ignored
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	plugins := scanQBittorrentPlugins(dir)
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	if plugins[0].FullName != "Test One" {
		t.Errorf("expected plugin name 'Test One', got %q", plugins[0].FullName)
	}
	if plugins[0].URL != "https://example.com/one" {
		t.Errorf("expected plugin URL 'https://example.com/one', got %q", plugins[0].URL)
	}
	if plugins[1].FullName != "Test Two" {
		t.Errorf("expected plugin name 'Test Two', got %q", plugins[1].FullName)
	}
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, c := range cases {
		got := FormatBytes(c.bytes)
		if got != c.expected {
			t.Errorf("FormatBytes(%d) = %q, want %q", c.bytes, got, c.expected)
		}
	}
}
