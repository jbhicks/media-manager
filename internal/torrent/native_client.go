package torrent

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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
