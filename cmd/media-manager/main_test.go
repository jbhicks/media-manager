package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/user/media-manager/internal/config"
)

// setTestHomeDir sets the home directory for tests, handling both Unix (HOME)
// and Windows (USERPROFILE) environment variables
func setTestHomeDir(t *testing.T, tempDir string) func() {
	t.Helper()
	oldHome := os.Getenv("HOME")
	oldUserProfile := os.Getenv("USERPROFILE")

	os.Setenv("HOME", tempDir)
	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", tempDir)
	}

	return func() {
		os.Setenv("HOME", oldHome)
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", oldUserProfile)
		}
	}
}

func TestMainDotArgument(t *testing.T) {
	os.Args = []string{"media-manager", "."}
	logOutput := captureLogOutput(func() { run(nil) })
	if !strings.Contains(logOutput, "Opening directory: .") {
		t.Errorf("Expected log output to contain 'Opening directory: .', got '%s'", logOutput)
	}
}

func TestMainNoArgument(t *testing.T) {
	// Use isolated temp directory for config
	tempDir := t.TempDir()
	cleanup := setTestHomeDir(t, tempDir)
	defer cleanup()

	os.Args = []string{"media-manager"}
	logOutput := captureLogOutput(func() { run(nil) })
	// When no config exists and no argument, it should use current directory
	if !strings.Contains(logOutput, "Opening directory: ") {
		t.Errorf("Expected log output to contain 'Opening directory: ', got '%s'", logOutput)
	}
}

func TestMainPathArgument(t *testing.T) {
	testPath := t.TempDir() // Use actual temp dir instead of /tmp for Windows compat
	os.Args = []string{"media-manager", testPath}
	logOutput := captureLogOutput(func() { run(nil) })
	if !strings.Contains(logOutput, "Opening directory: "+testPath) {
		t.Errorf("Expected log output to contain 'Opening directory: %s', got '%s'", testPath, logOutput)
	}
}

func TestMainFileArgumentUsesParentAndLaunchFile(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "sample.mp4")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	os.Args = []string{"media-manager", testFile}

	var gotDir string
	var gotLaunchFiles []string
	run(func(dir string, launchFiles []string) {
		gotDir = dir
		gotLaunchFiles = launchFiles
	})

	if gotDir != testDir {
		t.Fatalf("Expected parent dir %s, got %s", testDir, gotDir)
	}
	if len(gotLaunchFiles) != 1 || gotLaunchFiles[0] != testFile {
		t.Fatalf("Expected launch files [%s], got %v", testFile, gotLaunchFiles)
	}
}

func TestMainOpenParentArgumentDoesNotLaunchFileFocus(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "sample.mp4")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	os.Args = []string{"media-manager", "--open-parent", testFile}

	var gotDir string
	var gotLaunchFiles []string
	run(func(dir string, launchFiles []string) {
		gotDir = dir
		gotLaunchFiles = launchFiles
	})

	if gotDir != testDir {
		t.Fatalf("Expected parent dir %s, got %s", testDir, gotDir)
	}
	if len(gotLaunchFiles) != 0 {
		t.Fatalf("Expected no launch focus files, got %v", gotLaunchFiles)
	}
}

func TestMainMultiFileSelectionSameDirectory(t *testing.T) {
	testDir := t.TempDir()
	fileA := filepath.Join(testDir, "a.mp4")
	fileB := filepath.Join(testDir, "b.mp4")
	if err := os.WriteFile(fileA, []byte("a"), 0644); err != nil {
		t.Fatalf("Failed to create fileA: %v", err)
	}
	if err := os.WriteFile(fileB, []byte("b"), 0644); err != nil {
		t.Fatalf("Failed to create fileB: %v", err)
	}

	os.Args = []string{"media-manager", fileA, fileB}

	var gotDir string
	var gotLaunchFiles []string
	run(func(dir string, launchFiles []string) {
		gotDir = dir
		gotLaunchFiles = launchFiles
	})

	if gotDir != testDir {
		t.Fatalf("Expected selected dir %s, got %s", testDir, gotDir)
	}
	if len(gotLaunchFiles) != 2 {
		t.Fatalf("Expected 2 launch files, got %v", gotLaunchFiles)
	}
}

// TestLastFolderRemembered verifies that the last opened folder is saved and loaded on restart
func TestLastFolderRemembered(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()
	cleanup := setTestHomeDir(t, tempDir)
	defer cleanup()

	testFolder1 := filepath.Join(tempDir, "test-media-1")
	testFolder2 := filepath.Join(tempDir, "test-media-2")

	if err := os.MkdirAll(testFolder1, 0755); err != nil {
		t.Fatalf("Failed to create test folder 1: %v", err)
	}
	if err := os.MkdirAll(testFolder2, 0755); err != nil {
		t.Fatalf("Failed to create test folder 2: %v", err)
	}

	// Test 1: Open first folder explicitly
	os.Args = []string{"media-manager", testFolder1}
	logOutput := captureLogOutput(func() { run(nil) })

	if !strings.Contains(logOutput, "Opening directory: "+testFolder1) {
		t.Errorf("Expected to open %s, got: %s", testFolder1, logOutput)
	}

	// Simulate saving config with the folder
	cfg, err := config.LoadConfig(testFolder1)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test 2: Open second folder explicitly
	os.Args = []string{"media-manager", testFolder2}
	logOutput = captureLogOutput(func() { run(nil) })

	if !strings.Contains(logOutput, "Opening directory: "+testFolder2) {
		t.Errorf("Expected to open %s, got: %s", testFolder2, logOutput)
	}

	// Save config with second folder
	cfg, err = config.LoadConfig(testFolder2)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test 3: Start without arguments - should open last folder (testFolder2)
	os.Args = []string{"media-manager"}
	logOutput = captureLogOutput(func() { run(nil) })

	if !strings.Contains(logOutput, "Opening last directory from config: "+testFolder2) {
		t.Errorf("Expected to open last directory %s, got: %s", testFolder2, logOutput)
	}

	// Test 4: Verify explicit argument overrides remembered folder
	os.Args = []string{"media-manager", testFolder1}
	logOutput = captureLogOutput(func() { run(nil) })

	if !strings.Contains(logOutput, "Opening directory: "+testFolder1) {
		t.Errorf("Expected explicit directory %s to override remembered folder, got: %s", testFolder1, logOutput)
	}
}

// TestConfigPersistsMediaDirs verifies that MediaDirs are correctly saved and loaded
func TestConfigPersistsMediaDirs(t *testing.T) {
	tempDir := t.TempDir()
	cleanup := setTestHomeDir(t, tempDir)
	defer cleanup()

	testFolder := filepath.Join(tempDir, "test-media")
	if err := os.MkdirAll(testFolder, 0755); err != nil {
		t.Fatalf("Failed to create test folder: %v", err)
	}

	// Create and save config with a media directory
	cfg, err := config.LoadConfig(testFolder)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	if len(cfg.MediaDirs) == 0 || cfg.MediaDirs[0] != testFolder {
		t.Errorf("Expected MediaDirs to contain %s, got: %v", testFolder, cfg.MediaDirs)
	}

	// Save config
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config without specifying a directory
	cfg2, err := config.LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if len(cfg2.MediaDirs) == 0 || cfg2.MediaDirs[0] != testFolder {
		t.Errorf("Expected loaded config to have MediaDirs %s, got: %v", testFolder, cfg2.MediaDirs)
	}
}

// TestNoConfigFileUsesDefault verifies behavior when config file doesn't exist
func TestNoConfigFileUsesDefault(t *testing.T) {
	tempDir := t.TempDir()
	cleanup := setTestHomeDir(t, tempDir)
	defer cleanup()

	// With no config file and no arguments, should use current directory
	os.Args = []string{"media-manager"}
	cwd, _ := os.Getwd()

	logOutput := captureLogOutput(func() { run(nil) })

	if !strings.Contains(logOutput, "Opening directory: "+cwd) {
		t.Errorf("Expected to use current directory %s when no config exists, got: %s", cwd, logOutput)
	}
}

func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	f()
	return buf.String()
}
