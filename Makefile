# Makefile - Media Manager (VPN Required)
# This application ONLY functions with VPN active for security

.PHONY: all help up down status logs restart build dev test clean

# Default: start with VPN
all: up

help:
	@echo "Media Manager - VPN Required for Security"
	@echo "=========================================="
	@echo ""
	@echo "Usage:"
	@echo "  make up       - Start Media Manager with VPN (default)"
	@echo "  make down     - Stop Media Manager"
	@echo "  make status   - Check VPN and app status"
	@echo "  make logs     - View logs"
	@echo "  make restart  - Restart Media Manager"
	@echo "  make dev      - Development mode with auto-reload"
	@echo ""
	@echo "Other:"
	@echo "  make build    - Build Docker image"
	@echo "  make test     - Run tests"
	@echo "  make clean    - Clean up containers and images"
	@echo ""
	@echo "Your NordVPN token is already in .env"
	@echo "PiHole continues working normally."

# Start with VPN
up: build
	@docker-compose up -d
	@echo ""
	@echo "✓ Media Manager started with VPN"
	@echo "Web UI: http://localhost:8080"
	@echo ""
	@docker-compose ps

# Stop
down:
	@docker-compose down
	@echo "✓ Stopped"

# Status
status:
	@echo "Status:"
	@echo "======="
	@docker-compose ps
	@echo ""
	@echo "VPN:"
	@docker exec media-manager nordvpn status 2>/dev/null || echo "  Disconnected"
	@echo ""
	@echo "IP:"
	@docker exec media-manager curl -s https://api.ipify.org 2>/dev/null || echo "  N/A"

# Logs
logs:
	@docker-compose logs -f

# Restart
restart: down
	@sleep 1
	@$(MAKE) up

# Build
build:
	@docker-compose build

# Dev mode with auto-reload
dev:
	@echo "Development mode with VPN..."
	@docker-compose -f docker-compose.dev.yml up --build

# Tests
test:
	@go test ./...

# Clean
clean:
	@docker-compose down -v --rmi local
	@docker system prune -f
