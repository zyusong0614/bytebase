# Bytebase Informix Integration Documentation

## 项目状态概览

本文档描述了Bytebase与IBM Informix数据库集成的当前实现状态、部署方法和技术方案。

### 当前实现状态

- ✅ **基础连接功能**：支持Informix数据库连接和ping测试
- ✅ **查询功能**：支持SELECT查询（包括WHERE条件）
- ✅ **真实数据访问**：通过docker exec执行真实SQL查询
- ✅ **容器化部署**：完整的Docker Compose部署方案
- ✅ **混合连接策略**：优先尝试原生连接，失败时降级到docker exec
- ⚠️ **INSERT功能**：存在SQL引号转义问题，已有解决方案待测试
- ❌ **真正原生连接**：由于缺少IBM Client SDK，ifxgo驱动无法注册

### 技术架构

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Bytebase      │    │   PostgreSQL     │    │   Informix      │
│   Container     │    │   Container      │    │   Container     │
│                 │    │                  │    │                 │
│ - Docker CLI    │    │ - Metadata       │    │ - Test Database │
│ - ifxgo Driver  │◄──►│   Storage        │    │ - Sample Data   │
│ - Docker Socket │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │
        ▼
   Docker Socket
   (/var/run/docker.sock)
```

## 一键部署指南

### 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- 8GB+ 可用内存
- 网络端口：8080 (Bytebase), 9088-9089 (Informix)

### 部署步骤

#### 1. 获取代码
```bash
git clone <repository-url>
cd bytebase
git checkout ingest-informix-clean
```

#### 2. 构建Docker镜像
```bash
# 构建支持Informix的Bytebase镜像
docker build -f Dockerfile.informix -t bytebase-informix:latest .
```

#### 3. 启动服务
```bash
# 启动完整环境
docker-compose up -d

# 检查服务状态
docker-compose ps
```

#### 4. 访问应用
- Bytebase Web界面: http://localhost:8080
- 首次访问需要设置管理员账户

#### 5. 配置Informix连接
在Bytebase中添加数据库实例：
- **数据库类型**: INFORMIX
- **Host**: `informix-test`
- **Port**: `9088`
- **Database**: `order`
- **Username**: (留空或`informix`)
- **Password**: (留空)

### 测试查询

```sql
-- 查询所有数据
SELECT order_id, order_time, store_id FROM orders ORDER BY order_id;

-- 条件查询
SELECT order_id, order_time, store_id FROM orders WHERE order_id = 101;

-- 插入数据（当前有引号转义问题）
INSERT INTO orders VALUES (104, '2024-01-15 12:00:00', 3, '2024-01-15 12:00:00');
```

## 技术实现详情

### 连接策略

实现了混合连接模式：

```go
// 1. 优先尝试真正的ifxgo连接
sqlDB, err := sql.Open("informix", connStr)
if err != nil {
    // 2. 失败时降级到docker exec
    fmt.Printf("Failed to open ifxgo connection: %v, falling back to container exec\n", err)
}
```

### 查询执行

**原生连接模式**（当ifxgo可用时）：
```go
func (d *Driver) executeRealQuery(ctx context.Context, statement string) ([]*v1pb.QueryResult, error) {
    rows, err := d.odbcDriver.db.QueryContext(ctx, statement)
    // 标准database/sql接口处理
}
```

**Docker Exec模式**（当前默认）：
```go
func (d *Driver) executeInformixQuery(ctx context.Context, statement string) ([]*v1pb.QueryRow, []string, error) {
    cmd := fmt.Sprintf(`docker exec informix-test bash -c 'export INFORMIXDIR=/opt/ibm/informix && export INFORMIXSERVER=informix && cat <<EOF | /opt/ibm/informix/bin/dbaccess order
%s
EOF'`, cleanSQL)
}
```

### 文件结构

```
backend/
├── plugin/db/informix/
│   ├── informix.go              # 基础驱动定义
│   ├── informix_odbc.go         # 主要实现逻辑
│   └── sync.go                  # 同步功能
├── server/
│   └── informix.go              # 驱动注册（构建标签）
└── component/sheet/
    └── sheet.go                 # 语法检查支持

docker-compose.yml               # 完整部署配置
Dockerfile.informix             # Bytebase + Informix镜像
```

## 已知问题和限制

### 当前问题

1. **INSERT语句引号转义**
   - **问题**: SQL中的单引号导致bash命令解析错误
   - **状态**: 已实现heredoc解决方案，待测试
   - **影响**: INSERT、UPDATE、DELETE语句可能失败

2. **ifxgo驱动无法注册**
   - **问题**: 缺少IBM Informix Client SDK
   - **状态**: 已实现fallback机制
   - **影响**: 无法使用原生数据库连接

### 技术限制

1. **性能开销**：每次查询都要启动新的docker exec进程
2. **连接池**：无法实现真正的数据库连接池
3. **事务支持**：有限的事务管理能力
4. **高级功能**：存储过程、prepared statements等支持有限
5. **错误处理**：错误信息需要从文本输出解析

## 解决方案路线图

### 短期计划（1-2周）

- [ ] **修复INSERT功能**
  - 完成heredoc方案测试
  - 验证复杂SQL语句支持
  - 添加更完善的错误处理

- [ ] **性能优化**
  - 实现查询结果缓存
  - 优化docker exec调用
  - 添加查询超时控制

- [ ] **功能完善**
  - 支持更多SQL语句类型
  - 改进数据类型转换
  - 增强错误报告

### 中期计划（1-3个月）

- [ ] **Wire Protocol驱动研究**
  - 评估第三方Wire Protocol驱动
  - 性能基准测试
  - 兼容性验证

- [ ] **IBM Client SDK集成**
  - 研究Client SDK部署方案
  - 评估许可证要求
  - 实现完整原生连接

- [ ] **企业级功能**
  - 连接池管理
  - 完整事务支持
  - 存储过程调用
  - 批量操作优化

### 长期计划（3-6个月）

- [ ] **生产级部署**
  - 多架构支持（ARM64/AMD64）
  - 高可用性配置
  - 监控和日志集成
  - 安全加固

- [ ] **完整功能支持**
  - 所有Informix SQL特性
  - 高级数据类型（BLOB、JSON等）
  - 性能调优
  - 企业级安全功能

## 故障排除

### 常见问题

**1. 容器启动失败**
```bash
# 检查Docker资源
docker system df

# 检查端口占用
lsof -i :8080
lsof -i :9088

# 重新启动
docker-compose down
docker-compose up -d
```

**2. 数据库连接失败**
```bash
# 检查Informix容器状态
docker exec informix-test bash -c "export INFORMIXDIR=/opt/ibm/informix && export INFORMIXSERVER=informix && echo 'DATABASE sysmaster; SELECT name FROM sysdatabases;' | /opt/ibm/informix/bin/dbaccess"

# 重新创建测试数据库
docker exec informix-test bash -c "export INFORMIXDIR=/opt/ibm/informix && export INFORMIXSERVER=informix && echo 'CREATE DATABASE order;' | /opt/ibm/informix/bin/dbaccess"
```

**3. 查询失败**
```bash
# 检查Bytebase日志
docker logs bytebase-informix | grep -i "informix\|error"

# 检查连接方式
docker logs bytebase-informix | grep "fallback\|ifxgo"
```

### 日志分析

**成功的连接日志**：
```
Successfully established ifxgo connection!
Using real ifxgo database connection for query: SELECT ...
```

**Fallback模式日志**：
```
Failed to open ifxgo connection: sql: unknown driver "informix", falling back to container exec
Using docker exec fallback for query: SELECT ...
```

## 开发指南

### 本地开发环境

```bash
# 启动开发环境
docker-compose up -d bytebase-postgres informix-test

# 本地运行Bytebase（开发模式）
PG_URL=postgresql://bbdev:bbdev@localhost:5432/bbdev go run ./backend/bin/server/main.go --port 8080 --data . --debug

# 运行测试
go test -v -tags informix ./backend/plugin/db/informix/...
```

### 代码修改指南

1. **修改查询逻辑**：编辑 `backend/plugin/db/informix/informix_odbc.go`
2. **添加新功能**：在相应的方法中实现
3. **测试更改**：重新构建Docker镜像并测试

### 构建优化

```bash
# 快速构建（跳过缓存）
docker build -f Dockerfile.informix -t bytebase-informix:dev . --no-cache

# 清理构建缓存
docker builder prune
```

## 联系和支持

如遇到问题或需要支持，请：

1. 检查本文档的故障排除部分
2. 查看项目Issues
3. 提供详细的错误日志和环境信息

---

**文档版本**: 1.0  
**最后更新**: 2025-08-05  
**维护者**: Claude AI Assistant  
**项目状态**: 开发中 - 基础功能可用，原生连接待完善