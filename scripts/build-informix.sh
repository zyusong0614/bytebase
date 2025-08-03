#!/bin/bash

# Build Bytebase with Informix ODBC support

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🔨 Building Bytebase with Informix Support"
echo "========================================"

cd "$PROJECT_ROOT"

# Build the Docker image
echo "📦 Building Docker image..."
docker build -f docker/Dockerfile.informix -t bytebase-informix:latest .

echo ""
echo "✅ Build completed successfully!"
echo ""
echo "🚀 To start Bytebase with Informix support:"
echo "   ./scripts/start-informix.sh"
echo ""