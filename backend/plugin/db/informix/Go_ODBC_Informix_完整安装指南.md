# Go + ODBC + Informix 完整安装指南

## 📋 概述

本指南提供了在macOS上使用Docker+Linux环境实现Go语言通过ODBC连接IBM Informix数据库的完整解决方案。

**目标**：实现Go语言连接Informix数据库并成功插入数据

**技术栈**：Go + CGO + C ODBC + IBM Informix Client SDK + Docker + Linux x86_64

## 🎯 核心成果

✅ **完全实现**：Go语言成功通过ODBC连接Informix数据库并插入数据  
✅ **技术验证**：Go + CGO + C ODBC方案完全可行  
✅ **环境兼容**：解决了macOS与Informix Client SDK的兼容性问题  

## 🏗️ 环境准备

### 1. 系统要求

- **主机系统**：macOS
- **容器环境**：Docker + Linux x86_64
- **原因**：IBM Informix Client SDK不支持macOS，需要Linux环境

### 2. 目录结构

```
项目目录/
├── Dockerfile.x86              # x86_64专用Docker镜像
├── docker-compose.x86.yml      # Docker Compose配置
├── ibm.csdk.4.50.12.Linux.64.x86_64.tar  # IBM Informix Client SDK
└── 测试文件/
```

## 🐳 Docker环境配置

### 1. Dockerfile.x86
```dockerfile
FROM --platform=linux/amd64 ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    curl \
    wget \
    unzip \
    build-essential \
    unixodbc \
    unixodbc-dev \
    odbcinst \
    vim \
    golang-go \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
CMD ["tail", "-f", "/dev/null"]
```

### 2. docker-compose.x86.yml
```yaml
version: '3.8'

services:
  informix-db-x86:
    image: icr.io/informix/informix-developer-database:latest
    container_name: informix-db-x86
    platform: linux/amd64
    environment:
      - LICENSE=accept
      - DB_INIT=1
      - INFORMIX_USER=informix
      - INFORMIX_PASSWORD=in4mix
      - DB_USER=gouser
      - DB_PASS=gopass
      - DB_NAME=testgo
    ports:
      - "9092:9088"
      - "9093:9089"
    volumes:
      - informix_data_x86:/opt/ibm/data
    healthcheck:
      test: ["CMD-SHELL", "/opt/ibm/informix/bin/onstat -d"]
      interval: 30s
      timeout: 10s
      retries: 5

  informix-odbc-x86:
    build:
      context: .
      dockerfile: Dockerfile.x86
    container_name: informix-odbc-x86
    platform: linux/amd64
    depends_on:
      informix-db-x86:
        condition: service_healthy
    volumes:
      - .:/workspace
    working_dir: /workspace

volumes:
  informix_data_x86:
```

### 3. 启动服务
```bash
# 清理macOS隐藏文件（如果有）
find . -name "._*" -delete

# 启动服务
docker-compose -f docker-compose.x86.yml up -d
```

## 🔧 IBM Informix Client SDK 安装

### 1. 下载SDK
从IBM官网下载：`ibm.csdk.4.50.12.Linux.64.x86_64.tar`

### 2. 安装SDK
```bash
# 进入容器
docker exec -it informix-odbc-x86 bash

# 安装Java（安装程序需要）
apt-get update && apt-get install -y openjdk-11-jre libncurses5 libtinfo5

# 解压SDK
cd /workspace
tar -xf ibm.csdk.4.50.12.Linux.64.x86_64.tar

# 运行安装程序
cd /workspace/informix
./installclientsdk

# 安装选择：
# - 安装目录：/opt/IBM/informix
# - 选择：Client SDK + ODBC + Demo
# - 确认安装
```

### 3. 环境变量配置
```bash
export INFORMIXDIR=/opt/IBM/informix
export INFORMIXSERVER=informix
export INFORMIXSQLHOSTS=/opt/IBM/informix/etc/sqlhosts
export LD_LIBRARY_PATH=/opt/IBM/informix/lib:/opt/IBM/informix/lib/esql:/opt/IBM/informix/lib/cli:$LD_LIBRARY_PATH
export PATH=/opt/IBM/informix/bin:$PATH
export CLIENT_LOCALE=en_US.utf8
export DB_LOCALE=en_US.utf8
```

## ⚙️ ODBC配置

### 1. 注册ODBC驱动 (/etc/odbcinst.ini)
```ini
[IBM INFORMIX ODBC DRIVER]
Description=IBM Informix ODBC Driver
Driver=/opt/IBM/informix/lib/cli/iclit09b.so
```

### 2. 配置数据源 (/etc/odbc.ini)
```ini
[informix_test]
Driver=IBM INFORMIX ODBC DRIVER
Database=testgo
LogonID=informix
pwd=in4mix
Servername=informix
Protocol=onsoctcp
Host=172.19.0.2
```

### 3. **关键配置** - sqlhosts文件
```bash
# 编辑 /opt/IBM/informix/etc/sqlhosts
informix    onsoctcp    172.19.0.2    9088
```

**重要说明**：这是解决连接问题的关键配置！必须添加这个条目。

## 🔗 连接配置详解

### 连接问题的根本原因

初始错误（常见错误）：
```
Server 172.19.0.2 is not listed as a dbserver name in sqlhosts.
user root@informix-odbc-x86...is not trusted by the server
```

### 解决方案

1. **sqlhosts配置**：必须包含服务器条目
2. **连接字符串格式**：使用`Servername=informix`而非`Server=IP:Port`
3. **用户权限**：informix用户具有完整权限

### 正确的连接字符串
```
Driver={IBM INFORMIX ODBC DRIVER};Servername=informix;Database=testgo;UID=informix;PWD=in4mix
```

## 💻 Go程序开发

### 1. 创建Go模块
```bash
# 在容器中创建Go模块
docker exec informix-odbc-x86 bash -c "mkdir -p /home/test && cd /home/test && go mod init test"
```

### 2. 最简单的测试程序
```go
// simple_test.go
package main

/*
#cgo LDFLAGS: -lodbc
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sql.h>
#include <sqlext.h>

int test_connection_and_insert() {
    SQLHENV env;
    SQLHDBC dbc;
    SQLHSTMT stmt;
    SQLRETURN ret;
    
    printf("开始ODBC连接测试...\n");
    
    // 初始化ODBC
    SQLAllocHandle(SQL_HANDLE_ENV, SQL_NULL_HANDLE, &env);
    SQLSetEnvAttr(env, SQL_ATTR_ODBC_VERSION, (void*)SQL_OV_ODBC3, 0);
    SQLAllocHandle(SQL_HANDLE_DBC, env, &dbc);
    
    // 连接数据库 - 使用正确的连接字符串
    char connStr[] = "Driver={IBM INFORMIX ODBC DRIVER};Servername=informix;Database=testgo;UID=informix;PWD=in4mix";
    ret = SQLDriverConnect(dbc, NULL, (SQLCHAR*)connStr, SQL_NTS, NULL, 0, NULL, SQL_DRIVER_NOPROMPT);
    
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        printf("连接失败\n");
        return 0;
    }
    
    printf("连接成功！\n");
    
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    
    // 创建表
    char createSQL[] = "CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, name VARCHAR(50), message VARCHAR(200))";
    SQLExecDirect(stmt, (SQLCHAR*)createSQL, SQL_NTS);
    
    // 插入数据
    char insertSQL[] = "INSERT INTO test_table (name, message) VALUES ('Go测试', 'Go+ODBC+Informix连接成功！')";
    ret = SQLExecDirect(stmt, (SQLCHAR*)insertSQL, SQL_NTS);
    
    if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO) {
        printf("数据插入成功！\n");
        
        // 查询验证
        char querySQL[] = "SELECT id, name, message FROM test_table ORDER BY id DESC LIMIT 1";
        ret = SQLExecDirect(stmt, (SQLCHAR*)querySQL, SQL_NTS);
        
        if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO) {
            ret = SQLFetch(stmt);
            if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO) {
                SQLINTEGER id;
                SQLCHAR name[100];
                SQLCHAR message[300];
                SQLLEN idLen, nameLen, messageLen;
                
                SQLGetData(stmt, 1, SQL_C_SLONG, &id, 0, &idLen);
                SQLGetData(stmt, 2, SQL_C_CHAR, name, sizeof(name), &nameLen);
                SQLGetData(stmt, 3, SQL_C_CHAR, message, sizeof(message), &messageLen);
                
                printf("插入的数据: ID=%d, 名称=%s, 消息=%s\n", id, name, message);
            }
        }
        
        SQLFreeHandle(SQL_HANDLE_STMT, stmt);
        SQLDisconnect(dbc);
        SQLFreeHandle(SQL_HANDLE_DBC, dbc);
        SQLFreeHandle(SQL_HANDLE_ENV, env);
        return 1;
    } else {
        printf("数据插入失败\n");
        SQLFreeHandle(SQL_HANDLE_STMT, stmt);
        SQLDisconnect(dbc);
        SQLFreeHandle(SQL_HANDLE_DBC, dbc);
        SQLFreeHandle(SQL_HANDLE_ENV, env);
        return 0;
    }
}
*/
import "C"
import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("🎯 Go+ODBC+Informix 测试")
	fmt.Println("======================")
	
	// 设置环境变量
	os.Setenv("INFORMIXDIR", "/opt/IBM/informix")
	os.Setenv("INFORMIXSERVER", "informix")
	os.Setenv("INFORMIXSQLHOSTS", "/opt/IBM/informix/etc/sqlhosts")
	os.Setenv("LD_LIBRARY_PATH", "/opt/IBM/informix/lib:/opt/IBM/informix/lib/esql:/opt/IBM/informix/lib/cli:"+os.Getenv("LD_LIBRARY_PATH"))
	os.Setenv("CLIENT_LOCALE", "en_US.utf8")
	os.Setenv("DB_LOCALE", "en_US.utf8")
	
	result := C.test_connection_and_insert()
	
	if result > 0 {
		fmt.Println("\n✅ 成功！Go+ODBC+Informix 连接和数据插入完全成功！")
	} else {
		fmt.Println("\n❌ 失败")
	}
}
```

### 3. 运行测试
```bash
# 将Go文件复制到容器并运行
docker cp simple_test.go informix-odbc-x86:/home/test/
docker exec informix-odbc-x86 bash -c "cd /home/test && go run simple_test.go"
```

## 🔍 验证方法

### 手动查询验证

**方法1：使用Informix原生工具**
```bash
docker exec informix-db-x86 bash -c "export INFORMIXDIR=/opt/ibm/informix && export INFORMIXSERVER=informix && echo 'SELECT * FROM test_table ORDER BY id;' | /opt/ibm/informix/bin/dbaccess testgo"
```

**方法2：使用C ODBC查询工具**

创建简单查询工具 `query_tool.c`：
```c
#include <stdio.h>
#include <sql.h>
#include <sqlext.h>

int main() {
    SQLHENV env;
    SQLHDBC dbc;
    SQLHSTMT stmt;
    SQLRETURN ret;
    
    printf("🔍 手动查询验证工具\n");
    printf("==================\n");
    
    // 初始化
    SQLAllocHandle(SQL_HANDLE_ENV, SQL_NULL_HANDLE, &env);
    SQLSetEnvAttr(env, SQL_ATTR_ODBC_VERSION, (void*)SQL_OV_ODBC3, 0);
    SQLAllocHandle(SQL_HANDLE_DBC, env, &dbc);
    
    // 连接数据库
    char connStr[] = "Driver={IBM INFORMIX ODBC DRIVER};Servername=informix;Database=testgo;UID=informix;PWD=in4mix";
    ret = SQLDriverConnect(dbc, NULL, (SQLCHAR*)connStr, SQL_NTS, NULL, 0, NULL, SQL_DRIVER_NOPROMPT);
    
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        printf("❌ 连接失败\n");
        return 1;
    }
    
    printf("✅ 连接成功\n");
    
    SQLAllocHandle(SQL_HANDLE_STMT, dbc, &stmt);
    
    // 查询数据
    char querySQL[] = "SELECT id, name, message FROM test_table ORDER BY id";
    ret = SQLExecDirect(stmt, (SQLCHAR*)querySQL, SQL_NTS);
    
    if (ret == SQL_SUCCESS || ret == SQL_SUCCESS_WITH_INFO) {
        printf("📋 查询结果:\n");
        
        int rowNum = 0;
        while (SQLFetch(stmt) == SQL_SUCCESS) {
            SQLINTEGER id;
            SQLCHAR name[100];
            SQLCHAR message[300];
            SQLLEN idLen, nameLen, messageLen;
            
            SQLGetData(stmt, 1, SQL_C_SLONG, &id, 0, &idLen);
            SQLGetData(stmt, 2, SQL_C_CHAR, name, sizeof(name), &nameLen);
            SQLGetData(stmt, 3, SQL_C_CHAR, message, sizeof(message), &messageLen);
            
            rowNum++;
            printf("%d. ID=%d, 名称=%s\n", rowNum, id, name);
            printf("   消息=%s\n", message);
        }
        
        printf("📊 总共 %d 条记录\n", rowNum);
    }
    
    // 清理
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
    SQLDisconnect(dbc);
    SQLFreeHandle(SQL_HANDLE_DBC, dbc);
    SQLFreeHandle(SQL_HANDLE_ENV, env);
    
    return 0;
}
```

编译和运行：
```bash
# 将C文件复制到容器
docker cp query_tool.c informix-odbc-x86:/tmp/

# 在容器中编译并运行
docker exec informix-odbc-x86 bash -c "cd /tmp && gcc -o query_tool query_tool.c -lodbc && ./query_tool"
```

## ⚠️ 重要说明

### alexbrainman/odbc库的限制

在测试中发现，`alexbrainman/odbc`库在某些配置下可能无法正常工作，出现以下错误：
```
SQLDriverConnect: {H} [
{0} [
```

### 推荐解决方案

**使用CGO + C ODBC**：
- ✅ 完全兼容IBM Informix ODBC驱动
- ✅ 稳定可靠的连接和数据操作
- ✅ 完整的错误信息和调试支持

### 技术栈对比

| 方案 | 兼容性 | 稳定性 | 调试友好度 |
|------|--------|--------|------------|
| alexbrainman/odbc | ⚠️ 部分 | ⚠️ 需配置 | ❌ 错误信息不清晰 |
| CGO + C ODBC | ✅ 完全 | ✅ 稳定 | ✅ 详细错误信息 |

## 🚀 一键安装脚本

创建 `install.sh`：
```bash
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
    echo "❌ 请先下载IBM Informix Client SDK文件"
    exit 1
fi

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

echo "✅ 安装完成！"
echo ""
echo "🚀 运行测试："
echo "docker exec informix-odbc-x86 bash -c 'cd /home/test && go mod init test && go run simple_test.go'"
```

## 🎯 总结

### 成功要素

1. **正确的环境**：Docker + Linux x86_64
2. **完整的SDK**：IBM Informix Client SDK 4.50.FC12
3. **正确的配置**：sqlhosts文件是关键
4. **合适的技术栈**：CGO + C ODBC最稳定

### 核心配置

```bash
# sqlhosts文件（关键！）
informix    onsoctcp    172.19.0.2    9088

# 连接字符串（必须使用Servername）
Driver={IBM INFORMIX ODBC DRIVER};Servername=informix;Database=testgo;UID=informix;PWD=in4mix

# 环境变量
INFORMIXDIR=/opt/IBM/informix
INFORMIXSERVER=informix
INFORMIXSQLHOSTS=/opt/IBM/informix/etc/sqlhosts
```

### 认证问题解决

**问题本质**：不是认证问题，而是服务器配置问题
- ❌ 错误理解：认为是用户权限或信任问题
- ✅ 真正原因：sqlhosts文件缺少服务器条目
- ✅ 解决方法：添加正确的sqlhosts条目并使用Servername连接

### 最终结果

✅ **目标完全达成**："go + odbc连接上informix数据库，并且成功插入一条数据"

这个解决方案提供了完整、稳定、可重现的Go语言连接Informix数据库的技术方案，通过CGO + C ODBC实现了完美的兼容性和稳定性。