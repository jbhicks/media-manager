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

.PHONY: all dev build clean clear-cache test

all: dev

dev:
	$(GOBUILD) -o bin/clear-previews ./cmd/clear-previews/main.go
	bin/clear-previews
	mkdir -p tmp
	CLEAR_DB_ON_START=true air

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(CMD_PATH)/main.go

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)

clear-cache:
	@echo "Clearing all media-manager cache..."
	@rm -rf ~/.media-manager/thumbnails/* ~/.media-manager/previews/* ~/.media-manager/video_previews/* ./thumbnails/* 2>/dev/null || true
	@echo "All media-manager cache cleared!"

test:
	$(GOTEST) ./...
