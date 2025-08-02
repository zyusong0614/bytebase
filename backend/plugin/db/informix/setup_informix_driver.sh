#!/bin/bash

echo "🚀 Bytebase Informix Driver Setup"
echo "================================="

# Check if we're on macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "⚠️  macOS detected - IBM Informix Client SDK requires Linux environment"
    echo "    This script will set up Docker environment for Informix support"
    
    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker is required but not installed"
        echo "   Please install Docker Desktop for macOS"
        exit 1
    fi
    
    echo "✅ Docker found"
    
    # Check if IBM SDK file exists
    if [ ! -f "downloads/ibm.csdk.4.50.12.Linux.64.x86_64.tar" ]; then
        echo "❌ IBM Informix Client SDK not found"
        echo "   Expected: downloads/ibm.csdk.4.50.12.Linux.64.x86_64.tar"
        echo "   Make sure you have the complete package"
        exit 1
    fi
    
    echo "✅ IBM Informix Client SDK found"
    
    # Build Docker image with Informix SDK
    echo "🔨 Building Informix Docker environment..."
    docker build -f Dockerfile.informix -t bytebase-informix .
    
    if [ $? -eq 0 ]; then
        echo "✅ Docker environment built successfully"
        echo ""
        echo "🎯 Next steps:"
        echo "1. Start your Informix database server"
        echo "2. Update the connection parameters in your Bytebase configuration"
        echo "3. Test the connection using Bytebase UI"
        echo ""
        echo "💡 For testing, you can run:"
        echo "   docker run -it --rm bytebase-informix /workspace/simple_test.go"
    else
        echo "❌ Failed to build Docker environment"
        exit 1
    fi
    
else
    echo "🐧 Linux environment detected"
    echo "📦 Installing IBM Informix Client SDK directly..."
    
    # Check if IBM SDK file exists
    if [ ! -f "downloads/ibm.csdk.4.50.12.Linux.64.x86_64.tar" ]; then
        echo "❌ IBM Informix Client SDK not found"
        exit 1
    fi
    
    # Extract and install SDK
    echo "📦 Extracting IBM Informix Client SDK..."
    tar -xf downloads/ibm.csdk.4.50.12.Linux.64.x86_64.tar
    cd informix
    echo -e '1\n\n1\ny' | sudo ./installclientsdk
    
    # Set up environment
    echo "⚙️  Configuring environment..."
    
    # Create sqlhosts
    sudo mkdir -p /opt/IBM/informix/etc
    echo "informix    onsoctcp    localhost    9088" | sudo tee /opt/IBM/informix/etc/sqlhosts
    
    # Configure ODBC driver
    sudo tee /etc/odbcinst.ini > /dev/null << EOF
[IBM INFORMIX ODBC DRIVER]
Description=IBM Informix ODBC Driver
Driver=/opt/IBM/informix/lib/cli/iclit09b.so
EOF
    
    # Configure data source
    sudo tee /etc/odbc.ini > /dev/null << EOF
[informix_default]
Driver=IBM INFORMIX ODBC DRIVER
Database=bytebase
LogonID=informix
Servername=informix
Protocol=onsoctcp
Host=localhost
EOF
    
    # Set environment variables
    echo "export INFORMIXDIR=/opt/IBM/informix" | sudo tee -a /etc/environment
    echo "export INFORMIXSERVER=informix" | sudo tee -a /etc/environment
    echo "export INFORMIXSQLHOSTS=/opt/IBM/informix/etc/sqlhosts" | sudo tee -a /etc/environment
    echo "export LD_LIBRARY_PATH=/opt/IBM/informix/lib:/opt/IBM/informix/lib/esql:/opt/IBM/informix/lib/cli:\$LD_LIBRARY_PATH" | sudo tee -a /etc/environment
    echo "export CLIENT_LOCALE=en_US.utf8" | sudo tee -a /etc/environment
    echo "export DB_LOCALE=en_US.utf8" | sudo tee -a /etc/environment
    
    echo "✅ IBM Informix Client SDK installed successfully"
    echo ""
    echo "🔄 Please restart your shell or run:"
    echo "   source /etc/environment"
    echo ""
    echo "🎯 Then rebuild Bytebase with:"
    echo "   CGO_CFLAGS=\"-I/opt/IBM/informix/incl/cli\" CGO_LDFLAGS=\"-L/opt/IBM/informix/lib/cli -lodbc -liclit09b\" go build ..."
fi

echo ""
echo "📚 For more information, see:"
echo "   - Go_ODBC_Informix_完整安装指南.md"
echo "   - 快速开始.md"