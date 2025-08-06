#!/bin/bash
set -e

echo "=== Testing IBM Informix Client SDK Installation ==="
echo "Architecture: $(uname -m)"
echo "Platform: $(uname -s)-$(uname -r)"
echo ""

# Test x86_64 SDK if available
if [ -f "./backend/plugin/db/informix/ibm.csdk.4.50.12.Linux.64.x86_64.tar" ]; then
    echo "Found x86_64 SDK: ibm.csdk.4.50.12.Linux.64.x86_64.tar"
    
    # Extract and check contents
    mkdir -p /tmp/x86_test
    cd /tmp/x86_test
    cp /Users/vn58kzj/IdeaProjects/bytebase/backend/plugin/db/informix/ibm.csdk.4.50.12.Linux.64.x86_64.tar outer.tar
    tar -tf outer.tar | head -10
    echo ""
    
    # Extract the outer tar file to see structure
    tar -xf outer.tar
    
    # Check for nested tar file with different name
    inner_tar=$(find . -name "*.tar" -not -name "outer.tar" | head -1)
    if [ -n "$inner_tar" ] && [ -f "$inner_tar" ]; then
        echo "Found nested tar file: $inner_tar, extracting..."
        tar -xf "$inner_tar"
    fi
    
    # List contents
    echo "Extracted contents:"
    ls -la
    
    # Check installer
    if [ -f "installclientsdk" ]; then
        echo ""
        echo "Found installer: installclientsdk"
        file installclientsdk
        echo ""
        echo "Installer size: $(wc -c < installclientsdk) bytes"
    fi
    
    cd - > /dev/null
    rm -rf /tmp/x86_test
fi

# Test ARM64 SDK if available  
if [ -f "./backend/plugin/db/informix/ibm.csdk.4.50.FC11.ARMV8.tar" ]; then
    echo ""
    echo "Found ARM64 SDK: ibm.csdk.4.50.FC11.ARMV8.tar"
    
    # Extract and check contents
    mkdir -p /tmp/arm64_test
    cd /tmp/arm64_test
    cp /Users/vn58kzj/IdeaProjects/bytebase/backend/plugin/db/informix/ibm.csdk.4.50.FC11.ARMV8.tar outer.tar
    tar -tf outer.tar | head -10
    echo ""
    
    # Extract the outer tar file to see structure  
    tar -xf outer.tar
    
    # Check for nested tar file with different name
    inner_tar=$(find . -name "*.tar" -not -name "outer.tar" | head -1)
    if [ -n "$inner_tar" ] && [ -f "$inner_tar" ]; then
        echo "Found nested tar file: $inner_tar, extracting..."
        tar -xf "$inner_tar"
    fi
    
    # List contents
    echo "Extracted contents:"
    ls -la
    
    # Check installer
    if [ -f "installclientsdk" ]; then
        echo ""
        echo "Found installer: installclientsdk"
        file installclientsdk
        echo ""
        echo "Installer size: $(wc -c < installclientsdk) bytes"
    fi
    
    cd - > /dev/null
    rm -rf /tmp/arm64_test
fi

echo ""
echo "=== Library Analysis ==="
echo "System libraries available:"
echo "libc.so.6: $(find /usr/lib /lib -name 'libc.so.6' 2>/dev/null | head -3)"
echo "libpthread.so.0: $(find /usr/lib /lib -name 'libpthread.so.0' 2>/dev/null | head -3)"
echo "libdl.so.2: $(find /usr/lib /lib -name 'libdl.so.2' 2>/dev/null | head -3)"

echo ""
echo "=== Analysis Summary ==="
echo "The IBM Client SDK installer validation appears to be hardcoded to check specific library paths"
echo "that don't match modern Linux distributions. This is likely a bug in IBM's InstallAnywhere"
echo "installer that affects both ARM64 and x86_64 versions when run on Ubuntu/Debian systems."