#!/bin/bash

# Test ODBC connection to Informix independently

set -e

echo "🔌 ODBC Connection Test"
echo "======================"
echo ""

# Start only the Informix container if not running
if ! docker ps | grep -q informix-test-db; then
    echo "Starting Informix database..."
    docker run -d \
        --name informix-test-db \
        --platform linux/amd64 \
        -p 9088:9088 \
        -p 9089:9089 \
        -e LICENSE=accept \
        -e STORAGE=local \
        icr.io/informix/informix-developer-database:latest
    
    echo "Waiting for Informix to start (30s)..."
    sleep 30
fi

# Build and run ODBC test container
echo "Building ODBC test container..."
docker build -f - -t odbc-test:latest . << 'EOF'
FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    unixodbc \
    unixodbc-dev \
    odbcinst \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Set up Informix environment
ENV INFORMIXDIR=/opt/IBM/informix
ENV INFORMIXSERVER=informix_tcp
ENV PATH=$INFORMIXDIR/bin:$PATH
ENV LD_LIBRARY_PATH=$INFORMIXDIR/lib:$INFORMIXDIR/lib/cli:$LD_LIBRARY_PATH

# Create directories
RUN mkdir -p $INFORMIXDIR/etc

WORKDIR /test
EOF

# Create test program
cat > /tmp/odbc_test.c << 'EOF'
#include <stdio.h>
#include <stdlib.h>
#include <sql.h>
#include <sqlext.h>

int main() {
    SQLHENV env;
    SQLHDBC dbc;
    SQLHSTMT stmt;
    SQLRETURN ret;
    
    printf("Initializing ODBC...\n");
    
    // Allocate environment
    ret = SQLAllocHandle(SQL_HANDLE_ENV, SQL_NULL_HANDLE, &env);
    if (ret != SQL_SUCCESS) {
        printf("Failed to allocate environment\n");
        return 1;
    }
    
    SQLSetEnvAttr(env, SQL_ATTR_ODBC_VERSION, (void*)SQL_OV_ODBC3, 0);
    
    // Allocate connection
    ret = SQLAllocHandle(SQL_HANDLE_DBC, env, &dbc);
    if (ret != SQL_SUCCESS) {
        printf("Failed to allocate connection\n");
        return 1;
    }
    
    // Connect
    char* connStr = "Driver={IBM INFORMIX ODBC DRIVER};"
                   "Servername=informix_tcp;"
                   "Host=host.docker.internal;"
                   "Service=9088;"
                   "Protocol=onsoctcp;"
                   "Database=sysmaster;"
                   "UID=informix;"
                   "PWD=in4mix";
    
    printf("Connecting with: %s\n", connStr);
    
    ret = SQLDriverConnect(dbc, NULL, (SQLCHAR*)connStr, SQL_NTS, 
                          NULL, 0, NULL, SQL_DRIVER_NOPROMPT);
    
    if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO) {
        printf("✅ Connection successful!\n");
        
        // Execute a simple query
        SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
        ret = SQLExecDirect(stmt, 
            (SQLCHAR*)"SELECT FIRST 1 'Hello from Informix' FROM systables", 
            SQL_NTS);
        
        if (ret == SQL_SUCCESS) {
            char result[100];
            SQLLEN len;
            
            if (SQLFetch(stmt) == SQL_SUCCESS) {
                SQLGetData(stmt, 1, SQL_C_CHAR, result, sizeof(result), &len);
                printf("Query result: %s\n", result);
            }
        }
        
        SQLFreeHandle(SQL_HANDLE_STMT, stmt);
        SQLDisconnect(dbc);
    } else {
        printf("❌ Connection failed\n");
        
        // Get error details
        SQLCHAR sqlState[6], errorMsg[256];
        SQLINTEGER nativeError;
        SQLSMALLINT textLength;
        
        SQLGetDiagRec(SQL_HANDLE_DBC, dbc, 1, sqlState, &nativeError, 
                      errorMsg, sizeof(errorMsg), &textLength);
        printf("Error: %s\n", errorMsg);
    }
    
    SQLFreeHandle(SQL_HANDLE_DBC, dbc);
    SQLFreeHandle(SQL_HANDLE_ENV, env);
    
    return 0;
}
EOF

# Copy SDK to temp location
echo "Copying IBM SDK..."
cp /Volumes/ssd/Documents/bytebase/bytebase/backend/plugin/db/informix/Go_ODBC_Informix_Complete_Package/ibm_sdk/ibm.csdk.*.tar /tmp/

# Run test
echo "Running ODBC test..."
docker run --rm \
    --platform linux/amd64 \
    --add-host=host.docker.internal:host-gateway \
    -v /tmp:/workspace \
    odbc-test:latest \
    bash -c "
        cd /workspace
        tar -xf ibm.csdk.*.tar
        cd informix
        echo -e '1\n\n1\ny' | ./installclientsdk
        
        # Configure ODBC
        cat > /etc/odbcinst.ini << 'ODBC'
[IBM INFORMIX ODBC DRIVER]
Description=IBM Informix ODBC Driver
Driver=/opt/IBM/informix/lib/cli/iclit09b.so
ODBC

        # Configure sqlhosts
        echo 'informix_tcp onsoctcp host.docker.internal 9088' > /opt/IBM/informix/etc/sqlhosts
        
        # Compile and run test
        cd /workspace
        gcc -o odbc_test odbc_test.c -lodbc
        ./odbc_test
    "

echo ""
echo "Test complete!"