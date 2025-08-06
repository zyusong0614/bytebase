#!/bin/bash

# Bytebase Full Version (with Frontend) Podman Deployment Script
set -e

echo "🚀 Starting Bytebase FULL deployment with Frontend and Informix support..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if podman is installed
if ! command -v podman &> /dev/null; then
    echo -e "${RED}Error: Podman is not installed${NC}"
    echo "Please install Podman first: https://podman.io/getting-started/installation"
    exit 1
fi

# Create a pod for all services
echo -e "${GREEN}Creating Podman pod...${NC}"
podman pod create --name bytebase-pod \
    -p 8080:8080 \
    -p 9088-9089:9088-9089 \
    -p 27017-27018:27017-27018 \
    -p 27883:27883 || true

# Create volumes
echo -e "${GREEN}Creating volumes...${NC}"
podman volume create postgres_data || true
podman volume create bytebase_data || true

# Start PostgreSQL container
echo -e "${GREEN}Starting PostgreSQL container...${NC}"
podman run -d \
    --name bytebase-postgres \
    --pod bytebase-pod \
    -e POSTGRES_USER=bbdev \
    -e POSTGRES_PASSWORD=bbdev \
    -e POSTGRES_DB=bbdev \
    -v postgres_data:/var/lib/postgresql/data \
    --restart unless-stopped \
    postgres:14

# Start Informix container
echo -e "${GREEN}Starting Informix container...${NC}"
podman run -d \
    --name informix-test \
    --pod bytebase-pod \
    -e LICENSE=accept \
    -e DB_INIT=1 \
    --restart unless-stopped \
    icr.io/informix/informix-developer-database:latest

# Wait for PostgreSQL to be ready
echo -e "${YELLOW}Waiting for PostgreSQL to be ready...${NC}"
sleep 10
while ! podman exec bytebase-postgres pg_isready -U bbdev -d bbdev &>/dev/null; do
    echo -n "."
    sleep 2
done
echo -e "${GREEN}PostgreSQL is ready!${NC}"

# Wait for Informix to be ready
echo -e "${YELLOW}Waiting for Informix to be ready (this may take 2-3 minutes)...${NC}"
sleep 60
while ! podman exec informix-test timeout 10 bash -c '</dev/tcp/localhost/9088' &>/dev/null; do
    echo -n "."
    sleep 10
done
echo -e "${GREEN}Informix is ready!${NC}"

# Create order database in Informix
echo -e "${GREEN}Creating order database in Informix...${NC}"
podman exec informix-test bash -c "export INFORMIXDIR=/opt/ibm/informix && export INFORMIXSERVER=informix && echo 'CREATE DATABASE order;' | /opt/ibm/informix/bin/dbaccess" || true

# Start Bytebase FULL container with Frontend and Informix support
echo -e "${GREEN}Starting Bytebase FULL container (with Frontend UI)...${NC}"
podman run -d \
    --name bytebase \
    --pod bytebase-pod \
    --privileged \
    -e PG_URL=postgresql://bbdev:bbdev@localhost:5432/bbdev \
    -v bytebase_data:/var/opt/bytebase \
    -v /run/podman/podman.sock:/run/podman/podman.sock \
    --restart unless-stopped \
    bytebase-informix:odbc \
    --data /var/opt/bytebase --port 8080 --external-url http://localhost:8080 --disable-sample

# Check deployment status
echo -e "${GREEN}Checking deployment status...${NC}"
sleep 10
podman pod ps
podman ps --pod

# Verify Bytebase is accessible
echo -e "${YELLOW}Waiting for Bytebase to start...${NC}"
sleep 10
max_attempts=30
attempt=0
while ! curl -s -o /dev/null -w "%{http_code}" http://localhost:8080 | grep -q "200"; do
    echo -n "."
    sleep 2
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
        echo -e "${RED}Bytebase did not start within expected time${NC}"
        echo "Check logs with: podman logs bytebase"
        exit 1
    fi
done

echo -e "${GREEN}✅ Deployment complete!${NC}"
echo -e "${GREEN}🎉 Bytebase FULL VERSION with Web UI is running at: http://localhost:8080${NC}"
echo ""
echo -e "${YELLOW}Features included:${NC}"
echo "✓ Full Web UI Interface"
echo "✓ Informix Database Support"
echo "✓ PostgreSQL for metadata"
echo "✓ Complete functionality"
echo ""
echo "Useful commands:"
echo "- View pod status: podman pod ps"
echo "- View containers: podman ps --pod"
echo "- View Bytebase logs: podman logs bytebase"
echo "- View Informix logs: podman logs informix-test"
echo "- Stop all: podman pod stop bytebase-pod"
echo "- Remove all: podman pod rm -f bytebase-pod"
echo ""
echo "To add Informix database in Bytebase:"
echo "1. Visit http://localhost:8080"
echo "2. Create admin account (first time only)"
echo "3. Add database instance:"
echo "   - Type: INFORMIX"
echo "   - Host: localhost"
echo "   - Port: 9088"
echo "   - Database: order"
echo "   - Username: (leave empty)"
echo "   - Password: (leave empty)"