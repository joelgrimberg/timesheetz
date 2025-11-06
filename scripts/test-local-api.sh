#!/bin/bash

# Script to test the local API setup with Docker

set -e

echo "🚀 Starting Timesheetz API in Docker for local testing..."
echo ""

# Start the Docker container
docker-compose -f docker-compose.local.yml up -d

echo ""
echo "⏳ Waiting for API to be ready..."
sleep 5

# Check health
echo ""
echo "🔍 Checking API health..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ API is running and healthy!"
    echo ""
    echo "📋 API is available at: http://localhost:8080"
    echo ""
    echo "To test dual mode:"
    echo "  1. Set in your config: \"apiMode\": \"dual\", \"apiBaseURL\": \"http://localhost:8080\""
    echo "  2. Or set environment variables:"
    echo "     export TIMESHEETZ_API_MODE=dual"
    echo "     export TIMESHEETZ_API_URL=http://localhost:8080"
    echo "  3. Run your local TUI: go run ./cmd/timesheet"
    echo ""
    echo "To view logs: docker-compose -f docker-compose.local.yml logs -f"
    echo "To stop: docker-compose -f docker-compose.local.yml down"
else
    echo "❌ API health check failed"
    echo "Check logs: docker-compose -f docker-compose.local.yml logs"
    exit 1
fi

