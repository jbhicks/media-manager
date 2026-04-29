package torrent

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type NativeClient struct {
	client      *torrent.Client
	downloadDir string
}

func NewNativeClient(downloadDir string) (*NativeClient, error) {
	if downloadDir == "" {
		homeDir, _ := os.UserHomeDir()
		downloadDir = filepath.Join(homeDir, "Downloads")
	}

	os.MkdirAll(downloadDir, 0755)

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = downloadDir
	cfg.ListenPort = 0 // Let system auto-assign available port
	cfg.Seed = true
	cfg.Debug = false

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create torrent client: %w", err)
	}

	log.Printf("[TORRENT] Native torrent client initialized, download dir: %s", downloadDir)

	return &NativeClient{
		client:      client,
		downloadDir: downloadDir,
	}, nil
}

func (n *NativeClient) AddTorrent(magnetLink string, downloadDir string) (int, error) {
	if downloadDir != "" {
		n.downloadDir = downloadDir
	}

	t, err := n.client.AddMagnet(magnetLink)
	if err != nil {
		return 0, fmt.Errorf("failed to add magnet: %w", err)
	}

	log.Printf("[TORRENT] Added torrent: %s", t.Name())

	go func() {
		<-t.GotInfo()
		log.Printf("[TORRENT] Got info for: %s", t.Name())
		t.DownloadAll()
	}()

	return int(t.InfoHash().HexString()[0]), nil
}

func (n *NativeClient) GetTorrentStatus(torrentID int) (map[string]interface{}, error) {
	torrents := n.client.Torrents()
	if len(torrents) == 0 {
		return nil, fmt.Errorf("no torrents found")
	}

	for _, t := range torrents {
		info := t.Info()
		if info == nil {
			continue
		}

		status := map[string]interface{}{
			"id":             torrentID,
			"name":           t.Name(),
			"percentDone":    float64(t.BytesCompleted()) / float64(t.Length()),
			"totalSize":      t.Length(),
			"downloadDir":    n.downloadDir,
			"peersConnected": t.Stats().ActivePeers,
		}

		return status, nil
	}

	return nil, fmt.Errorf("torrent not found")
}

func (n *NativeClient) RemoveTorrent(torrentID int, deleteData bool) error {
	torrents := n.client.Torrents()
	for _, t := range torrents {
		t.Drop()
		if deleteData {
			log.Printf("[TORRENT] Removing torrent data (not implemented)")
		}
	}
	return nil
}

func (n *NativeClient) Close() {
	if n.client != nil {
		n.client.Close()
	}
}

func (n *NativeClient) AddTorrentFromFile(torrentFile string, downloadDir string) error {
	if downloadDir != "" {
		n.downloadDir = downloadDir
	}

	mi, err := metainfo.LoadFromFile(torrentFile)
	if err != nil {
		return fmt.Errorf("failed to load torrent file: %w", err)
	}

	t, err := n.client.AddTorrent(mi)
	if err != nil {
		return fmt.Errorf("failed to add torrent: %w", err)
	}

	log.Printf("[TORRENT] Added torrent from file: %s", t.Name())

	go func() {
		<-t.GotInfo()
		log.Printf("[TORRENT] Got info for: %s", t.Name())
		t.DownloadAll()
	}()

	return nil
}

// ExtractImagesFromTorrent adds a magnet, waits for metadata, and downloads only image files
func (n *NativeClient) ExtractImagesFromTorrent(magnetLink string, timeout time.Duration) ([]string, error) {
	t, err := n.client.AddMagnet(magnetLink)
	if err != nil {
		return nil, fmt.Errorf("failed to add magnet: %w", err)
	}

	log.Printf("[TORRENT] Inspecting torrent for images: %s", t.Name())

	// Wait for metadata with timeout
	select {
	case <-t.GotInfo():
		log.Printf("[TORRENT] Got metadata for: %s", t.Name())
	case <-time.After(timeout):
		t.Drop()
		return nil, fmt.Errorf("timeout waiting for torrent metadata")
	}

	// Find image files
	info := t.Info()
	if info == nil {
		t.Drop()
		return nil, fmt.Errorf("no torrent info available")
	}

	var imageFiles []string
	for i, file := range info.Files {
		pathStr := filepath.Join(file.Path...)
		name := strings.ToLower(pathStr)
		if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") ||
			strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".gif") ||
			strings.HasSuffix(name, ".webp") {
			log.Printf("[TORRENT] Found image file: %s", pathStr)
			imageFiles = append(imageFiles, pathStr)
			// Download only this file
			t.Files()[i].Download()
		}
	}

	if len(imageFiles) == 0 {
		log.Printf("[TORRENT] No image files found in torrent: %s", t.Name())
		t.Drop()
		return nil, nil
	}

	// Wait for images to download
	log.Printf("[TORRENT] Downloading %d image files...", len(imageFiles))
	deadline := time.Now().Add(timeout)
	for {
		allDone := true
		for i := range info.Files {
			pathStr := filepath.Join(info.Files[i].Path...)
			name := strings.ToLower(pathStr)
			if strings.HasSuffix(name, ".jpg") || strings.HasSuffix(name, ".jpeg") ||
				strings.HasSuffix(name, ".png") || strings.HasSuffix(name, ".gif") ||
				strings.HasSuffix(name, ".webp") {
				if t.Files()[i].BytesCompleted() < t.Files()[i].Length() {
					allDone = false
					break
				}
			}
		}
		if allDone {
			break
		}
		if time.Now().After(deadline) {
			log.Printf("[TORRENT] Timeout downloading images")
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Build full paths
	var imagePaths []string
	for _, imgPath := range imageFiles {
		fullPath := filepath.Join(n.downloadDir, imgPath)
		if _, err := os.Stat(fullPath); err == nil {
			imagePaths = append(imagePaths, fullPath)
		}
	}

	log.Printf("[TORRENT] Extracted %d images from torrent", len(imagePaths))

	// Drop the torrent so we don't seed the full content
	t.Drop()

	return imagePaths, nil
}

func (n *NativeClient) WaitForCompletion(timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			torrents := n.client.Torrents()
			allComplete := true
			for _, t := range torrents {
				stats := t.Stats()
				if stats.BytesWrittenData.Int64() < t.Length() {
					allComplete = false
					log.Printf("[TORRENT] %s: %.2f%% complete",
						t.Name(),
						float64(stats.BytesWrittenData.Int64())/float64(t.Length())*100)
				}
			}
			if allComplete && len(torrents) > 0 {
				log.Println("[TORRENT] All downloads complete")
				return
			}
			if time.Now().After(deadline) {
				log.Println("[TORRENT] Download timeout reached")
				return
			}
		}
	}
}
