# Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run

# Project variables
BINARY_NAME=media-manager
CMD_PATH=./cmd/media-manager

.PHONY: all dev build clean clear-cache

all: dev

dev:
	mkdir -p tmp
	air

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(CMD_PATH)/main.go

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)

clear-cache:
	@echo "Clearing thumbnail cache..."
	@rm -rf ~/.media-manager/thumbnails/* ./thumbnails/* 2>/dev/null || true
	@echo "Thumbnail cache cleared!"
