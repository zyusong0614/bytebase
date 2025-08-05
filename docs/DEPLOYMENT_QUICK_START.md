# Bytebase Informix Integration - Quick Start Deployment Guide

## 🚀 在新机器上一键部署

### 前置要求

```bash
# 1. 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 2. 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 3. 验证安装
docker --version
docker-compose --version
```

### 部署步骤

```bash
# Step 1: 克隆代码
git clone <your-repository>
cd bytebase
git checkout ingest-informix-clean

# Step 2: 一键启动
docker-compose up -d

# Step 3: 等待服务启动（约2-3分钟）
docker-compose ps

# Step 4: 访问应用
open http://localhost:8080
```

### 验证部署

```bash
# 检查所有服务状态
docker-compose ps

# 应该看到以下输出：
# NAME                IMAGE                     STATUS
# bytebase-informix   bytebase-informix:latest  Up X minutes
# bytebase-postgres   postgres:14               Up X minutes (healthy)
# informix-test       icr.io/informix/...       Up X minutes (healthy)

# 检查网络连接
curl -I http://localhost:8080
# 应该返回: HTTP/1.1 200 OK
```

## 🗂️ 项目文件结构

```
bytebase/
├── docker-compose.yml           # 🔧 主要部署配置
├── Dockerfile.informix          # 🐳 Bytebase+Informix镜像
├── docs/                        # 📚 文档目录
│   ├── INFORMIX_INTEGRATION.md  # 完整技术文档
│   └── DEPLOYMENT_QUICK_START.md # 本文件
├── backend/
│   ├── plugin/db/informix/      # 🔌 Informix驱动实现
│   │   ├── informix.go
│   │   ├── informix_odbc.go     # 主要实现
│   │   └── sync.go
│   └── server/
│       └── informix.go          # 驱动注册
└── frontend/                    # 🎨 Web界面（标准Bytebase）
```

## 🔧 配置Informix连接

### 在Bytebase中添加数据库

1. **访问**: http://localhost:8080
2. **首次设置**: 创建管理员账户
3. **添加数据库实例**:
   - Database Type: **INFORMIX**
   - Host: `informix-test`
   - Port: `9088`
   - Database: `order`
   - Username: (留空)
   - Password: (留空)
4. **测试连接**: 点击 "Test Connection"

### 测试查询

```sql
-- 查看所有订单
SELECT order_id, order_time, store_id FROM orders ORDER BY order_id;

-- 条件查询
SELECT * FROM orders WHERE order_id = 101;
```

## 🚨 故障排除

### 常见问题

**问题1: 端口占用**
```bash
# 检查端口占用
sudo lsof -i :8080
sudo lsof -i :9088

# 停止占用进程或修改docker-compose.yml中的端口映射
```

**问题2: 容器启动失败**
```bash
# 查看详细日志
docker-compose logs bytebase-informix
docker-compose logs informix-test

# 重新启动
docker-compose down
docker-compose up -d
```

**问题3: 数据库连接失败**
```bash
# 检查Informix数据库状态
docker exec informix-test bash -c "export INFORMIXDIR=/opt/ibm/informix && export INFORMIXSERVER=informix && echo 'DATABASE sysmaster; SELECT name FROM sysdatabases;' | /opt/ibm/informix/bin/dbaccess"

# 如果看不到'order'数据库，重新创建：
docker exec informix-test bash -c "export INFORMIXDIR=/opt/ibm/informix && export INFORMIXSERVER=informix && echo 'CREATE DATABASE order;' | /opt/ibm/informix/bin/dbaccess"
```

### 重置环境

```bash
# 完全重置（会删除所有数据）
docker-compose down -v
docker system prune -f
docker-compose up -d
```

## 📊 系统资源要求

| 组件 | CPU | 内存 | 存储 |
|------|-----|------|------|
| Bytebase | 0.5 核 | 1GB | 500MB |
| PostgreSQL | 0.2 核 | 512MB | 1GB |
| Informix | 1.0 核 | 2GB | 2GB |
| **总计** | **1.7 核** | **3.5GB** | **3.5GB** |

## 🔄 更新和维护

### 更新到最新版本
```bash
# 获取最新代码
git pull origin ingest-informix-clean

# 重新构建镜像
docker-compose down
docker build -f Dockerfile.informix -t bytebase-informix:latest .
docker-compose up -d
```

### 备份数据
```bash
# 备份PostgreSQL数据（Bytebase元数据）
docker exec bytebase-postgres pg_dump -U bbdev bbdev > bytebase_backup.sql

# 备份Informix数据
docker exec informix-test bash -c "export INFORMIXDIR=/opt/ibm/informix && export INFORMIXSERVER=informix && echo 'UNLOAD TO /tmp/orders_backup.txt SELECT * FROM orders;' | /opt/ibm/informix/bin/dbaccess order"
```

## 🛠️ 开发环境

### 本地开发
```bash
# 只启动依赖服务
docker-compose up -d bytebase-postgres informix-test

# 本地运行Bytebase
PG_URL=postgresql://bbdev:bbdev@localhost:5432/bbdev go run ./backend/bin/server/main.go --port 8080 --data . --debug
```

### 构建自定义镜像
```bash
# 构建开发版本
docker build -f Dockerfile.informix -t bytebase-informix:dev .

# 使用开发版本
# 修改docker-compose.yml中的image: bytebase-informix:dev
docker-compose up -d
```

## 📞 获取帮助

1. **查看完整文档**: `docs/INFORMIX_INTEGRATION.md`
2. **检查日志**: `docker-compose logs -f bytebase-informix`
3. **系统状态**: `docker-compose ps`
4. **资源使用**: `docker stats`

---

**快速参考**:
- Bytebase UI: http://localhost:8080
- Informix Port: 9088-9089
- PostgreSQL: 5432 (内部访问)
- 启动命令: `docker-compose up -d`
- 停止命令: `docker-compose down`
- 查看日志: `docker-compose logs -f`