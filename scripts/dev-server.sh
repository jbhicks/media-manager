#!/bin/bash
# Development server wrapper script
# Ensures only one instance of each service runs

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PID_DIR="$PROJECT_DIR/tmp/pids"
LOCK_FILE="$PROJECT_DIR/tmp/dev.lock"

# Function to check if a process is running
check_process() {
    local pattern="$1"
    pgrep -f "$pattern" > /dev/null 2>&1
}

# Function to get PID of a process
get_pid() {
    local pattern="$1"
    pgrep -f "$pattern" | head -1
}

# Function to kill processes safely
kill_processes() {
    echo "Stopping existing development processes..."
    
    # Kill air
    if check_process "air -c .air-service.toml"; then
        pkill -f "air -c .air-service.toml" 2>/dev/null || true
        echo "  ✓ Stopped air"
    fi
    
    # Kill vite from this project
    if check_process "media-manager.*vite"; then
        pgrep -f "media-manager.*vite" | xargs -r kill 2>/dev/null || true
        echo "  ✓ Stopped vite"
    fi
    
    # Kill media-manager-service
    if check_process "media-manager-service"; then
        pkill -f "media-manager-service" 2>/dev/null || true
        echo "  ✓ Stopped service"
    fi
    
    # Clean up PID files and lock file
    rm -f "$PID_DIR"/*.pid 2>/dev/null || true
    rm -f "$LOCK_FILE" 2>/dev/null || true
    
    sleep 1
    echo "All processes stopped"
}

# Function to start services
start_services() {
    # Check if already running
    if [ -f "$LOCK_FILE" ]; then
        LOCK_PID=$(cat "$LOCK_FILE" 2>/dev/null)
        if ps -p "$LOCK_PID" > /dev/null 2>&1; then
            echo "Development server already running!"
            echo "Use './scripts/dev-server.sh stop' to stop it first"
            echo "Or './scripts/dev-server.sh restart' to restart"
            exit 1
        else
            rm -f "$LOCK_FILE"
        fi
    fi
    
    # Check for existing processes
    if check_process "air -c .air-service.toml" || check_process "media-manager-service"; then
        echo "WARNING: Existing development processes detected"
        echo "Stopping them first..."
        kill_processes
    fi
    
    # Ensure directories exist
    mkdir -p "$PID_DIR"
    
    echo "Starting development services..."
    echo ""
    
    # Create lock file
    echo $$ > "$LOCK_FILE"
    
    # Start backend service
    echo "  → Starting backend service on port 8084..."
    cd "$PROJECT_DIR"
    PORT=8084 air -c .air-service.toml > "$PROJECT_DIR/tmp/service.log" 2>&1 &
    echo $! > "$PID_DIR/service.pid"
    sleep 3
    
    # Start web frontend
    echo "  → Starting web frontend..."
    cd "$PROJECT_DIR/web"
    npm run dev > "$PROJECT_DIR/tmp/web.log" 2>&1 &
    echo $! > "$PID_DIR/web.pid"
    sleep 2
    
    # Get the web port
    WEB_URL=$(grep -oE 'http://localhost:[0-9]+/' "$PROJECT_DIR/tmp/web.log" | tail -1)
    
    echo ""
    echo "✓ Development services started!"
    echo ""
    echo "  Backend PID: $(cat $PID_DIR/service.pid)"
    echo "  Frontend PID: $(cat $PID_DIR/web.pid)"
    echo ""
    echo "  Web UI: $WEB_URL"
    echo ""
    echo "Useful commands:"
    echo "  ./scripts/dev-server.sh stop     - Stop all services"
    echo "  ./scripts/dev-server.sh restart  - Restart services"
    echo "  ./scripts/dev-server.sh status   - Check status"
    echo "  make logs-service                - View backend logs"
    echo "  make logs-web                    - View frontend logs"
}

# Function to show status
show_status() {
    echo "Development Server Status"
    echo "========================="
    echo ""
    
    if [ -f "$LOCK_FILE" ]; then
        LOCK_PID=$(cat "$LOCK_FILE" 2>/dev/null)
        echo "Lock file: $LOCK_PID"
    fi
    
    echo -n "Air (backend): "
    if check_process "air -c .air-service.toml"; then
        echo "RUNNING (PID: $(get_pid "air -c .air-service.toml"))"
    else
        echo "STOPPED"
    fi
    
    echo -n "Vite (frontend): "
    if check_process "media-manager.*vite"; then
        echo "RUNNING (PID: $(get_pid "media-manager.*vite"))"
    else
        echo "STOPPED"
    fi
    
    echo -n "Service: "
    if check_process "media-manager-service"; then
        echo "RUNNING (PID: $(get_pid "media-manager-service"))"
    else
        echo "STOPPED"
    fi
}

# Main command handler
case "${1:-start}" in
    start)
        start_services
        ;;
    stop)
        kill_processes
        ;;
    restart)
        kill_processes
        sleep 2
        start_services
        ;;
    status)
        show_status
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
