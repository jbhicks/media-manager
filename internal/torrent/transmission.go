package torrent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hekmon/transmissionrpc/v3"
)

type TransmissionClient struct {
	client *transmissionrpc.Client
}

func NewTransmissionClient(host, username, password string) (*TransmissionClient, error) {
	rpcURL, err := url.Parse(fmt.Sprintf("http://%s/transmission/rpc", host))
	if err != nil {
		return nil, fmt.Errorf("failed to parse transmission URL: %w", err)
	}

	if username != "" {
		rpcURL.User = url.UserPassword(username, password)
	}

	config := &transmissionrpc.Config{
		CustomClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	client, err := transmissionrpc.New(rpcURL, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create transmission client: %w", err)
	}

	return &TransmissionClient{
		client: client,
	}, nil
}

func (t *TransmissionClient) AddTorrent(magnetLink string, downloadDir string) (int, error) {
	ctx := context.Background()

	payload := transmissionrpc.TorrentAddPayload{
		Filename: &magnetLink,
	}

	if downloadDir != "" {
		payload.DownloadDir = &downloadDir
	}

	torrent, err := t.client.TorrentAdd(ctx, payload)
	if err != nil {
		return 0, fmt.Errorf("failed to add torrent: %w", err)
	}

	if torrent.ID == nil {
		return 0, fmt.Errorf("torrent ID is nil")
	}

	return int(*torrent.ID), nil
}

func (t *TransmissionClient) GetTorrentStatus(torrentID int) (map[string]interface{}, error) {
	ctx := context.Background()

	torrents, err := t.client.TorrentGetAllFor(ctx, []int64{int64(torrentID)})
	if err != nil {
		return nil, fmt.Errorf("failed to get torrent status: %w", err)
	}

	if len(torrents) == 0 {
		return nil, fmt.Errorf("torrent not found")
	}

	torrent := torrents[0]

	status := map[string]interface{}{
		"id": torrentID,
	}

	if torrent.Name != nil {
		status["name"] = *torrent.Name
	}
	if torrent.Status != nil {
		status["status"] = int64(*torrent.Status)
	}
	if torrent.PercentDone != nil {
		status["percentDone"] = *torrent.PercentDone
	}
	if torrent.DownloadDir != nil {
		status["downloadDir"] = *torrent.DownloadDir
	}
	if torrent.TotalSize != nil {
		status["totalSize"] = *torrent.TotalSize
	}
	if torrent.ETA != nil {
		status["eta"] = *torrent.ETA
	}
	if torrent.RateDownload != nil {
		status["rateDownload"] = *torrent.RateDownload
	}
	if torrent.RateUpload != nil {
		status["rateUpload"] = *torrent.RateUpload
	}
	if torrent.PeersConnected != nil {
		status["peersConnected"] = *torrent.PeersConnected
	}
	if torrent.Error != nil {
		status["error"] = *torrent.Error
	}
	if torrent.ErrorString != nil {
		status["errorString"] = *torrent.ErrorString
	}

	return status, nil
}

func (t *TransmissionClient) RemoveTorrent(torrentID int, deleteData bool) error {
	ctx := context.Background()

	payload := transmissionrpc.TorrentRemovePayload{
		IDs:             []int64{int64(torrentID)},
		DeleteLocalData: deleteData,
	}

	err := t.client.TorrentRemove(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to remove torrent: %w", err)
	}

	return nil
}
