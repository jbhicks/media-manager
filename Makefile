# Makefile - Media Manager

.PHONY: all help dev logs logs-service logs-web build-web test

# Default: development mode
all: dev

help:
	@echo "Media Manager"
	@echo "============="
	@echo ""
	@echo "Usage:"
	@echo "  make dev         - Start development server with auto-reload"
	@echo "  make logs        - View service and web logs in tmux"
	@echo "  make logs-service - View service logs"
	@echo "  make logs-web    - View web logs"
	@echo "  make build-web   - Build web UI"
	@echo "  make test        - Run tests"

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

# Build web UI
build-web:
	@echo "Building web UI..."
	@cd web && npm install && npm run build
	@echo "✓ Web UI built"

# Tests
test:
	@go test ./...
