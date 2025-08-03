#!/bin/bash

# Comprehensive test script for Bytebase Informix integration

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test results
print_test_result() {
    local test_name=$1
    local result=$2
    local details=$3
    
    if [ "$result" = "PASS" ]; then
        echo -e "${GREEN}✓ $test_name${NC}"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ $test_name${NC}"
        echo -e "  ${YELLOW}Details: $details${NC}"
        ((TESTS_FAILED++))
    fi
}

echo "🧪 Bytebase Informix Integration Test Suite"
echo "=========================================="
echo ""

cd "$PROJECT_ROOT"

# Test 1: Docker Environment Check
echo "📋 Test 1: Docker Environment Check"
echo "-----------------------------------"

if docker info > /dev/null 2>&1; then
    print_test_result "Docker is running" "PASS"
else
    print_test_result "Docker is running" "FAIL" "Docker is not running"
    exit 1
fi

if docker-compose version > /dev/null 2>&1; then
    print_test_result "Docker Compose is available" "PASS"
else
    print_test_result "Docker Compose is available" "FAIL" "Docker Compose not found"
    exit 1
fi

echo ""

# Test 2: Build Docker Images
echo "📋 Test 2: Build Docker Images"
echo "------------------------------"

echo "Building Bytebase with Informix support..."
if docker build -f docker/Dockerfile.informix -t bytebase-informix:test . > /tmp/build.log 2>&1; then
    print_test_result "Docker image build" "PASS"
else
    print_test_result "Docker image build" "FAIL" "Check /tmp/build.log for details"
fi

echo ""

# Test 3: Start Services
echo "📋 Test 3: Start Services"
echo "------------------------"

echo "Starting services..."
if docker-compose -f docker-compose.informix.yml up -d > /tmp/compose.log 2>&1; then
    print_test_result "Services started" "PASS"
    
    # Wait for services to be ready
    echo "Waiting for services to initialize (30s)..."
    sleep 30
else
    print_test_result "Services started" "FAIL" "Check /tmp/compose.log for details"
fi

echo ""

# Test 4: Service Health Checks
echo "📋 Test 4: Service Health Checks"
echo "--------------------------------"

# Check PostgreSQL
if docker-compose -f docker-compose.informix.yml exec -T postgres pg_isready -U bbdev > /dev/null 2>&1; then
    print_test_result "PostgreSQL is healthy" "PASS"
else
    print_test_result "PostgreSQL is healthy" "FAIL" "PostgreSQL not responding"
fi

# Check Informix
if docker-compose -f docker-compose.informix.yml exec -T informix onstat - > /dev/null 2>&1; then
    print_test_result "Informix is healthy" "PASS"
else
    print_test_result "Informix is healthy" "FAIL" "Informix not responding"
fi

# Check Bytebase
if curl -f http://localhost:8080/healthz > /dev/null 2>&1; then
    print_test_result "Bytebase is healthy" "PASS"
else
    print_test_result "Bytebase is healthy" "FAIL" "Bytebase not responding on port 8080"
fi

echo ""

# Test 5: ODBC Configuration Test
echo "📋 Test 5: ODBC Configuration Test"
echo "----------------------------------"

# Create test script for ODBC
cat > /tmp/test_odbc.sh << 'EOF'
#!/bin/bash
# Test ODBC configuration inside container

# Check ODBC drivers
echo "Checking ODBC drivers..."
odbcinst -q -d

# Check ODBC data sources
echo "Checking ODBC data sources..."
odbcinst -q -s

# Test connection with isql
echo "Testing ODBC connection..."
echo "SELECT FIRST 1 'ODBC Connection Successful' FROM systables;" | isql -v informix_default informix in4mix -b

exit $?
EOF

chmod +x /tmp/test_odbc.sh

if docker-compose -f docker-compose.informix.yml exec -T bytebase bash /tmp/test_odbc.sh > /tmp/odbc_test.log 2>&1; then
    print_test_result "ODBC configuration" "PASS"
else
    print_test_result "ODBC configuration" "FAIL" "Check /tmp/odbc_test.log for details"
fi

echo ""

# Test 6: Informix Driver Integration
echo "📋 Test 6: Informix Driver Integration"
echo "--------------------------------------"

# Create Go test program
cat > /tmp/test_informix_driver.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "os"
    
    storepb "github.com/bytebase/bytebase/backend/generated-go/store"
    "github.com/bytebase/bytebase/backend/plugin/db"
)

func main() {
    config := db.ConnectionConfig{
        DataSource: &storepb.DataSource{
            Host:     "informix",
            Port:     "9088",
            Username: "informix",
        },
        ConnectionContext: db.ConnectionContext{
            DatabaseName: "sysmaster",
        },
        Password: "in4mix",
    }
    
    driver, err := db.Open(context.Background(), storepb.Engine_INFORMIX, config)
    if err != nil {
        fmt.Printf("Failed to open driver: %v\n", err)
        os.Exit(1)
    }
    defer driver.Close(context.Background())
    
    err = driver.Ping(context.Background())
    if err != nil {
        fmt.Printf("Failed to ping: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Driver integration successful!")
}
EOF

# Test driver inside container
if docker-compose -f docker-compose.informix.yml exec -T bytebase bash -c "cd /tmp && go run -tags informix test_informix_driver.go" > /tmp/driver_test.log 2>&1; then
    print_test_result "Informix driver integration" "PASS"
else
    print_test_result "Informix driver integration" "FAIL" "Check /tmp/driver_test.log for details"
fi

echo ""

# Test 7: Bytebase API Test
echo "📋 Test 7: Bytebase API Test"
echo "----------------------------"

# Test API endpoint
if curl -s http://localhost:8080/api/v1/auth/login > /dev/null; then
    print_test_result "Bytebase API responding" "PASS"
else
    print_test_result "Bytebase API responding" "FAIL" "API not accessible"
fi

echo ""

# Test Summary
echo "📊 Test Summary"
echo "---------------"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All tests passed!${NC}"
    echo ""
    echo "🎯 Next steps:"
    echo "1. Access Bytebase UI at http://localhost:8080"
    echo "2. Add an Informix instance in Bytebase"
    echo "3. Test SQL queries and schema operations"
else
    echo -e "${RED}❌ Some tests failed. Please check the logs for details.${NC}"
    echo ""
    echo "📝 Debug commands:"
    echo "- View logs: docker-compose -f docker-compose.informix.yml logs"
    echo "- Enter Bytebase container: docker-compose -f docker-compose.informix.yml exec bytebase bash"
    echo "- Check ODBC: docker-compose -f docker-compose.informix.yml exec bytebase odbcinst -j"
fi

echo ""
echo "🛑 To stop services: docker-compose -f docker-compose.informix.yml down"
echo ""

exit $TESTS_FAILED