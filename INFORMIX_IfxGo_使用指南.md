# Bytebase Informix IfxGo 驱动使用指南

## ✅ 当前状态

**已完成的集成：**
- ✅ 集成 OpenInformix/IfxGo 驱动
- ✅ 安装 unixODBC 库支持
- ✅ 配置 CGO 编译环境
- ✅ 成功编译并启动后端服务
- ✅ 前端 Informix 界面完全就绪

## 🎯 关于 OpenInformix/IfxGo

**项目信息：**
- GitHub: https://github.com/OpenInformix/IfxGo
- 基于: alexbrainman/odbc 的 fork
- 专门针对: IBM Informix 数据库优化
- 许可证: BSD-3-Clause
- 语言: Go (93% codebase)

**技术特点：**
- 实现标准 database/sql 接口
- 跨平台 ODBC 连接支持
- Windows 下调用 ODBC DLL
- 非 Windows 平台使用 cgo + unixODBC

## 🛠 环境要求

### 系统依赖
1. **unixODBC 库** (已安装)
   ```bash
   brew install unixodbc  # macOS
   ```

2. **CGO 环境变量** (已配置)
   ```bash
   export CGO_CFLAGS="-I/opt/homebrew/include"
   export CGO_LDFLAGS="-L/opt/homebrew/lib"
   ```

3. **Go 编译环境**
   - Go 1.19+ 
   - CGO 支持启用

### Informix 服务器要求
- IBM Informix 数据库服务器
- 支持 ODBC 连接
- 网络可访问性

## 🚀 使用方式

### 1. 在 Bytebase 中创建 Informix 实例

1. **访问实例管理页面**
   ```
   http://localhost:3000/instances
   ```

2. **点击 "创建实例"**

3. **选择 Informix 引擎**
   - 引擎类型: Informix
   - 默认端口: 9088

4. **配置连接参数**
   ```
   主机: your-informix-server
   端口: 9088
   用户名: informix
   密码: your-password
   数据库: your-database
   ```

5. **高级配置（额外连接参数）**
   ```
   SERVER=your-server-name
   PROTOCOL=onsoctcp
   ```

### 2. 连接字符串格式

IfxGo 驱动生成的 ODBC 连接字符串：
```
DRIVER={IBM INFORMIX ODBC DRIVER};HOST=localhost;SERVICE=9088;UID=informix;PWD=password;SERVER=informix;DATABASE=testdb;PROTOCOL=onsoctcp
```

### 3. 测试连接

点击 "测试连接" 按钮验证配置是否正确。

## 🔍 当前限制

### IfxGo 驱动限制
- **社区维护项目** - 不是官方支持
- **有限文档** - 主要依赖源码理解
- **无正式版本** - 使用开发版本
- **ODBC 依赖** - 仍需要系统 ODBC 库

### 需要 Informix ODBC 驱动
⚠️ **重要提醒：** IfxGo 提供了 Go 语言接口，但仍然需要：
1. IBM Informix ODBC 驱动安装
2. 正确配置 ODBC 数据源

## 📋 下一步配置

### 安装 IBM Informix ODBC 驱动

1. **下载驱动**
   - 访问 IBM Informix 下载页面
   - 下载适合 macOS 的 ODBC 驱动

2. **安装驱动**
   ```bash
   # 按照 IBM 安装说明进行
   # 通常安装到 /opt/informix/ 目录
   ```

3. **配置 ODBC**
   
   **编辑 `/etc/odbcinst.ini`:**
   ```ini
   [IBM INFORMIX ODBC DRIVER]
   Description=IBM Informix ODBC Driver
   Driver=/opt/informix/lib/libodbc.so
   Setup=/opt/informix/lib/libodbc.so
   FileUsage=1
   ```

   **编辑 `/etc/odbc.ini`:**
   ```ini
   [informix_dsn]
   Description=Informix Database
   Driver=IBM INFORMIX ODBC DRIVER
   Host=localhost
   Service=9088
   Server=informix
   Database=testdb
   UID=informix
   PWD=password
   Protocol=onsoctcp
   ```

4. **测试 ODBC 连接**
   ```bash
   isql informix_dsn informix password
   ```

## ✅ 成功标志

当配置完成后，你应该能够：
- ✅ 在 Bytebase 界面看到 Informix 选项
- ✅ 成功测试数据库连接
- ✅ 浏览数据库表和架构
- ✅ 执行 SQL 查询
- ✅ 使用 Bytebase 的所有功能

## 🔧 故障排除

### 编译问题
```bash
# 如果遇到编译错误，确保环境变量设置正确
export CGO_CFLAGS="-I/opt/homebrew/include"
export CGO_LDFLAGS="-L/opt/homebrew/lib"
```

### 连接问题
1. **查看后端日志**
   ```bash
   tail -f backend.log
   ```

2. **验证 ODBC 配置**
   ```bash
   odbcinst -j
   odbcinst -q -d
   ```

3. **测试基础连接**
   ```bash
   isql -v DSN_NAME username password
   ```

## 🎉 总结

OpenInformix/IfxGo 为 Bytebase 提供了一个可行的 Informix 集成方案：

**优势：**
- ✅ 纯 Go 实现的数据库接口
- ✅ 专门针对 Informix 优化
- ✅ 标准 database/sql 兼容
- ✅ 已成功集成到 Bytebase

**仍需完成：**
- 🔧 安装 IBM Informix ODBC 驱动
- 🔧 配置系统 ODBC 数据源
- 🔧 测试实际数据库连接

现在的状态是：**框架完全就绪，只需要 Informix ODBC 驱动配置！**