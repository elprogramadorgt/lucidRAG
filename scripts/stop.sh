#!/bin/bash

# lucidRAG Stop Script
# Stops Go backend, Angular frontend, and optionally MongoDB

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PID_DIR="$PROJECT_DIR/.pids"

echo "Stopping lucidRAG..."

# Stop Angular frontend
FRONTEND_STOPPED=false

# Try PID file first
if [ -f "$PID_DIR/frontend.pid" ]; then
    PID=$(cat "$PID_DIR/frontend.pid")
    if kill -0 "$PID" 2>/dev/null; then
        echo "   Stopping frontend (PID: $PID)..."
		PGID=$(ps -o pgid= -p $PID)
        kill -- "-$PGID" 2>/dev/null || true
        FRONTEND_STOPPED=true
    fi
    rm -f "$PID_DIR/frontend.pid"
fi

if [ "$FRONTEND_STOPPED" = true ]; then
    echo "   Frontend stopped"
else
    echo "   Frontend not running"
fi

# Stop Go backend
BACKEND_STOPPED=false

# Try PID file first
if [ -f "$PID_DIR/backend.pid" ]; then
    PID=$(cat "$PID_DIR/backend.pid")
    if kill -0 "$PID" 2>/dev/null; then
        echo "   Stopping backend (PID: $PID)..."
		PGID=$(ps -o pgid= -p $PID)
        kill -- "-$PGID" 2>/dev/null || true
        BACKEND_STOPPED=true
    fi
    rm -f "$PID_DIR/backend.pid"
fi

if [ "$BACKEND_STOPPED" = true ]; then
    echo "   Backend stopped"
else
    echo "   Backend not running"
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

# Verify ports are free
echo ""
echo "Verifying ports..."
sleep 1

PORT_8080_FREE=true
PORT_4200_FREE=true

if lsof -t -iTCP:8080 -sTCP:LISTEN >/dev/null 2>&1; then
    echo "   ⚠️  Port 8080 is still in use"
    PORT_8080_FREE=false
else
    echo "   ✓ Port 8080 is free"
fi

if lsof -t -iTCP:4200 -sTCP:LISTEN >/dev/null 2>&1; then
    echo "   ⚠️  Port 4200 is still in use"
    PORT_4200_FREE=false
else
    echo "   ✓ Port 4200 is free"
fi

echo ""
if [ "$PORT_8080_FREE" = true ] && [ "$PORT_4200_FREE" = true ]; then
    echo "lucidRAG stopped."
else
    echo "lucidRAG stopped with warnings."
    exit 1
fi
