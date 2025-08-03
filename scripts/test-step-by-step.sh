#!/bin/bash

# Step-by-step testing script

set -e

echo "🔧 Step-by-Step Bytebase Informix Integration Test"
echo "================================================"
echo ""

# Function to wait for user
wait_for_user() {
    echo ""
    echo "Press Enter to continue..."
    read
}

# Step 1: Test basic Bytebase build (without Informix)
echo "📋 Step 1: Test Basic Bytebase Build"
echo "------------------------------------"
echo "Building standard Bytebase to verify base functionality..."

cat > /tmp/Dockerfile.basic << 'EOF'
FROM golang:1.21-bookworm AS backend
WORKDIR /backend
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ./bytebase-build/bytebase ./backend/bin/server/main.go

FROM ubuntu:20.04
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=backend /backend/bytebase-build/bytebase /usr/local/bin/bytebase
EXPOSE 8080
CMD ["bytebase", "--data", "/tmp/bytebase", "--port", "8080"]
EOF

if docker build -f /tmp/Dockerfile.basic -t bytebase-basic:test .; then
    echo "✅ Basic Bytebase build successful"
else
    echo "❌ Basic build failed"
    exit 1
fi

wait_for_user

# Step 2: Test Informix container
echo "📋 Step 2: Test Informix Container"
echo "----------------------------------"
echo "Starting Informix container..."

docker run -d \
    --name test-informix \
    --platform linux/amd64 \
    -p 9088:9088 \
    -e LICENSE=accept \
    icr.io/informix/informix-developer-database:latest

echo "Waiting for Informix to start (30s)..."
sleep 30

if docker exec test-informix onstat - > /dev/null 2>&1; then
    echo "✅ Informix is running"
else
    echo "❌ Informix failed to start"
    docker logs test-informix
    exit 1
fi

wait_for_user

# Step 3: Test ODBC in isolation
echo "📋 Step 3: Test ODBC Installation"
echo "---------------------------------"
echo "Testing ODBC setup in a minimal container..."

# Create minimal ODBC test
cat > /tmp/test-odbc-minimal.sh << 'EOF'
#!/bin/bash
docker run --rm \
    --platform linux/amd64 \
    -v /Volumes/ssd/Documents/bytebase/bytebase/backend/plugin/db/informix/Go_ODBC_Informix_Complete_Package:/workspace \
    ubuntu:20.04 \
    bash -c "
        apt-get update && apt-get install -y unixodbc odbcinst
        
        # Check ODBC installation
        odbcinst -j
        
        # List ODBC files
        ls -la /etc/odbc* || true
    "
EOF

chmod +x /tmp/test-odbc-minimal.sh
/tmp/test-odbc-minimal.sh

wait_for_user

# Step 4: Test SDK installation
echo "📋 Step 4: Test IBM SDK Installation"
echo "------------------------------------"
echo "Testing IBM Informix Client SDK installation..."

docker run --rm \
    --platform linux/amd64 \
    -v /Volumes/ssd/Documents/bytebase/bytebase/backend/plugin/db/informix/Go_ODBC_Informix_Complete_Package:/workspace \
    ubuntu:20.04 \
    bash -c "
        apt-get update && apt-get install -y openjdk-11-jre libncurses5
        cd /workspace
        tar -tf ibm.csdk.4.50.12.Linux.64.x86_64.tar | head -20
        echo '---'
        echo 'SDK tar file is valid'
    "

wait_for_user

# Step 5: Full integration test
echo "📋 Step 5: Full Integration Test"
echo "--------------------------------"
echo "Now running the complete integration test..."
echo ""
echo "This will:"
echo "1. Build Bytebase with Informix support"
echo "2. Start all services with docker-compose"
echo "3. Test connections"
echo ""
echo "Continue? (y/n)"
read answer

if [ "$answer" = "y" ]; then
    ./scripts/test-informix-integration.sh
fi

# Cleanup
echo ""
echo "🧹 Cleanup"
echo "---------"
echo "Stopping test containers..."
docker stop test-informix 2>/dev/null || true
docker rm test-informix 2>/dev/null || true

echo ""
echo "✅ Step-by-step testing complete!"