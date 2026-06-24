package torrent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SkYNewZ/go-flaresolverr"
	"github.com/google/uuid"
)

// FlareSolverrTransport is an HTTP transport that falls back to FlareSolverr on Cloudflare challenges
type FlareSolverrTransport struct {
	Base         http.RoundTripper
	FlareSolverr flaresolverr.Client
}

// RoundTrip implements http.RoundTripper
func (t *FlareSolverrTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Try normal request first
	resp, err := t.Base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// Check if we got a Cloudflare challenge
	if isCloudflareChallenge(resp) {
		log.Printf("[FlareSolverr] Detected Cloudflare challenge for %s, using fallback", req.URL)
		resp.Body.Close()

		if t.FlareSolverr == nil {
			return nil, fmt.Errorf("cloudflare challenge detected but flaresolverr not configured")
		}

		// Use FlareSolverr to bypass
		flareResp, err := t.FlareSolverr.Get(req.Context(), req.URL.String(), uuid.Nil, nil)
		if err != nil {
			return nil, fmt.Errorf("flaresolverr fallback failed: %w", err)
		}

		// Create a synthetic response
		statusCode := 200
		responseBody := ""
		if flareResp.Solution != nil {
			statusCode = flareResp.Solution.Status
			responseBody = flareResp.Solution.Response
		}
		return &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Header:     http.Header{},
			Request:    req,
		}, nil
	}

	return resp, nil
}

// isCloudflareChallenge checks if the response is a Cloudflare challenge page
func isCloudflareChallenge(resp *http.Response) bool {
	if resp.StatusCode == 403 {
		return true
	}

	// Check for Cloudflare headers
	if resp.Header.Get("CF-RAY") != "" {
		return true
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		// Read a bit of the body to check for challenge
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		resp.Body = io.NopCloser(io.MultiReader(strings.NewReader(string(body)), resp.Body))

		bodyStr := string(body)
		if strings.Contains(bodyStr, "cf-browser-verification") ||
			strings.Contains(bodyStr, "challenge-platform") ||
			strings.Contains(bodyStr, "Just a moment") {
			return true
		}
	}

	return false
}

// NewFlareSolverrHTTPClient creates an HTTP client with FlareSolverr fallback
func NewFlareSolverrHTTPClient(flareSolverrURL string, timeout time.Duration) *http.Client {
	baseTransport := http.DefaultTransport

	if flareSolverrURL == "" {
		return &http.Client{
			Transport: baseTransport,
			Timeout:   timeout,
		}
	}

	flareClient := flaresolverr.New(flareSolverrURL, timeout, nil)

	transport := &FlareSolverrTransport{
		Base:         baseTransport,
		FlareSolverr: flareClient,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
}
