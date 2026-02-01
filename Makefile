# Makefile

# Platform detection - more robust for Windows, MSYS2, Cygwin, WSL, Linux, macOS
ifeq '$(findstring ;,$(PATH))' ';'
    DETECTED_OS := Windows
else
    DETECTED_OS := $(shell uname -s 2>/dev/null || echo "Unknown")
endif

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

# Binary extension based on platform
ifeq ($(DETECTED_OS),Windows)
    BINARY_EXT := .exe
else
    BINARY_EXT :=
endif

.PHONY: all dev build build-service clean clear-cache test install install-service

all: dev

dev:
ifeq ($(DETECTED_OS),Windows)
	if not exist tmp mkdir tmp
	powershell -Command "air -c .air.toml *>&1 | Tee-Object tmp/app.log"
else
	mkdir -p tmp
	air -c .air.toml 2>&1 | tee tmp/app.log
endif

# View logs only (when dev is already running)
logs:
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "Get-Content tmp/app.log -Wait -Tail 50"
else
	tail -f tmp/app.log
endif

build:
ifeq ($(DETECTED_OS),Windows)
	$(GOBUILD) -o tmp/$(BINARY_NAME).exe $(CMD_PATH)/main.go
else
	$(GOBUILD) -o tmp/$(BINARY_NAME) $(CMD_PATH)/main.go
endif

build-service:
ifeq ($(DETECTED_OS),Windows)
	$(GOBUILD) -o $(CURDIR)/bin/$(SERVICE_BINARY).exe $(SERVICE_PATH)/main.go
else
	$(GOBUILD) -o $(CURDIR)/bin/$(SERVICE_BINARY) $(SERVICE_PATH)/main.go
endif

build-all: build build-service

clean:
	$(GOCLEAN)
ifeq ($(DETECTED_OS),Windows)
	del tmp\$(BINARY_NAME).exe
	del bin\$(SERVICE_BINARY).exe
else
	rm -f tmp/$(BINARY_NAME)
	rm -f bin/$(SERVICE_BINARY)
endif

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
