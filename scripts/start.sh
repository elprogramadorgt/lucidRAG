#!/bin/bash

# lucidRAG Start Script
# Starts MongoDB, Go backend, and Angular frontend

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PID_DIR="$PROJECT_DIR/.pids"

# Create PID directory
mkdir -p "$PID_DIR"

echo "🚀 Starting lucidRAG..."

# Check if MongoDB is running via docker-compose
echo "📦 Checking MongoDB..."
if ! docker ps | grep -q "lucidrag.*mongo"; then
    echo "   Starting MongoDB via docker-compose..."
    cd "$PROJECT_DIR"
    docker-compose up -d mongodb
    sleep 3
fi
echo "   ✅ MongoDB is running"

# Start Go backend
echo "🔧 Starting Go backend..."
cd "$PROJECT_DIR"
go run ./cmd/api/main.go > "$PID_DIR/backend.log" 2>&1 &
echo $! > "$PID_DIR/backend.pid"
sleep 2

if kill -0 $(cat "$PID_DIR/backend.pid") 2>/dev/null; then
    echo "   ✅ Backend running on http://localhost:8080 (PID: $(cat "$PID_DIR/backend.pid"))"
else
    echo "   ❌ Backend failed to start. Check $PID_DIR/backend.log"
    exit 1
fi

# Start Angular frontend
echo "🎨 Starting Angular frontend..."
cd "$PROJECT_DIR/ui"
npm start > "$PID_DIR/frontend.log" 2>&1 &
echo $! > "$PID_DIR/frontend.pid"
sleep 5

if kill -0 $(cat "$PID_DIR/frontend.pid") 2>/dev/null; then
    echo "   ✅ Frontend running on http://localhost:4200 (PID: $(cat "$PID_DIR/frontend.pid"))"
else
    echo "   ❌ Frontend failed to start. Check $PID_DIR/frontend.log"
    exit 1
fi

echo ""
echo "✨ lucidRAG is running!"
echo ""
echo "   Frontend: http://localhost:4200"
echo "   Backend:  http://localhost:8080"
echo ""
echo "   Logs:"
echo "   - Backend:  $PID_DIR/backend.log"
echo "   - Frontend: $PID_DIR/frontend.log"
echo ""
echo "   Run './scripts/stop.sh' to stop all services"
