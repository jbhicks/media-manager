package torrent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/SkYNewZ/go-flaresolverr"
	"github.com/google/uuid"
)

// FlareSolverrClient wraps the go-flaresolverr library
type FlareSolverrClient struct {
	client  flaresolverr.Client
	baseURL string
	enabled bool
	timeout time.Duration
}

// NewFlareSolverrClient creates a new FlareSolverr client
func NewFlareSolverrClient(baseURL string, timeout time.Duration) *FlareSolverrClient {
	if baseURL == "" {
		return &FlareSolverrClient{enabled: false}
	}

	client := flaresolverr.New(baseURL, timeout, nil)
	
	return &FlareSolverrClient{
		client:  client,
		baseURL: baseURL,
		enabled: true,
		timeout: timeout,
	}
}

// IsEnabled returns whether FlareSolverr is configured
func (f *FlareSolverrClient) IsEnabled() bool {
	return f.enabled && f.client != nil
}

// Get performs a GET request through FlareSolverr
func (f *FlareSolverrClient) Get(ctx context.Context, url string) (*flaresolverr.Response, error) {
	if !f.IsEnabled() {
		return nil, fmt.Errorf("flaresolverr not configured")
	}

	log.Printf("[FlareSolverr] Bypassing Cloudflare for: %s", url)
	
	resp, err := f.client.Get(ctx, url, uuid.Nil, nil)
	if err != nil {
		return nil, fmt.Errorf("flaresolverr request failed: %w", err)
	}

	log.Printf("[FlareSolverr] Successfully bypassed Cloudflare for: %s (status: %s)", url, resp.Status)
	return resp, nil
}

// Post performs a POST request through FlareSolverr
func (f *FlareSolverrClient) Post(ctx context.Context, url string, postData string) (*flaresolverr.Response, error) {
	if !f.IsEnabled() {
		return nil, fmt.Errorf("flaresolverr not configured")
	}

	log.Printf("[FlareSolverr] POST request through proxy: %s", url)
	
	resp, err := f.client.Post(ctx, url, uuid.Nil, postData, nil)
	if err != nil {
		return nil, fmt.Errorf("flaresolverr POST request failed: %w", err)
	}

	return resp, nil
}

// CreateSession creates a new browser session
func (f *FlareSolverrClient) CreateSession() (string, error) {
	if !f.IsEnabled() {
		return "", fmt.Errorf("flaresolverr not configured")
	}

	sessionID := uuid.New()
	_, err := f.client.CreateSession(context.Background(), sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	log.Printf("[FlareSolverr] Created session: %s", sessionID)
	return sessionID.String(), nil
}

// DestroySession destroys a browser session
func (f *FlareSolverrClient) DestroySession(sessionID string) error {
	if !f.IsEnabled() {
		return nil
	}

	id, err := uuid.Parse(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	err = f.client.DestroySession(context.Background(), id)
	if err != nil {
		return fmt.Errorf("failed to destroy session: %w", err)
	}

	log.Printf("[FlareSolverr] Destroyed session: %s", sessionID)
	return nil
}

// HealthCheck checks if FlareSolverr is reachable
func (f *FlareSolverrClient) HealthCheck() error {
	if !f.IsEnabled() {
		return fmt.Errorf("flaresolverr not configured")
	}

	// Try to list sessions (lightweight operation)
	_, err := f.client.ListSessions(context.Background())
	if err != nil {
		return fmt.Errorf("flaresolverr health check failed: %w", err)
	}

	return nil
}