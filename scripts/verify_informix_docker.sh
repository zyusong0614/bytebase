#!/bin/bash
# Script to verify Informix ODBC driver and SDK integration in Docker image

echo "=== Informix Docker Integration Verification ==="
echo

# Check if docker image exists
IMAGE_NAME="${1:-bytebase-informix:latest}"
echo "Checking Docker image: $IMAGE_NAME"

# Create verification script
cat > /tmp/verify_informix_in_container.sh << 'EOF'
#!/bin/bash
echo "1. Checking Informix environment variables:"
env | grep INFORMIX

echo -e "\n2. Checking ODBC configuration:"
cat /etc/odbcinst.ini 2>/dev/null || echo "No odbcinst.ini found"

echo -e "\n3. Checking Informix directory structure:"
ls -la /opt/IBM/informix/ 2>/dev/null || echo "No Informix directory found"

echo -e "\n4. Checking ODBC driver files:"
ls -la /opt/IBM/informix/lib/cli/ 2>/dev/null || echo "No CLI directory found"

echo -e "\n5. Checking if iclit09b.so exists:"
find /opt/IBM/ -name "iclit09b.so" 2>/dev/null || echo "iclit09b.so not found"

echo -e "\n6. Checking LD_LIBRARY_PATH:"
echo $LD_LIBRARY_PATH

echo -e "\n7. Checking Bytebase binary:"
ldd /usr/local/bin/bytebase 2>/dev/null | grep -i odbc || echo "No ODBC dependency found"

echo -e "\n8. Testing ODBC driver availability:"
if command -v odbcinst &> /dev/null; then
    odbcinst -q -d
else
    echo "odbcinst command not found"
fi

echo -e "\n9. Checking Bytebase build info:"
/usr/local/bin/bytebase version 2>&1 || echo "Cannot run bytebase version"

echo -e "\n10. Checking if Informix support is compiled in:"
strings /usr/local/bin/bytebase | grep -i "informix" | head -20
EOF

# Run verification in container
echo -e "\nRunning verification script in container...\n"
docker run --rm --entrypoint /bin/bash -v /tmp/verify_informix_in_container.sh:/verify.sh $IMAGE_NAME /verify.sh

# Cleanup
rm -f /tmp/verify_informix_in_container.sh