# Informix Driver Setup Guide

## Current Status

✅ **Completed:**
- Informix engine integration in Bytebase frontend and backend
- Driver registration and framework setup
- Frontend UI support (engine selection, port configuration, SSL/SSH support)
- Complete database metadata synchronization implementation
- Error handling and user guidance

⚠️ **Requires Setup:**
- ODBC driver installation and configuration
- Actual database driver connection

## What You Have Now

The Informix integration is **functionally complete** but uses a placeholder driver that provides clear setup instructions. When you try to connect to an Informix database, you'll get a detailed error message with setup instructions.

## File Locations

### Backend Files
- **Driver Implementation:** `backend/plugin/db/informix/informix.go`
- **Metadata Sync:** `backend/plugin/db/informix/sync.go`
- **Engine Configuration:** `backend/common/engine.go`
- **Driver Registration:** `backend/server/ultimate.go`

### Frontend Files
- **Engine Support:** `frontend/src/utils/v1/instance.ts`
- **Port/Icon Config:** `frontend/src/components/InstanceForm/constants.ts`
- **Icon Asset:** `frontend/src/assets/db/informix.png`

### Protocol Buffers
- **Store Proto:** `proto/store/store/common.proto` (INFORMIX = 29)
- **API Proto:** `proto/v1/v1/common.proto` (INFORMIX = 29)

## To Enable Actual Informix Connections

### Step 1: Install ODBC Development Libraries

**Ubuntu/Debian:**
```bash
sudo apt-get install unixodbc-dev
```

**macOS:**
```bash
brew install libiodbc
# or
brew install unixodbc
```

**RHEL/CentOS:**
```bash
yum install unixODBC-devel
```

### Step 2: Install IBM Informix ODBC Driver

⚠️ **Important:** Do NOT use the Maven JAR file (com.ibm.informix:odbc:4.10.10). 
That is a Java driver and cannot be used with Go applications.

**Download the correct driver:**
1. Visit IBM Informix Downloads page
2. Look for "Informix Client SDK" or "Informix ODBC Driver"
3. Download the **native C/C++ version** for your platform:
   - Linux: `.so` library files
   - macOS: `.dylib` library files  
   - Windows: `.dll` library files
4. Install the driver package following IBM's instructions
5. Configure the driver in your system ODBC configuration

**Alternative sources:**
- IBM Informix Client SDK (includes ODBC driver)
- Third-party ODBC drivers compatible with Informix
- Docker containers with pre-configured Informix ODBC

### Step 3: Update the Code

In `backend/plugin/db/informix/informix.go`:

1. **Uncomment the ODBC import:**
   ```go
   _ "github.com/alexbrainman/odbc"
   ```

2. **Replace the error return in the Open method with:**
   ```go
   sqlDB, err := sql.Open("odbc", dsn)
   if err != nil {
       return nil, errors.Wrapf(err, "failed to open Informix connection")
   }
   sqlDB.SetConnMaxLifetime(10 * time.Minute)
   sqlDB.SetMaxOpenConns(10)
   sqlDB.SetMaxIdleConns(5)
   d.db = sqlDB
   d.serverName = serverName
   d.connectionCtx = config.ConnectionContext
   return d, nil
   ```

3. **Add the dependency:**
   ```bash
   go get github.com/alexbrainman/odbc
   ```

4. **Rebuild the backend:**
   ```bash
   go build -ldflags "-w -s" -p=16 -o ./bytebase-build/bytebase ./backend/bin/server/main.go
   ```

## Connection Configuration

The driver expects these connection parameters:

- **Host:** Informix server hostname/IP
- **Port:** Default 9088 (configurable)
- **Username/Password:** Database credentials
- **Database:** Target database name
- **SERVER:** Informix server name (configurable via extra parameters)
- **PROTOCOL:** Default "onsoctcp" (configurable via extra parameters)

### Example DSN Format
```
DRIVER={IBM INFORMIX ODBC DRIVER};HOST=localhost;SERVICE=9088;UID=username;PWD=password;SERVER=informix;DATABASE=testdb;PROTOCOL=onsoctcp
```

## Features Supported

- ✅ Connection management
- ✅ Database metadata synchronization
- ✅ Schema introspection
- ✅ Table, column, index metadata
- ✅ SSL connection support
- ✅ SSH tunneling support
- ✅ Extra connection parameters
- ✅ Schema dumping framework
- ✅ Query execution framework

## Alternative Driver Options

If ODBC setup is problematic, consider these alternatives:

1. **Native Go driver** (if available for Informix)
2. **CGO wrapper** around Informix client libraries
3. **Network protocol implementation** (requires reverse engineering)

## Testing

Once set up, you can test the connection in Bytebase:

1. Go to `http://localhost:3000/instances`
2. Click "Create Instance"
3. Select "Informix" from the engine dropdown
4. Configure connection parameters
5. Test the connection

The driver includes comprehensive error handling and logging to help diagnose connection issues.