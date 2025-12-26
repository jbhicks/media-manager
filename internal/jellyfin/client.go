package jellyfin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

type LibraryRefreshResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// RefreshLibrary triggers a library scan in Jellyfin
func (c *Client) RefreshLibrary() error {
	if c.apiKey == "" {
		log.Println("[JELLYFIN] API key not configured, skipping library refresh")
		return nil
	}

	url := fmt.Sprintf("%s/Library/Refresh", c.baseURL)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MediaBrowser-Token", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh library: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jellyfin returned status %d", resp.StatusCode)
	}

	log.Println("[JELLYFIN] Library refresh triggered successfully")
	return nil
}

// GetSystemInfo retrieves Jellyfin system information (for testing connection)
func (c *Client) GetSystemInfo() (map[string]any, error) {
	url := fmt.Sprintf("%s/System/Info", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MediaBrowser-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyfin returned status %d", resp.StatusCode)
	}

	var info map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return info, nil
}

// CreateApiKey creates an API key for the media manager
// Note: This requires authentication with a user account first
func (c *Client) CreateApiKey(appName string) (string, error) {
	url := fmt.Sprintf("%s/Auth/Keys", c.baseURL)

	body := map[string]string{
		"App": appName,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MediaBrowser-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create API key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return "", fmt.Errorf("jellyfin returned status %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"AccessToken"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.AccessToken, nil
}

// TestConnection tests if Jellyfin is reachable
func (c *Client) TestConnection() error {
	url := fmt.Sprintf("%s/System/Info/Public", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to Jellyfin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jellyfin returned status %d", resp.StatusCode)
	}

	log.Println("[JELLYFIN] Connection test successful")
	return nil
}

// JellyfinItem represents a media item from Jellyfin
type JellyfinItem struct {
	Name              string                `json:"Name"`
	Id                string                `json:"Id"`
	Type              string                `json:"Type"`
	ProductionYear    int                   `json:"ProductionYear,omitempty"`
	ImageTags         map[string]string     `json:"ImageTags,omitempty"`
	BackdropImageTags []string              `json:"BackdropImageTags,omitempty"`
	Overview          string                `json:"Overview,omitempty"`
	CommunityRating   float64               `json:"CommunityRating,omitempty"`
	OfficialRating    string                `json:"OfficialRating,omitempty"`
	RunTimeTicks      int64                 `json:"RunTimeTicks,omitempty"`
	Path              string                `json:"Path,omitempty"`
	MediaSources      []JellyfinMediaSource `json:"MediaSources,omitempty"`
}

type JellyfinMediaSource struct {
	Size int64 `json:"Size,omitempty"`
}

type JellyfinItemsResponse struct {
	Items            []JellyfinItem `json:"Items"`
	TotalRecordCount int            `json:"TotalRecordCount"`
}

// GetLibraryItems retrieves all items from Jellyfin libraries
func (c *Client) GetLibraryItems(itemType string) ([]JellyfinItem, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("jellyfin API key not configured")
	}

	// Get the first user ID (needed for API calls)
	userID, err := c.getUserID()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	url := fmt.Sprintf("%s/Users/%s/Items?IncludeItemTypes=%s&Recursive=true&Fields=Path,MediaSources,Overview,CommunityRating,OfficialRating",
		c.baseURL, userID, itemType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MediaBrowser-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get library items: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyfin returned status %d", resp.StatusCode)
	}

	var itemsResp JellyfinItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	log.Printf("[JELLYFIN] Retrieved %d %s items from library", len(itemsResp.Items), itemType)
	return itemsResp.Items, nil
}

// getUserID gets the first available user ID
func (c *Client) getUserID() (string, error) {
	url := fmt.Sprintf("%s/Users", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MediaBrowser-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get users: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("jellyfin returned status %d", resp.StatusCode)
	}

	var users []struct {
		Id   string `json:"Id"`
		Name string `json:"Name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(users) == 0 {
		return "", fmt.Errorf("no users found in Jellyfin")
	}

	return users[0].Id, nil
}

// GetImageURL returns the URL for an item's primary image
func (c *Client) GetImageURL(itemID, imageTag string) string {
	if imageTag == "" {
		return ""
	}
	return fmt.Sprintf("%s/Items/%s/Images/Primary?tag=%s", c.baseURL, itemID, imageTag)
}

// RefreshItemMetadata triggers metadata refresh for a specific item
func (c *Client) RefreshItemMetadata(itemID string, replaceAllImages bool) error {
	if c.apiKey == "" {
		return fmt.Errorf("jellyfin API key not configured")
	}

	url := fmt.Sprintf("%s/Items/%s/Refresh", c.baseURL, itemID)

	// Build query parameters
	params := "?Recursive=true&MetadataRefreshMode=FullRefresh&ImageRefreshMode=FullRefresh"
	if replaceAllImages {
		params += "&ReplaceAllImages=true"
	} else {
		params += "&ReplaceAllImages=false"
	}

	url += params

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MediaBrowser-Token", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to refresh item metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jellyfin returned status %d", resp.StatusCode)
	}

	log.Printf("[JELLYFIN] Metadata refresh triggered for item %s", itemID)
	return nil
}

// SearchByPath finds a Jellyfin item by its file path
func (c *Client) SearchByPath(filePath string) (*JellyfinItem, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("jellyfin API key not configured")
	}

	userID, err := c.getUserID()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	url := fmt.Sprintf("%s/Users/%s/Items?Recursive=true&Fields=Path&Limit=1000", c.baseURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MediaBrowser-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search items: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jellyfin returned status %d", resp.StatusCode)
	}

	var itemsResp JellyfinItemsResponse
	if err := json.NewDecoder(resp.Body).Decode(&itemsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Find item matching the path
	for _, item := range itemsResp.Items {
		if item.Path == filePath {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("no item found with path: %s", filePath)
}
