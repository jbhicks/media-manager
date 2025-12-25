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
SERVICE_BINARY=media-manager-service
CMD_PATH=./cmd/media-manager
SERVICE_PATH=./cmd/media-manager-service

.PHONY: all dev build build-service clean clear-cache test install install-service

all: dev

# 'dev' now always runs the app with --dev-reset (see .air.toml), so cache and database are cleared on each start
dev:
	$(GOBUILD) -o bin/clear-previews ./cmd/clear-previews/main.go
	bin/clear-previews
	mkdir -p tmp
	CLEAR_DB_ON_START=true air

build:
	$(GOBUILD) -o $(CURDIR)/bin/$(BINARY_NAME) $(CMD_PATH)/main.go

build-service:
	$(GOBUILD) -o $(CURDIR)/bin/$(SERVICE_BINARY) $(SERVICE_PATH)/main.go

build-all: build build-service

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)

clear-cache:
	@echo "Clearing all media-manager cache..."
	@rm -rf ~/.media-manager/thumbnails/* ~/.media-manager/previews/* ~/.media-manager/video_previews/* ./thumbnails/* 2>/dev/null || true
	@echo "All media-manager cache cleared!"

test:
	$(GOTEST) ./...

install:
	$(MAKE) build
	@if [ -n "$$GOBIN" ]; then \
		install_dir="$$GOBIN"; \
	elif [ -n "$$GOPATH" ]; then \
		install_dir="$$GOPATH/bin"; \
	else \
		install_dir="$$HOME/go/bin"; \
	fi; \
	mkdir -p "$$install_dir"; \
	cp bin/$(BINARY_NAME) "$$install_dir/$(BINARY_NAME)"; \
	echo "Installed $(BINARY_NAME) to $$install_dir"

install-service:
	$(MAKE) build-service
	@echo "Installing service binary to /usr/local/bin (requires sudo)..."
	@echo "sudo cp bin/$(SERVICE_BINARY) /usr/local/bin/$(SERVICE_BINARY)" > sudo_install_service.sh
	@echo "sudo chmod +x /usr/local/bin/$(SERVICE_BINARY)" >> sudo_install_service.sh
	@echo "sudo cp media-manager-service@.service /etc/systemd/system/" >> sudo_install_service.sh
	@echo "sudo systemctl daemon-reload" >> sudo_install_service.sh
	@echo "Please run: ./sudo_install_service.sh"
	@chmod +x sudo_install_service.sh
