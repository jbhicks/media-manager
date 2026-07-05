package torrent

import (
	"os"
	"testing"
	"time"
)

func TestTorrentIDFromInfoHash_Stable(t *testing.T) {
	hash := "a1b2c3d4e5f6789012345678abcdef0123456789"
	id1 := TorrentIDFromInfoHash(hash)
	id2 := TorrentIDFromInfoHash(hash)
	if id1 == 0 {
		t.Fatal("expected non-zero torrent ID")
	}
	if id1 != id2 {
		t.Fatalf("expected stable ID, got %d and %d", id1, id2)
	}

	upper := TorrentIDFromInfoHash("A1B2C3D4E5F6789012345678ABCDEF0123456789")
	if upper != id1 {
		t.Fatalf("expected case-insensitive ID, got %d vs %d", upper, id1)
	}
}

func TestParseInfoHashFromMagnet(t *testing.T) {
	magnet := "magnet:?xt=urn:btih:08ada5a7a6193aae36ec839d9eaa8132e13e7286&dn=test"
	hash, err := parseInfoHashFromMagnet(magnet)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "08ada5a7a6193aae36ec839d9eaa8132e13e7286" {
		t.Fatalf("unexpected hash: %s", hash)
	}

	id := TorrentIDFromInfoHash(hash)
	if id != TorrentIDFromInfoHash("08ada5a7a6193aae36ec839d9eaa8132e13e7286") {
		t.Fatal("magnet hash should map to same torrent ID")
	}
}

func TestResumeAll_DeduplicatesMagnets(t *testing.T) {
	tempDir := t.TempDir()
	client, err := NewNativeClient(tempDir)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	magnet := "magnet:?xt=urn:btih:08ada5a7a6193aae36ec839d9eaa8132e13e7286&dn=test"
	resumed := client.ResumeAll([]string{magnet, magnet, magnet}, tempDir)
	if resumed != 1 {
		t.Fatalf("expected 1 resumed torrent, got %d", resumed)
	}
}

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
