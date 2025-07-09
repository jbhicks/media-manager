package main

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestMainDotArgument(t *testing.T) {
	os.Args = []string{"media-manager", "."}
	logOutput := captureLogOutput(func() { run(nil) })
	if !strings.Contains(logOutput, "Opening directory: .") {
		t.Errorf("Expected log output to contain 'Opening directory: .', got '%s'", logOutput)
	}
}

func TestMainNoArgument(t *testing.T) {
	os.Args = []string{"media-manager"}
	logOutput := captureLogOutput(func() { run(nil) })
	if !strings.Contains(logOutput, "Opening directory: ") {
		t.Errorf("Expected log output to contain 'Opening directory: ', got '%s'", logOutput)
	}
}

func TestMainPathArgument(t *testing.T) {
	os.Args = []string{"media-manager", "/tmp"}
	logOutput := captureLogOutput(func() { run(nil) })
	if !strings.Contains(logOutput, "Opening directory: /tmp") {
		t.Errorf("Expected log output to contain 'Opening directory: /tmp', got '%s'", logOutput)
	}
}


func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	f()
	return buf.String()
}