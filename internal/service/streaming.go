package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// StreamHandler handles video streaming with HLS transcoding
type StreamHandler struct {
	mediaDir    string
	hlsDir      string
	ffmpegPath  string
	activeJobs  map[string]*TranscodeJob
}

// TranscodeJob represents an active transcoding job
type TranscodeJob struct {
	VideoPath   string    `json:"video_path"`
	HLSDir      string    `json:"hls_dir"`
	Status      string    `json:"status"` // "preparing", "transcoding", "ready", "error"
	Progress    float64   `json:"progress"`
	CreatedAt   time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(mediaDir string) *StreamHandler {
	hlsDir := filepath.Join(os.TempDir(), "media-manager-hls")
	os.MkdirAll(hlsDir, 0755)
	
	return &StreamHandler{
		mediaDir:   mediaDir,
		hlsDir:     hlsDir,
		ffmpegPath: "ffmpeg",
		activeJobs: make(map[string]*TranscodeJob),
	}
}

// findVideoFile finds the video file in a movie directory
func (sh *StreamHandler) findVideoFile(moviePath string) (string, error) {
	// If it's already a video file, return it
	ext := strings.ToLower(filepath.Ext(moviePath))
	videoExts := map[string]bool{".mp4": true, ".mkv": true, ".avi": true, ".mov": true, ".m4v": true, ".wmv": true, ".webm": true}
	if videoExts[ext] {
		return moviePath, nil
	}
	
	// Otherwise, look for video files in the directory
	info, err := os.Stat(moviePath)
	if err != nil {
		return "", err
	}
	
	var searchDir string
	if info.IsDir() {
		searchDir = moviePath
	} else {
		searchDir = filepath.Dir(moviePath)
	}
	
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return "", err
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if videoExts[ext] {
			return filepath.Join(searchDir, name), nil
		}
	}
	
	return "", fmt.Errorf("no video file found in %s", searchDir)
}

// getJobID generates a unique job ID from a video path
func getJobID(videoPath string) string {
	// Use the directory name as the job ID for consistency
	return filepath.Base(filepath.Dir(videoPath))
}

// StartTranscode starts an HLS transcode job for a video
func (sh *StreamHandler) StartTranscode(videoPath string) (*TranscodeJob, error) {
	jobID := getJobID(videoPath)
	
	// Check if job already exists and is ready
	if job, exists := sh.activeJobs[jobID]; exists {
		job.LastAccessed = time.Now()
		if job.Status == "ready" {
			return job, nil
		}
		if job.Status == "transcoding" || job.Status == "preparing" {
			return job, nil
		}
	}
	
	// Create HLS output directory
	hlsOutputDir := filepath.Join(sh.hlsDir, jobID)
	os.MkdirAll(hlsOutputDir, 0755)
	
	job := &TranscodeJob{
		VideoPath:    videoPath,
		HLSDir:       hlsOutputDir,
		Status:       "preparing",
		Progress:     0,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}
	sh.activeJobs[jobID] = job
	
	// Start transcoding in background
	go sh.transcode(job)
	
	return job, nil
}

// transcode runs the FFmpeg HLS transcoding
func (sh *StreamHandler) transcode(job *TranscodeJob) {
	job.Status = "transcoding"
	
	playlistPath := filepath.Join(job.HLSDir, "playlist.m3u8")
	segmentPattern := filepath.Join(job.HLSDir, "segment_%03d.ts")
	
	// FFmpeg command for HLS streaming
	// Using multiple quality levels for adaptive streaming
	cmd := exec.Command(sh.ffmpegPath,
		"-i", job.VideoPath,
		"-c:v", "libx264",           // Video codec
		"-c:a", "aac",               // Audio codec
		"-b:v", "2000k",             // Video bitrate
		"-b:a", "128k",              // Audio bitrate
		"-maxrate", "2500k",         // Max bitrate
		"-bufsize", "4000k",         // Buffer size
		"-vf", "scale=1920:-2",      // Scale to 1080p max, maintain aspect ratio
		"-preset", "fast",           // Encoding preset (balance speed/quality)
		"-crf", "23",                // Quality level
		"-hls_time", "10",           // Segment duration (seconds)
		"-hls_list_size", "0",       // Keep all segments
		"-hls_segment_filename", segmentPattern,
		"-f", "hls",
		"-y",                        // Overwrite output
		playlistPath,
	)
	
	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[STREAM] Transcode failed for %s: %v\nOutput: %s", job.VideoPath, err, string(output))
		job.Status = "error"
		return
	}
	
	job.Status = "ready"
	job.Progress = 100
	log.Printf("[STREAM] Transcode complete for %s", job.VideoPath)
}

// GetPlaylistPath returns the path to the HLS playlist
func (sh *StreamHandler) GetPlaylistPath(jobID string) string {
	return filepath.Join(sh.hlsDir, jobID, "playlist.m3u8")
}

// GetSegmentPath returns the path to an HLS segment
func (sh *StreamHandler) GetSegmentPath(jobID, segmentName string) string {
	return filepath.Join(sh.hlsDir, jobID, segmentName)
}

// GetJob returns a transcode job by ID
func (sh *StreamHandler) GetJob(jobID string) (*TranscodeJob, bool) {
	job, exists := sh.activeJobs[jobID]
	if exists {
		job.LastAccessed = time.Now()
	}
	return job, exists
}

// CleanupOldJobs removes transcode jobs that haven't been accessed recently
func (sh *StreamHandler) CleanupOldJobs(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)
	for id, job := range sh.activeJobs {
		if job.LastAccessed.Before(cutoff) && job.Status != "transcoding" {
			// Remove HLS files
			os.RemoveAll(job.HLSDir)
			delete(sh.activeJobs, id)
			log.Printf("[STREAM] Cleaned up old job: %s", id)
		}
	}
}

// HTTP Handlers

// HandleStreamInit initializes a stream for a video file
func (sh *StreamHandler) HandleStreamInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Get movie path from request body
	type StreamRequest struct {
		Path string `json:"path"`
	}
	
	var req StreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	if req.Path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}
	
	// Find the actual video file
	videoPath, err := sh.findVideoFile(req.Path)
	if err != nil {
		log.Printf("[STREAM] Video not found: %s - %v", req.Path, err)
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}
	
	// Start transcode
	job, err := sh.StartTranscode(videoPath)
	if err != nil {
		log.Printf("[STREAM] Failed to start transcode: %v", err)
		http.Error(w, "Failed to start stream", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":   getJobID(videoPath),
		"status":   job.Status,
		"progress": job.Progress,
	})
}

// HandleStreamPlaylist serves the HLS playlist (m3u8)
func (sh *StreamHandler) HandleStreamPlaylist(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job")
	if jobID == "" {
		http.Error(w, "Missing job parameter", http.StatusBadRequest)
		return
	}
	
	job, exists := sh.GetJob(jobID)
	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}
	
	if job.Status == "error" {
		http.Error(w, "Transcode failed", http.StatusInternalServerError)
		return
	}
	
	if job.Status != "ready" {
		// Return a "not ready" response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   job.Status,
			"progress": job.Progress,
		})
		return
	}
	
	playlistPath := sh.GetPlaylistPath(jobID)
	
	// Check if playlist exists
	if _, err := os.Stat(playlistPath); os.IsNotExist(err) {
		// Wait a bit and check again (race condition with transcode)
		time.Sleep(500 * time.Millisecond)
		if _, err := os.Stat(playlistPath); os.IsNotExist(err) {
			http.Error(w, "Playlist not ready", http.StatusNotFound)
			return
		}
	}
	
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, playlistPath)
}

// HandleStreamSegment serves HLS segments (.ts files)
func (sh *StreamHandler) HandleStreamSegment(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job")
	segment := r.URL.Query().Get("segment")
	
	if jobID == "" || segment == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}
	
	// Security: prevent directory traversal
	if strings.Contains(segment, "..") || strings.Contains(segment, "/") {
		http.Error(w, "Invalid segment name", http.StatusBadRequest)
		return
	}
	
	segmentPath := sh.GetSegmentPath(jobID, segment)
	
	// Verify the segment is within the HLS directory
	absSegment, _ := filepath.Abs(segmentPath)
	absHLS, _ := filepath.Abs(sh.hlsDir)
	if !strings.HasPrefix(absSegment, absHLS) {
		http.Error(w, "Invalid segment path", http.StatusForbidden)
		return
	}
	
	if _, err := os.Stat(segmentPath); os.IsNotExist(err) {
		http.Error(w, "Segment not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, segmentPath)
}

// HandleStreamStatus returns the status of a stream job
func (sh *StreamHandler) HandleStreamStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("job")
	if jobID == "" {
		http.Error(w, "Missing job parameter", http.StatusBadRequest)
		return
	}
	
	job, exists := sh.GetJob(jobID)
	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":   jobID,
		"status":   job.Status,
		"progress": job.Progress,
		"video_path": job.VideoPath,
	})
}

// HandleDirectStream serves a video file directly (no transcoding)
func (sh *StreamHandler) HandleDirectStream(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}
	
	// Security: verify path is within media directory
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	absMedia, _ := filepath.Abs(sh.mediaDir)
	if !strings.HasPrefix(absPath, absMedia) {
		http.Error(w, "Path outside media directory", http.StatusForbidden)
		return
	}
	
	videoPath, err := sh.findVideoFile(path)
	if err != nil {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}
	
	// Serve the file with proper headers for streaming
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	
	file, err := os.Open(videoPath)
	if err != nil {
		http.Error(w, "Failed to open video", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	
	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "Failed to stat video", http.StatusInternalServerError)
		return
	}
	
	fileSize := stat.Size()
	
	// Handle range requests for seeking
	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" {
		// Parse range header
		var start, end int64
		fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
		if end == 0 || end >= fileSize {
			end = fileSize - 1
		}
		
		w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
		w.WriteHeader(http.StatusPartialContent)
		
		file.Seek(start, io.SeekStart)
		io.CopyN(w, file, end-start+1)
		return
	}
	
	w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
	io.Copy(w, file)
}

// HandleStreamStop stops and cleans up a stream
func (sh *StreamHandler) HandleStreamStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	jobID := r.URL.Query().Get("job")
	if jobID == "" {
		http.Error(w, "Missing job parameter", http.StatusBadRequest)
		return
	}
	
	job, exists := sh.GetJob(jobID)
	if !exists {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}
	
	// Clean up HLS files
	os.RemoveAll(job.HLSDir)
	delete(sh.activeJobs, jobID)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Stream stopped",
		"job_id": jobID,
	})
}
