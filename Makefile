# Makefile - Media Manager

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

.PHONY: all help dev logs logs-service logs-web build build-service build-all build-web installer clean clear-cache test install uninstall-shell install-service

# Default: development mode
all: dev

help:
	@echo "Media Manager"
	@echo "============="
	@echo ""
	@echo "Usage:"
	@echo "  make dev             - Start development services"
	@echo "  make logs            - View service and web logs in tmux"
	@echo "  make logs-service    - View service logs"
	@echo "  make logs-web        - View web logs"
	@echo "  make build           - Build the desktop app"
	@echo "  make build-service   - Build the backend service"
	@echo "  make build-web       - Build web UI"
	@echo "  make installer       - Build the Windows installer"
	@echo "  make test            - Run tests"
	@echo "  make install         - Install the desktop app and Explorer shell entries"
	@echo "  make uninstall-shell - Remove Explorer shell entries"

# Development mode with auto-reload (native)
dev:
	@echo "Development mode..."
	@./scripts/dev-server.sh restart

# Logs
logs:
	@TERM=screen-256color tmux new-session -d -s media-manager-logs
	@tmux split-window -v
	@tmux send-keys -t media-manager-logs:0.0 'tail -f tmp/service.log' C-m
	@tmux send-keys -t media-manager-logs:0.1 'tail -f tmp/web.log' C-m
	@tmux attach-session -t media-manager-logs

logs-service:
	@tail -f tmp/service.log

logs-web:
	@tail -f tmp/web.log

build:
	$(GOBUILD) -ldflags="-H=windowsgui" -o tmp/$(BINARY_NAME).exe $(CMD_PATH)/main.go

build-service:
	$(GOBUILD) -o $(CURDIR)/bin/$(SERVICE_BINARY) $(SERVICE_PATH)/main.go

build-all: build build-service

# Build web UI
build-web:
	@echo "Building web UI..."
	@cd web && npm install && npm run build
	@echo "✓ Web UI built"

installer:
	goreleaser build --snapshot --clean
	@echo "Creating Windows installer..."
	@if exist "dist\media-manager_windows_amd64_v1\media-manager.exe" ( \
		copy "dist\media-manager_windows_amd64_v1\media-manager.exe" "." && \
		copy "dist\media-manager-service_windows_amd64_v1\media-manager-service.exe" "." && \
		"C:\Program Files (x86)\NSIS\makensis.exe" installer.nsi && \
		move "media-manager_installer.exe" "dist\" && \
		del "media-manager.exe" && \
		del "media-manager-service.exe" \
	) else ( \
		echo "Build failed - binaries not found" \
	)

clean:
	$(GOCLEAN)
ifeq ($(OS),Windows_NT)
	del tmp\$(BINARY_NAME).exe
else
	rm -f bin/$(BINARY_NAME)
endif

clear-cache:
	@echo "Clearing all media-manager cache..."
	@rm -rf ~/.media-manager/thumbnails/* ~/.media-manager/previews/* ~/.media-manager/video_previews/* ./thumbnails/* 2>/dev/null || true
	@echo "All media-manager cache cleared!"

# Tests
test:
	$(GOTEST) ./...

install:
	$(MAKE) build
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -Command "$$installDir = $$env:GOBIN; if ([string]::IsNullOrWhiteSpace($$installDir)) { if (-not [string]::IsNullOrWhiteSpace($$env:GOPATH)) { $$installDir = Join-Path $$env:GOPATH 'bin' } else { $$installDir = Join-Path $$env:USERPROFILE 'go\\bin' } }; New-Item -ItemType Directory -Force -Path $$installDir | Out-Null; $$exePath = Join-Path $$installDir '$(BINARY_NAME).exe'; Copy-Item 'tmp/$(BINARY_NAME).exe' $$exePath -Force; Write-Host ('Installed $(BINARY_NAME).exe to ' + $$installDir); Write-Host 'Registering per-user Explorer context menu entries...'; $$cmdOpen = '\"' + $$exePath + '\" \"%%*\"'; $$cmdOpenParent = '\"' + $$exePath + '\" --open-parent \"%%*\"'; $$cmdDir = '\"' + $$exePath + '\" \"%%1\"'; $$cmdBackground = '\"' + $$exePath + '\" \"%%V\"'; & reg add 'HKCU\Software\Classes\*\shell\MediaManager.Open' /ve /d 'Open in Media Manager' /f | Out-Null; & reg add 'HKCU\Software\Classes\*\shell\MediaManager.Open' /v 'Icon' /d $$exePath /f | Out-Null; & reg add 'HKCU\Software\Classes\*\shell\MediaManager.Open' /v 'MultiSelectModel' /d 'Player' /f | Out-Null; & reg add 'HKCU\Software\Classes\*\shell\MediaManager.Open\command' /ve /d $$cmdOpen /f | Out-Null; & reg add 'HKCU\Software\Classes\*\shell\MediaManager.OpenParent' /ve /d 'Open Parent Folder in Media Manager' /f | Out-Null; & reg add 'HKCU\Software\Classes\*\shell\MediaManager.OpenParent' /v 'Icon' /d $$exePath /f | Out-Null; & reg add 'HKCU\Software\Classes\*\shell\MediaManager.OpenParent' /v 'MultiSelectModel' /d 'Player' /f | Out-Null; & reg add 'HKCU\Software\Classes\*\shell\MediaManager.OpenParent\command' /ve /d $$cmdOpenParent /f | Out-Null; & reg add 'HKCU\Software\Classes\Directory\shell\MediaManager.Open' /ve /d 'Open in Media Manager' /f | Out-Null; & reg add 'HKCU\Software\Classes\Directory\shell\MediaManager.Open' /v 'Icon' /d $$exePath /f | Out-Null; & reg add 'HKCU\Software\Classes\Directory\shell\MediaManager.Open\command' /ve /d $$cmdDir /f | Out-Null; & reg add 'HKCU\Software\Classes\Directory\Background\shell\MediaManager.Open' /ve /d 'Open in Media Manager' /f | Out-Null; & reg add 'HKCU\Software\Classes\Directory\Background\shell\MediaManager.Open' /v 'Icon' /d $$exePath /f | Out-Null; & reg add 'HKCU\Software\Classes\Directory\Background\shell\MediaManager.Open\command' /ve /d $$cmdBackground /f | Out-Null; & reg add 'HKCU\Software\Classes\Drive\shell\MediaManager.Open' /ve /d 'Open in Media Manager' /f | Out-Null; & reg add 'HKCU\Software\Classes\Drive\shell\MediaManager.Open' /v 'Icon' /d $$exePath /f | Out-Null; & reg add 'HKCU\Software\Classes\Drive\shell\MediaManager.Open\command' /ve /d $$cmdDir /f | Out-Null; Write-Host 'Explorer context menu integration added for current user.'"
else
	@install_dir="${GOBIN}"; \
	if [ -z "$$install_dir" ]; then install_dir="${GOPATH}/bin"; fi; \
	if [ -z "${GOPATH}" ] && [ -z "${GOBIN}" ]; then install_dir="$$HOME/go/bin"; fi; \
	mkdir -p "$$install_dir"; \
	cp "tmp/$(BINARY_NAME).exe" "$$install_dir/$(BINARY_NAME)"; \
	echo "Installed $(BINARY_NAME) to $$install_dir"
endif

uninstall-shell:
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -Command "& reg delete 'HKCU\Software\Classes\*\shell\MediaManager.Open' /f 2>$$null | Out-Null; & reg delete 'HKCU\Software\Classes\*\shell\MediaManager.OpenParent' /f 2>$$null | Out-Null; & reg delete 'HKCU\Software\Classes\Directory\shell\MediaManager.Open' /f 2>$$null | Out-Null; & reg delete 'HKCU\Software\Classes\Directory\Background\shell\MediaManager.Open' /f 2>$$null | Out-Null; & reg delete 'HKCU\Software\Classes\Drive\shell\MediaManager.Open' /f 2>$$null | Out-Null; Write-Host 'Removed per-user Explorer context menu integration.'"
else
	@echo "uninstall-shell is only applicable on Windows."
endif

install-service:
	$(MAKE) build-service
ifeq ($(OS),Windows_NT)
	@echo "install-service is a Linux/systemd workflow and is not available on Windows."
	@echo "Service binary built at bin/$(SERVICE_BINARY)."
else
	@echo "Installing service binary to /usr/local/bin (requires sudo)..."
	@echo "sudo cp bin/$(SERVICE_BINARY) /usr/local/bin/$(SERVICE_BINARY)" > sudo_install_service.sh
	@echo "sudo chmod +x /usr/local/bin/$(SERVICE_BINARY)" >> sudo_install_service.sh
	@echo "sudo cp media-manager-service@.service /etc/systemd/system/" >> sudo_install_service.sh
	@echo "sudo systemctl daemon-reload" >> sudo_install_service.sh
	@echo "Please run: ./sudo_install_service.sh"
	@chmod +x sudo_install_service.sh
endif
