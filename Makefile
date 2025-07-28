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
ifeq ($(OS),Windows_NT)
	if not exist tmp mkdir tmp
	air
else
	mkdir -p tmp
	air
endif

build:
	$(GOBUILD) -o tmp/$(BINARY_NAME).exe $(CMD_PATH)/main.go

clean:
	$(GOCLEAN)
ifeq ($(OS),Windows_NT)
	del tmp\$(BINARY_NAME).exe
else
	rm -f bin/$(BINARY_NAME)
endif

clear-cache:
	@echo "Clearing thumbnail cache..."
	@rm -rf ~/.media-manager/thumbnails/* ./thumbnails/* 2>/dev/null || true
	@echo "Thumbnail cache cleared!"
