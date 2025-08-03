#!/bin/bash

# Test basic Bytebase Docker deployment

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Testing Basic Bytebase Docker Deployment${NC}"
echo "=============================================="
echo ""

cd "$PROJECT_ROOT"

# Check Docker
echo -e "${YELLOW}📋 Checking Docker environment...${NC}"
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Docker is running${NC}"

# Clean up any existing containers
echo -e "${YELLOW}🧹 Cleaning up existing containers...${NC}"
docker-compose -f docker-compose.basic.yml down -v 2>/dev/null || true
docker system prune -f > /dev/null 2>&1 || true

# Build the image
echo -e "${YELLOW}📦 Building Bytebase Docker image...${NC}"
echo "This may take several minutes..."
if docker build -f docker/Dockerfile.basic -t bytebase-basic:latest . > /tmp/build.log 2>&1; then
    echo -e "${GREEN}✅ Docker image built successfully${NC}"
else
    echo -e "${RED}❌ Docker build failed${NC}"
    echo "Build log:"
    tail -20 /tmp/build.log
    exit 1
fi

# Start services
echo -e "${YELLOW}🎯 Starting services...${NC}"
if docker-compose -f docker-compose.basic.yml up -d > /tmp/compose.log 2>&1; then
    echo -e "${GREEN}✅ Services started${NC}"
else
    echo -e "${RED}❌ Failed to start services${NC}"
    cat /tmp/compose.log
    exit 1
fi

# Wait for services to be ready
echo -e "${YELLOW}⏳ Waiting for services to initialize...${NC}"
echo "Checking PostgreSQL..."
for i in {1..30}; do
    if docker-compose -f docker-compose.basic.yml exec -T postgres pg_isready -U bbdev > /dev/null 2>&1; then
        echo -e "${GREEN}✅ PostgreSQL is ready${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ PostgreSQL failed to start${NC}"
        docker-compose -f docker-compose.basic.yml logs postgres
        exit 1
    fi
    sleep 2
done

echo "Checking Bytebase..."
for i in {1..60}; do
    if curl -s http://localhost:8080/healthz > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Bytebase is ready${NC}"
        break
    fi
    if [ $i -eq 60 ]; then
        echo -e "${RED}❌ Bytebase failed to start${NC}"
        echo "Bytebase logs:"
        docker-compose -f docker-compose.basic.yml logs bytebase
        exit 1
    fi
    sleep 2
done

# Test basic functionality
echo -e "${YELLOW}🧪 Testing basic functionality...${NC}"

# Test health endpoint
if curl -s http://localhost:8080/healthz | grep -q "OK"; then
    echo -e "${GREEN}✅ Health endpoint working${NC}"
else
    echo -e "${RED}❌ Health endpoint not responding correctly${NC}"
fi

# Test API endpoint
if curl -s http://localhost:8080/api/v1/actuator/info > /dev/null; then
    echo -e "${GREEN}✅ API endpoint accessible${NC}"
else
    echo -e "${RED}❌ API endpoint not accessible${NC}"
fi

# Show service status
echo ""
echo -e "${YELLOW}📊 Service Status:${NC}"
docker-compose -f docker-compose.basic.yml ps

echo ""
echo -e "${GREEN}🎉 Basic Bytebase deployment test completed successfully!${NC}"
echo ""
echo -e "${BLUE}📌 Access Points:${NC}"
echo "   • Bytebase UI: http://localhost:8080"
echo "   • PostgreSQL: localhost:5432 (bbdev/bbdev)"
echo ""
echo -e "${BLUE}📝 Useful Commands:${NC}"
echo "   • View logs: docker-compose -f docker-compose.basic.yml logs -f"
echo "   • Stop services: docker-compose -f docker-compose.basic.yml down"
echo "   • Enter container: docker-compose -f docker-compose.basic.yml exec bytebase bash"
echo ""
echo -e "${YELLOW}⚠️  This is the basic version without Informix support${NC}"
echo "   Next step: Add Informix support after verifying this works"
echo ""