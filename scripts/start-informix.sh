#!/bin/bash

# Bytebase with Informix Support - Start Script

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🚀 Starting Bytebase with Informix Support"
echo "========================================"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

cd "$PROJECT_ROOT"

# Build images if needed
echo "📦 Building Docker images..."
docker-compose -f docker-compose.informix.yml build

# Start services
echo "🎯 Starting services..."
docker-compose -f docker-compose.informix.yml up -d

# Wait for services to be ready
echo "⏳ Waiting for services to start..."
sleep 10

# Check service status
echo "📊 Checking service status..."
docker-compose -f docker-compose.informix.yml ps

# Show logs
echo ""
echo "✅ Services started successfully!"
echo ""
echo "📌 Access points:"
echo "   - Bytebase UI: http://localhost:8080"
echo "   - Informix DB: localhost:9088"
echo ""
echo "📝 View logs:"
echo "   docker-compose -f docker-compose.informix.yml logs -f bytebase"
echo ""
echo "🛑 Stop services:"
echo "   docker-compose -f docker-compose.informix.yml down"
echo ""