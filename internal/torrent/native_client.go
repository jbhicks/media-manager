package torrent

import (
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type NativeClient struct {
	client      *torrent.Client
	downloadDir string
	mu          sync.RWMutex
}

// TorrentIDFromInfoHash returns a stable client ID for a torrent info hash.
func TorrentIDFromInfoHash(hash string) int {
	hash = strings.ToLower(strings.TrimSpace(hash))
	if hash == "" {
		return 0
	}
	return int(crc32.ChecksumIEEE([]byte(hash)))
}

func parseInfoHashFromMagnet(magnetLink string) (string, error) {
	m, err := metainfo.ParseMagnetUri(magnetLink)
	if err != nil {
		return "", err
	}
	return strings.ToLower(m.InfoHash.String()), nil
}

func NewNativeClient(downloadDir string) (*NativeClient, error) {
	if downloadDir == "" {
		homeDir, _ := os.UserHomeDir()
		downloadDir = filepath.Join(homeDir, "Downloads")
	}

	if err := os.MkdirAll(downloadDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create download dir: %w", err)
	}

	cfg := torrent.NewDefaultClientConfig()
	// Persist piece progress in the download dir (.torrent.db) so restarts can resume.
	cfg.DataDir = downloadDir
	cfg.ListenPort = 0
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

func (n *NativeClient) findTorrentByHash(hash string) *torrent.Torrent {
	hash = strings.ToLower(strings.TrimSpace(hash))
	for _, t := range n.client.Torrents() {
		if strings.EqualFold(t.InfoHash().String(), hash) || strings.EqualFold(t.InfoHash().HexString(), hash) {
			return t
		}
	}
	return nil
}

func (n *NativeClient) findTorrentByID(torrentID int) *torrent.Torrent {
	for _, t := range n.client.Torrents() {
		if TorrentIDFromInfoHash(t.InfoHash().String()) == torrentID {
			return t
		}
	}
	return nil
}

func (n *NativeClient) ensureDownloading(t *torrent.Torrent) {
	if t.Info() != nil {
		t.DownloadAll()
		return
	}
	go func() {
		select {
		case <-t.GotInfo():
			log.Printf("[TORRENT] Got info for: %s", t.Name())
			t.DownloadAll()
		case <-time.After(5 * time.Minute):
			log.Printf("[TORRENT] Timeout waiting for torrent metadata: %s", t.Name())
		}
	}()
}

func (n *NativeClient) AddTorrent(magnetLink string, downloadDir string) (int, error) {
	if downloadDir != "" {
		n.downloadDir = downloadDir
	}

	hash, err := parseInfoHashFromMagnet(magnetLink)
	if err != nil {
		return 0, fmt.Errorf("invalid magnet link: %w", err)
	}

	id := TorrentIDFromInfoHash(hash)

	n.mu.Lock()
	defer n.mu.Unlock()

	if existing := n.findTorrentByHash(hash); existing != nil {
		log.Printf("[TORRENT] Torrent already active, resuming: %s", existing.Name())
		n.ensureDownloading(existing)
		return id, nil
	}

	t, err := n.client.AddMagnet(magnetLink)
	if err != nil {
		return 0, fmt.Errorf("failed to add magnet: %w", err)
	}

	log.Printf("[TORRENT] Added torrent: %s (hash: %s)", t.Name(), hash[:min(8, len(hash))])
	n.ensureDownloading(t)
	return id, nil
}

func (n *NativeClient) statusFromTorrent(t *torrent.Torrent, torrentID int) map[string]interface{} {
	percentDone := 0.0
	length := t.Length()
	if length > 0 {
		percentDone = float64(t.BytesCompleted()) / float64(length)
	}

	return map[string]interface{}{
		"id":             torrentID,
		"name":           t.Name(),
		"percentDone":    percentDone,
		"totalSize":      length,
		"downloadDir":    n.downloadDir,
		"peersConnected": t.Stats().ActivePeers,
		"infoHash":       t.InfoHash().String(),
	}
}

func (n *NativeClient) GetTorrentStatus(torrentID int) (map[string]interface{}, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if t := n.findTorrentByID(torrentID); t != nil {
		return n.statusFromTorrent(t, torrentID), nil
	}

	return nil, fmt.Errorf("torrent not found")
}

func (n *NativeClient) RemoveTorrent(torrentID int, deleteData bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	t := n.findTorrentByID(torrentID)
	if t == nil {
		return fmt.Errorf("torrent not found")
	}

	t.Drop()
	log.Printf("[TORRENT] Removed torrent: %s (deleteData=%v)", t.Name(), deleteData)
	return nil
}

// ResumeAll re-attaches magnets after a service restart so partial files can continue.
func (n *NativeClient) ResumeAll(magnets []string, downloadDir string) int {
	resumed := 0
	seen := make(map[string]struct{})

	for _, magnet := range magnets {
		magnet = strings.TrimSpace(magnet)
		if magnet == "" {
			continue
		}

		hash, err := parseInfoHashFromMagnet(magnet)
		if err != nil {
			log.Printf("[TORRENT] Skipping invalid magnet during resume: %v", err)
			continue
		}
		if _, ok := seen[hash]; ok {
			continue
		}
		seen[hash] = struct{}{}

		if _, err := n.AddTorrent(magnet, downloadDir); err != nil {
			log.Printf("[TORRENT] Failed to resume torrent %s: %v", hash[:min(8, len(hash))], err)
			continue
		}
		resumed++
	}

	return resumed
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
	n.ensureDownloading(t)
	return nil
}

// ExtractImagesFromTorrent adds a magnet, waits for metadata, and downloads only image files
func (n *NativeClient) ExtractImagesFromTorrent(magnetLink string, timeout time.Duration) ([]string, error) {
	t, err := n.client.AddMagnet(magnetLink)
	if err != nil {
		return nil, fmt.Errorf("failed to add magnet: %w", err)
	}

	log.Printf("[TORRENT] Inspecting torrent for images: %s", t.Name())

	select {
	case <-t.GotInfo():
		log.Printf("[TORRENT] Got metadata for: %s", t.Name())
	case <-time.After(timeout):
		t.Drop()
		return nil, fmt.Errorf("timeout waiting for torrent metadata")
	}

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
			t.Files()[i].Download()
		}
	}

	if len(imageFiles) == 0 {
		log.Printf("[TORRENT] No image files found in torrent: %s", t.Name())
		t.Drop()
		return nil, nil
	}

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

	var imagePaths []string
	for _, imgPath := range imageFiles {
		fullPath := filepath.Join(n.downloadDir, imgPath)
		if _, err := os.Stat(fullPath); err == nil {
			imagePaths = append(imagePaths, fullPath)
		}
	}

	log.Printf("[TORRENT] Extracted %d images from torrent", len(imagePaths))
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}