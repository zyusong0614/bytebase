#!/bin/bash

echo "🚀 Go+ODBC+Informix 一键安装"
echo "==========================="

# 1. 清理隐藏文件
echo "1. 清理macOS隐藏文件..."
find . -name "._*" -delete

# 2. 启动Docker服务
echo "2. 启动Docker服务..."
docker-compose -f docker-compose.x86.yml up -d

# 等待服务启动
echo "等待Informix数据库启动..."
sleep 30

# 3. 安装依赖
echo "3. 安装依赖..."
docker exec informix-odbc-x86 bash -c "apt-get update && apt-get install -y openjdk-11-jre libncurses5 libtinfo5"

# 4. 检查SDK文件
if [ ! -f "ibm.csdk.4.50.12.Linux.64.x86_64.tar" ]; then
    echo "❌ IBM Informix Client SDK文件不存在"
    echo "   预期文件: ibm.csdk.4.50.12.Linux.64.x86_64.tar"
    echo "   请确保使用完整包: Go_ODBC_Informix_Complete_Package.tar.gz"
    exit 1
fi

echo "✅ 找到IBM Informix Client SDK文件"

# 5. 安装SDK
echo "5. 安装IBM Informix Client SDK..."
docker exec informix-odbc-x86 bash -c "
cd /workspace
tar -xf ibm.csdk.4.50.12.Linux.64.x86_64.tar
cd informix
echo -e '1\n\n1\ny' | ./installclientsdk
"

# 6. 配置环境
echo "6. 配置ODBC环境..."
docker exec informix-odbc-x86 bash -c '
# 创建sqlhosts
mkdir -p /opt/IBM/informix/etc
echo "informix    onsoctcp    172.19.0.2    9088" > /opt/IBM/informix/etc/sqlhosts

# 配置ODBC驱动
cat > /etc/odbcinst.ini << EOF
[IBM INFORMIX ODBC DRIVER]
Description=IBM Informix ODBC Driver
Driver=/opt/IBM/informix/lib/cli/iclit09b.so
EOF

# 配置数据源  
cat > /etc/odbc.ini << EOF
[informix_test]
Driver=IBM INFORMIX ODBC DRIVER
Database=testgo
LogonID=informix
pwd=in4mix
Servername=informix
Protocol=onsoctcp
Host=172.19.0.2
EOF
'

# 7. 设置Go模块
echo "7. 设置Go模块..."
docker exec informix-odbc-x86 bash -c "mkdir -p /home/test && cd /home/test && go mod init test"

echo "✅ 安装完成！"
echo ""
echo "🚀 运行测试："
echo "docker cp simple_test.go informix-odbc-x86:/home/test/"
echo "docker exec informix-odbc-x86 bash -c 'cd /home/test && go run simple_test.go'"
echo ""
echo "🔍 手动验证："
echo "docker cp query_tool.c informix-odbc-x86:/tmp/"
echo "docker exec informix-odbc-x86 bash -c 'cd /tmp && gcc -o query_tool query_tool.c -lodbc && ./query_tool'"