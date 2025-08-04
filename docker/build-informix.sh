#!/bin/bash
# Build script for Bytebase with Informix support

set -e

echo "=== Building Bytebase with Informix Support ==="
echo

# Check if SDK file exists
SDK_PATH="backend/plugin/db/informix/Go_ODBC_Informix_Complete_Package/ibm_sdk/ibm.csdk.4.50.12.Linux.64.x86_64.tar"
if [ ! -f "$SDK_PATH" ]; then
    echo "ERROR: IBM Informix Client SDK not found at $SDK_PATH"
    exit 1
fi

echo "1. Found IBM Informix Client SDK"

# Clean extended attributes
echo "2. Cleaning macOS extended attributes..."
find . -name "._*" -delete 2>/dev/null || true
xattr -rc . 2>/dev/null || true

# Build Docker image
echo "3. Building Docker image..."
docker build -f docker/Dockerfile.informix.fixed -t bytebase-informix:latest .

echo
echo "Build complete! Image tagged as bytebase-informix:latest"
echo
echo "To verify the build:"
echo "  ./scripts/verify_informix_docker.sh bytebase-informix:latest"
echo
echo "To run Bytebase with Informix support:"
echo "  docker run -d --name bytebase-informix -p 8080:8080 bytebase-informix:latest"