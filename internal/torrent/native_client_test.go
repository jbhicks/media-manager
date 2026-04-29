package torrent

import (
	"os"
	"testing"
	"time"
)

func TestExtractImagesFromTorrent_MagnetWithImages(t *testing.T) {
	// Create temp download dir
	tempDir := t.TempDir()

	client, err := NewNativeClient(tempDir)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test with a known magnet that has images
	// This is a test magnet - replace with a real one for live testing
	magnet := "magnet:?xt=urn:btih:TEST123&dn=test"

	images, err := client.ExtractImagesFromTorrent(magnet, 5*time.Second)

	// We expect an error or timeout for a fake magnet
	if err == nil {
		t.Logf("Got images (unexpected for fake magnet): %v", images)
	} else {
		t.Logf("Expected error for fake magnet: %v", err)
	}
}

func TestExtractImagesFromTorrent_FilePathHandling(t *testing.T) {
	// Test that file paths are handled correctly
	tempDir := t.TempDir()

	client, err := NewNativeClient(tempDir)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Just verify the method exists and can be called
	// Real testing requires actual torrents
	magnet := "magnet:?xt=urn:btih:FAKE123&dn=fake"

	_, err = client.ExtractImagesFromTorrent(magnet, 1*time.Second)
	if err == nil {
		t.Error("Expected error for fake magnet")
	}
}

func TestNewNativeClient_DefaultDir(t *testing.T) {
	client, err := NewNativeClient("")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if client.downloadDir == "" {
		t.Error("Expected default download directory")
	}
}

func TestNewNativeClient_CustomDir(t *testing.T) {
	tempDir := t.TempDir()

	client, err := NewNativeClient(tempDir)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if client.downloadDir != tempDir {
		t.Errorf("Expected download dir %s, got %s", tempDir, client.downloadDir)
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Download directory should exist")
	}
}
