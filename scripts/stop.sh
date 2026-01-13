#!/bin/bash

# lucidRAG Stop Script
# Stops Go backend, Angular frontend, and optionally MongoDB

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PID_DIR="$PROJECT_DIR/.pids"

echo "Stopping lucidRAG..."

# Stop Angular frontend
if [ -f "$PID_DIR/frontend.pid" ]; then
    PID=$(cat "$PID_DIR/frontend.pid")
    if kill -0 "$PID" 2>/dev/null; then
        echo "   Stopping frontend (PID: $PID)..."
        kill "$PID" 2>/dev/null || true
        # Also kill any child processes (node)
        pkill -P "$PID" 2>/dev/null || true
    fi
    rm -f "$PID_DIR/frontend.pid"
    echo "   Frontend stopped"
else
    echo "   Frontend not running (no PID file)"
fi

# Stop Go backend
if [ -f "$PID_DIR/backend.pid" ]; then
    PID=$(cat "$PID_DIR/backend.pid")
    if kill -0 "$PID" 2>/dev/null; then
        echo "   Stopping backend (PID: $PID)..."
        kill "$PID" 2>/dev/null || true
    fi
    rm -f "$PID_DIR/backend.pid"
    echo "   Backend stopped"
else
    echo "   Backend not running (no PID file)"
fi

# Stop MongoDB (optional - pass --all flag)
if [ "$1" = "--all" ] || [ "$1" = "-a" ]; then
    echo "   Stopping MongoDB..."
    cd "$PROJECT_DIR"
    docker-compose down
    echo "   MongoDB stopped"
else
    echo "   MongoDB left running (use --all to stop)"
fi

# Clean up log files if requested
if [ "$1" = "--clean" ] || [ "$1" = "-c" ]; then
    echo "   Cleaning up logs..."
    rm -f "$PID_DIR"/*.log
    echo "   Logs cleaned"
fi

echo ""
echo "lucidRAG stopped."
