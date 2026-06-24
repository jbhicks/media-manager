package torrent

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/user/media-manager/pkg/models"
)

// QBittorrentSearchProvider uses qBittorrent's Web API search endpoints.
// It starts a search job, polls for completion, and retrieves results.
type QBittorrentSearchProvider struct {
	name     string
	enabled  bool
	sourceID uint
	baseURL  string
	username string
	password string
	client   *http.Client
	cookies  *cookiejar.Jar
}

// NewQBittorrentSearchProvider creates a provider from a DownloadSource.
func NewQBittorrentSearchProvider(source *models.DownloadSource) *QBittorrentSearchProvider {
	jar, _ := cookiejar.New(nil)
	return &QBittorrentSearchProvider{
		name:     source.Name,
		enabled:  source.Enabled,
		sourceID: source.ID,
		baseURL:  strings.TrimSuffix(source.URL, "/"),
		username: source.Username,
		password: source.Password,
		cookies:  jar,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

func (q *QBittorrentSearchProvider) GetName() string {
	return q.name
}

func (q *QBittorrentSearchProvider) IsEnabled() bool {
	return q.enabled
}

func (q *QBittorrentSearchProvider) SetEnabled(enabled bool) {
	q.enabled = enabled
}

// login authenticates with the qBittorrent Web UI and stores the session cookie.
func (q *QBittorrentSearchProvider) login() error {
	if q.username == "" && q.password == "" {
		// Try unauthenticated access
		return nil
	}

	loginURL := fmt.Sprintf("%s/api/v2/auth/login", q.baseURL)
	data := url.Values{}
	data.Set("username", q.username)
	data.Set("password", q.password)

	resp, err := q.client.PostForm(loginURL, data)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(body), "Ok") {
		return fmt.Errorf("login failed: %s", string(body))
	}
	return nil
}

// Search starts a search job across enabled qBittorrent plugins and returns results.
func (q *QBittorrentSearchProvider) Search(query string, category string, indexers []string) ([]models.SearchResult, error) {
	// Ensure authenticated
	if err := q.login(); err != nil {
		return nil, err
	}

	// Normalize category
	cat := "all"
	if category != "" && category != "all" {
		cat = category
	}

	// Determine plugins
	plugins := "enabled"
	if len(indexers) > 0 {
		plugins = strings.Join(indexers, "|")
	}

	// Start search
	jobID, err := q.startSearch(query, plugins, cat)
	if err != nil {
		return nil, fmt.Errorf("failed to start search: %w", err)
	}
	defer q.deleteSearch(jobID)

	// Poll for completion
	if err := q.waitForSearch(jobID, 30*time.Second); err != nil {
		return nil, err
	}

	// Fetch results
	return q.getResults(jobID)
}

func (q *QBittorrentSearchProvider) startSearch(query, plugins, category string) (int, error) {
	startURL := fmt.Sprintf("%s/api/v2/search/start", q.baseURL)
	data := url.Values{}
	data.Set("pattern", query)
	data.Set("plugins", plugins)
	data.Set("category", category)

	resp, err := q.client.PostForm(startURL, data)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("search start failed: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode search start response: %w", err)
	}
	return result.ID, nil
}

func (q *QBittorrentSearchProvider) waitForSearch(jobID int, timeout time.Duration) error {
	statusURL := fmt.Sprintf("%s/api/v2/search/status", q.baseURL)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := q.client.Get(fmt.Sprintf("%s?id=%d", statusURL, jobID))
		if err != nil {
			return err
		}

		var statuses []struct {
			ID     int    `json:"id"`
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&statuses); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()

		for _, s := range statuses {
			if s.ID == jobID {
				if s.Status == "Stopped" {
					return nil
				}
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("search timed out")
}

func (q *QBittorrentSearchProvider) getResults(jobID int) ([]models.SearchResult, error) {
	resultsURL := fmt.Sprintf("%s/api/v2/search/results", q.baseURL)
	expiresAt := time.Now().Add(24 * time.Hour)
	var allResults []models.SearchResult
	offset := 0
	limit := 100

	for {
		resp, err := q.client.Get(fmt.Sprintf("%s?id=%d&limit=%d&offset=%d", resultsURL, jobID, limit, offset))
		if err != nil {
			return nil, err
		}

		var result struct {
			Results []struct {
				FileName   string `json:"fileName"`
				FileSize   int64  `json:"fileSize"`
				FileURL    string `json:"fileUrl"`
				NbSeeders  int    `json:"nbSeeders"`
				NbLeechers int    `json:"nbLeechers"`
				SiteURL    string `json:"siteUrl"`
				DescrLink  string `json:"descrLink"`
			} `json:"results"`
			Status string `json:"status"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		for _, r := range result.Results {
			allResults = append(allResults, models.SearchResult{
				SourceID:   q.sourceID,
				Indexer:    r.SiteURL,
				Title:      r.FileName,
				MagnetLink: r.FileURL,
				TorrentURL: r.FileURL,
				Size:       r.FileSize,
				Seeders:    r.NbSeeders,
				Leechers:   r.NbLeechers,
				ExpiresAt:  expiresAt,
			})
		}

		if len(result.Results) < limit {
			break
		}
		offset += limit
	}

	return allResults, nil
}

func (q *QBittorrentSearchProvider) deleteSearch(jobID int) {
	deleteURL := fmt.Sprintf("%s/api/v2/search/delete", q.baseURL)
	data := url.Values{}
	data.Set("id", strconv.Itoa(jobID))
	resp, err := q.client.PostForm(deleteURL, data)
	if err != nil {
		log.Printf("[qBittorrent] Failed to delete search job %d: %v", jobID, err)
		return
	}
	resp.Body.Close()
}

// TestConnection attempts to login and returns nil on success.
func (q *QBittorrentSearchProvider) TestConnection() error {
	return q.login()
}

// QBittorrentPlugin represents metadata extracted from a qBittorrent search plugin file.
type QBittorrentPlugin struct {
	Name       string
	URL        string
	FullName   string
	Version    string
	Categories []string
}

// DiscoverQBittorrentEndpoints checks for qBittorrent Web UI instances and installed plugins.
// It returns candidate DownloadSource entries for any discovered endpoints.
func DiscoverQBittorrentEndpoints() []models.DownloadSource {
	var sources []models.DownloadSource

	// Try common local endpoints
	candidates := []string{
		"http://localhost:8080",
		"http://127.0.0.1:8080",
		"http://localhost:8081",
		"http://127.0.0.1:8081",
	}

	for _, endpoint := range candidates {
		if isQBittorrentReachable(endpoint) {
			sources = append(sources, models.DownloadSource{
				Name:    fmt.Sprintf("qBittorrent (%s)", endpoint),
				Type:    "qbittorrent",
				URL:     endpoint,
				Enabled: true,
			})
		}
	}

	// Also scan installed plugin files for informational defaults
	pluginDir := qBittorrentPluginDir()
	if pluginDir != "" {
		plugins := scanQBittorrentPlugins(pluginDir)
		for _, p := range plugins {
			// Only add as fallback if no Web UI endpoint was found
			if len(sources) == 0 {
				sources = append(sources, models.DownloadSource{
					Name:    p.FullName,
					Type:    "qbittorrent-plugin",
					URL:     p.URL,
					Enabled: false,
				})
			}
		}
	}

	return sources
}

func isQBittorrentReachable(endpoint string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("%s/api/v2/app/version", endpoint))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden
}

func qBittorrentPluginDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(homeDir, "AppData", "Local")
		}
		return filepath.Join(localAppData, "qBittorrent", "nova3", "engines")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "qBittorrent", "nova3", "engines")
	default:
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			configHome = filepath.Join(homeDir, ".config")
		}
		return filepath.Join(configHome, "qBittorrent", "nova3", "engines")
	}
}

func scanQBittorrentPlugins(dir string) []QBittorrentPlugin {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var plugins []QBittorrentPlugin
	nameRe := regexp.MustCompile(`(?i)^\s*name\s*=\s*['"]([^'"]+)['"]`)
	urlRe := regexp.MustCompile(`(?i)^\s*url\s*=\s*['"]([^'"]+)['"]`)
	versionRe := regexp.MustCompile(`(?i)^\s*#VERSION:\s*(\S+)`)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".py") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		plugin := QBittorrentPlugin{}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if m := nameRe.FindStringSubmatch(line); m != nil {
				plugin.FullName = m[1]
			}
			if m := urlRe.FindStringSubmatch(line); m != nil {
				plugin.URL = m[1]
			}
			if m := versionRe.FindStringSubmatch(line); m != nil {
				plugin.Version = m[1]
			}
		}

		if plugin.FullName == "" {
			plugin.FullName = strings.TrimSuffix(entry.Name(), ".py")
		}
		plugin.Name = strings.TrimSuffix(entry.Name(), ".py")

		if plugin.URL != "" {
			plugins = append(plugins, plugin)
		}
	}

	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].FullName < plugins[j].FullName
	})

	return plugins
}

// FormatBytes returns a human-readable string for a byte count.
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
