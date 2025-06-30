package main

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestMainDotArgument(t *testing.T) {
	os.Args = []string{"media-manager", "."}
	logOutput := captureLogOutput(func() { main() })
	if logOutput != "Opening directory: .\n" {
		t.Errorf("Expected log output 'Opening directory: .', got '%s'", logOutput)
	}
}

func TestMainNoArgument(t *testing.T) {
	os.Args = []string{"media-manager"}
	logOutput := captureLogOutput(func() { main() })
	if logOutput == "" || logOutput[:19] != "Opening directory: " {
		t.Errorf("Expected log output to start with 'Opening directory: ', got '%s'", logOutput)
	}
}

func TestMainPathArgument(t *testing.T) {
	os.Args = []string{"media-manager", "/tmp"}
	logOutput := captureLogOutput(func() { main() })
	if logOutput != "Opening directory: /tmp\n" {
		t.Errorf("Expected log output 'Opening directory: /tmp', got '%s'", logOutput)
	}
}

func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)
	f()
	return buf.String()
}
