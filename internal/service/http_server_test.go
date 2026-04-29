package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/pkg/models"
)

// setupTestServer creates a test HTTP server with in-memory database
func setupTestServer(t *testing.T) (*HTTPServer, *db.Database, func()) {
	dbPath := "test_http_server.db"
	database, err := db.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cfg := &models.ServiceConfig{
		DownloadPath:      "/tmp/downloads",
		LibraryPath:       "/tmp/library",
		TorrentClientType: "native",
	}

	dm, err := NewDownloadManager(database, cfg)
	if err != nil {
		t.Fatalf("Failed to create download manager: %v", err)
	}

	server := NewHTTPServer(":0", database, dm)

	cleanup := func() {
		database.Close()
		os.Remove(dbPath)
	}

	return server, database, cleanup
}

func TestHandleTasks(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/tasks", nil)
	w := httptest.NewRecorder()

	server.handleTasks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("Expected JSON content type, got %s", contentType)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify response structure - must have 'tasks' and 'stats' keys
	tasks, ok := response["tasks"]
	if !ok {
		t.Fatal("Response missing 'tasks' key - frontend expects { tasks: [...], stats: {...} }")
	}

	if _, ok := tasks.([]interface{}); !ok {
		t.Fatalf("Expected tasks to be array, got %T", tasks)
	}

	if _, ok := response["stats"]; !ok {
		t.Fatal("Response missing 'stats' key")
	}
}

func TestHandleStats(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/stats", nil)
	w := httptest.NewRecorder()

	server.handleStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify has downloads key
	if _, ok := response["downloads"]; !ok {
		t.Fatal("Response missing 'downloads' key")
	}

	// Verify has suggestions key
	if _, ok := response["suggestions"]; !ok {
		t.Fatal("Response missing 'suggestions' key")
	}
}

func TestHandleVPNStatus(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/vpn/status", nil)
	w := httptest.NewRecorder()

	server.handleVPNStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify response structure
	if _, ok := response["active"]; !ok {
		t.Fatal("Response missing 'active' key")
	}

	if _, ok := response["provider"]; !ok {
		t.Fatal("Response missing 'provider' key")
	}

	if _, ok := response["status"]; !ok {
		t.Fatal("Response missing 'status' key")
	}
}

func TestHandleSources(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/sources", nil)
	w := httptest.NewRecorder()

	server.handleSources(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have sources key
	if _, ok := response["sources"]; !ok {
		t.Fatal("Response missing 'sources' key")
	}
}

func TestHandleRules(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/rules", nil)
	w := httptest.NewRecorder()

	server.handleRules(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have rules key
	if _, ok := response["rules"]; !ok {
		t.Fatal("Response missing 'rules' key")
	}
}

func TestHandleTasksResponseStructure(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/tasks", nil)
	w := httptest.NewRecorder()

	server.handleTasks(w, req)

	body := w.Body.String()

	// Verify it's valid JSON and has expected structure
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		t.Fatalf("Response is not valid JSON: %v\nBody: %s", err, body)
	}

	// The frontend expects: { tasks: [...], stats: {...} }
	// Make sure it's NOT just an array
	var arrayResponse []interface{}
	if err := json.Unmarshal([]byte(body), &arrayResponse); err == nil {
		t.Fatal("API returned array instead of object with 'tasks' and 'stats' keys. Frontend expects { tasks: [...], stats: {...} }")
	}
}

func TestHandleCancelTaskMethod(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	// GET should not be allowed
	req := httptest.NewRequest("GET", "/api/tasks/cancel", nil)
	w := httptest.NewRecorder()

	server.handleCancelTask(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status 405 for GET request, got %d", w.Code)
	}
}

func TestHandleLibraryMovies(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/library/movies", nil)
	w := httptest.NewRecorder()

	server.handleLibraryMovies(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have movies array
	if _, ok := response["movies"]; !ok {
		t.Fatal("Response missing 'movies' key - frontend expects { movies: [...], count: N, dataSource: '...' }")
	}

	// Should have count
	if _, ok := response["count"]; !ok {
		t.Fatal("Response missing 'count' key")
	}

	// Should have dataSource
	if _, ok := response["dataSource"]; !ok {
		t.Fatal("Response missing 'dataSource' key")
	}
}

func TestHandleTags(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/library/tags", nil)
	w := httptest.NewRecorder()

	server.handleTags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if _, ok := response["tags"]; !ok {
		t.Fatal("Response missing 'tags' key")
	}
}

func TestHandleReprocessLibraryMethod(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	// GET should not be allowed
	req := httptest.NewRequest("GET", "/api/library/reprocess", nil)
	w := httptest.NewRecorder()

	server.handleReprocessLibrary(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status 405 for GET request, got %d", w.Code)
	}
}

func TestHandleDeleteMovieMethod(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	// GET should not be allowed
	req := httptest.NewRequest("GET", "/api/library/delete", nil)
	w := httptest.NewRecorder()

	server.handleDeleteMovie(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status 405 for GET request, got %d", w.Code)
	}
}

func TestHandleCreateTag(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	// POST should create a tag
	req := httptest.NewRequest("POST", "/api/library/tag/create", strings.NewReader(`{"name":"Action","color":"#ff0000"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleCreateTag(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["success"] != true {
		t.Fatalf("Expected success=true, got %v", response["success"])
	}

	// GET should not be allowed
	req2 := httptest.NewRequest("GET", "/api/library/tag/create", nil)
	w2 := httptest.NewRecorder()
	server.handleCreateTag(w2, req2)

	if w2.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Expected status 405 for GET request, got %d", w2.Code)
	}
}

func TestHandleHealth(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Fatalf("Expected status=healthy, got %v", response["status"])
	}
}

func TestHandleSearchResponseStructure(t *testing.T) {
	server, _, cleanup := setupTestServer(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/search?status=pending", nil)
	w := httptest.NewRecorder()

	server.handleSearch(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should have results array
	if _, ok := response["results"]; !ok {
		t.Fatal("Response missing 'results' key")
	}

	// Should have query info
	if _, ok := response["query"]; !ok {
		t.Fatal("Response missing 'query' key")
	}
}
