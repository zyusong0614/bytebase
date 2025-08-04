//go:build informix
// +build informix

package informix

/*
#cgo LDFLAGS: -lodbc
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sql.h>
#include <sqlext.h>

typedef struct {
    SQLHENV env;
    SQLHDBC dbc;
    int connected;
    char error[1024];
} ODBCConnection;

ODBCConnection* odbc_connect(const char* connStr) {
    ODBCConnection* conn = (ODBCConnection*)malloc(sizeof(ODBCConnection));
    conn->connected = 0;
    conn->error[0] = '\0';
    
    SQLRETURN ret;
    
    // Allocate environment handle
    ret = SQLAllocHandle(SQL_HANDLE_ENV, SQL_NULL_HANDLE, &conn->env);
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        strcpy(conn->error, "Failed to allocate environment handle");
        return conn;
    }
    
    // Set ODBC version
    ret = SQLSetEnvAttr(conn->env, SQL_ATTR_ODBC_VERSION, (void*)SQL_OV_ODBC3, 0);
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        strcpy(conn->error, "Failed to set ODBC version");
        SQLFreeHandle(SQL_HANDLE_ENV, conn->env);
        return conn;
    }
    
    // Allocate connection handle
    ret = SQLAllocHandle(SQL_HANDLE_DBC, conn->env, &conn->dbc);
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        strcpy(conn->error, "Failed to allocate connection handle");
        SQLFreeHandle(SQL_HANDLE_ENV, conn->env);
        return conn;
    }
    
    // Connect
    ret = SQLDriverConnect(conn->dbc, NULL, (SQLCHAR*)connStr, SQL_NTS, NULL, 0, NULL, SQL_DRIVER_NOPROMPT);
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        SQLCHAR sqlState[6], errorMsg[1024];
        SQLINTEGER nativeError;
        SQLSMALLINT textLength;
        SQLGetDiagRec(SQL_HANDLE_DBC, conn->dbc, 1, sqlState, &nativeError, errorMsg, sizeof(errorMsg), &textLength);
        snprintf(conn->error, sizeof(conn->error), "Connection failed: %s", errorMsg);
        SQLFreeHandle(SQL_HANDLE_DBC, conn->dbc);
        SQLFreeHandle(SQL_HANDLE_ENV, conn->env);
        return conn;
    }
    
    conn->connected = 1;
    return conn;
}

void odbc_disconnect(ODBCConnection* conn) {
    if (conn && conn->connected) {
        SQLDisconnect(conn->dbc);
        SQLFreeHandle(SQL_HANDLE_DBC, conn->dbc);
        SQLFreeHandle(SQL_HANDLE_ENV, conn->env);
        conn->connected = 0;
    }
    if (conn) {
        free(conn);
    }
}

int odbc_execute(ODBCConnection* conn, const char* sql, char* error) {
    if (!conn || !conn->connected) {
        strcpy(error, "Not connected");
        return -1;
    }
    
    SQLHSTMT stmt;
    SQLRETURN ret;
    
    ret = SQLAllocHandle(SQL_HANDLE_STMT, conn->dbc, &stmt);
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        strcpy(error, "Failed to allocate statement handle");
        return -1;
    }
    
    ret = SQLExecDirect(stmt, (SQLCHAR*)sql, SQL_NTS);
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        SQLCHAR sqlState[6], errorMsg[1024];
        SQLINTEGER nativeError;
        SQLSMALLINT textLength;
        SQLGetDiagRec(SQL_HANDLE_STMT, stmt, 1, sqlState, &nativeError, errorMsg, sizeof(errorMsg), &textLength);
        snprintf(error, 1024, "Execute failed: %s", errorMsg);
        SQLFreeHandle(SQL_HANDLE_STMT, stmt);
        return -1;
    }
    
    SQLLEN rowCount;
    SQLRowCount(stmt, &rowCount);
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
    
    return (int)rowCount;
}

// Simple query result structure
typedef struct {
    char** data;
    int rows;
    int cols;
    char* colNames;
    char error[1024];
} QueryResult;

QueryResult* odbc_query(ODBCConnection* conn, const char* sql) {
    QueryResult* result = (QueryResult*)malloc(sizeof(QueryResult));
    result->data = NULL;
    result->rows = 0;
    result->cols = 0;
    result->colNames = NULL;
    result->error[0] = '\0';
    
    if (!conn || !conn->connected) {
        strcpy(result->error, "Not connected");
        return result;
    }
    
    SQLHSTMT stmt;
    SQLRETURN ret;
    
    ret = SQLAllocHandle(SQL_HANDLE_STMT, conn->dbc, &stmt);
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        strcpy(result->error, "Failed to allocate statement handle");
        return result;
    }
    
    ret = SQLExecDirect(stmt, (SQLCHAR*)sql, SQL_NTS);
    if (ret != SQL_SUCCESS && ret != SQL_SUCCESS_WITH_INFO) {
        SQLCHAR sqlState[6], errorMsg[1024];
        SQLINTEGER nativeError;
        SQLSMALLINT textLength;
        SQLGetDiagRec(SQL_HANDLE_STMT, stmt, 1, sqlState, &nativeError, errorMsg, sizeof(errorMsg), &textLength);
        snprintf(result->error, sizeof(result->error), "Query failed: %s", errorMsg);
        SQLFreeHandle(SQL_HANDLE_STMT, stmt);
        return result;
    }
    
    // Get column count
    SQLSMALLINT colCount;
    SQLNumResultCols(stmt, &colCount);
    result->cols = colCount;
    
    // For simplicity, we'll just count rows for now
    // In a real implementation, you'd fetch all data
    SQLLEN rowCount = 0;
    while (SQLFetch(stmt) == SQL_SUCCESS) {
        rowCount++;
    }
    result->rows = (int)rowCount;
    
    SQLFreeHandle(SQL_HANDLE_STMT, stmt);
    return result;
}

void free_query_result(QueryResult* result) {
    if (result) {
        if (result->data) {
            free(result->data);
        }
        if (result->colNames) {
            free(result->colNames);
        }
        free(result);
    }
}
*/
import "C"
import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"unsafe"

	"github.com/pkg/errors"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/plugin/db"
)

// ODBCDriver is the Informix driver using ODBC
type ODBCDriver struct {
	connectionString string
	databaseName     string
	conn             *C.ODBCConnection
	connectionCtx    db.ConnectionContext
}

// Open opens an Informix driver using ODBC
func (d *Driver) Open(ctx context.Context, _ storepb.Engine, config db.ConnectionConfig) (db.Driver, error) {
	port, err := strconv.Atoi(config.DataSource.Port)
	if err != nil {
		return nil, errors.Errorf("invalid port %q", config.DataSource.Port)
	}

	// Build ODBC connection string
	var connectionParams []string
	connectionParams = append(connectionParams, "Driver={IBM INFORMIX ODBC DRIVER}")
	
	// Server name from extra parameters or default
	serverName := config.DataSource.GetExtraConnectionParameters()["SERVER"]
	if serverName == "" {
		serverName = "informix_tcp"
	}
	connectionParams = append(connectionParams, fmt.Sprintf("Servername=%s", serverName))
	
	// Basic connection info
	connectionParams = append(connectionParams, fmt.Sprintf("Host=%s", config.DataSource.Host))
	connectionParams = append(connectionParams, fmt.Sprintf("Service=%d", port))
	connectionParams = append(connectionParams, fmt.Sprintf("UID=%s", config.DataSource.Username))
	connectionParams = append(connectionParams, fmt.Sprintf("PWD=%s", config.Password))
	
	// Database name
	database := config.ConnectionContext.DatabaseName
	if database == "" {
		database = "sysmaster" // Default system database
	}
	connectionParams = append(connectionParams, fmt.Sprintf("Database=%s", database))
	
	// Protocol
	protocol := config.DataSource.GetExtraConnectionParameters()["PROTOCOL"]
	if protocol == "" {
		protocol = "onsoctcp"
	}
	connectionParams = append(connectionParams, fmt.Sprintf("Protocol=%s", protocol))
	
	connString := strings.Join(connectionParams, ";")
	
	// Connect using ODBC
	cConnStr := C.CString(connString)
	defer C.free(unsafe.Pointer(cConnStr))
	
	odbcConn := C.odbc_connect(cConnStr)
	if odbcConn.connected == 0 {
		errorMsg := C.GoString(&odbcConn.error[0])
		C.free(unsafe.Pointer(odbcConn))
		return nil, errors.Errorf("failed to connect to Informix: %s", errorMsg)
	}
	
	driver := &ODBCDriver{
		connectionString: connString,
		databaseName:     database,
		conn:             odbcConn,
		connectionCtx:    config.ConnectionContext,
	}
	
	slog.Info("Informix ODBC driver connected", 
		"server", serverName, 
		"database", database,
		"host", config.DataSource.Host,
		"port", port)
	
	// Return the base Driver with embedded ODBCDriver
	d.connectionString = connString
	d.databaseName = database
	d.connectionCtx = config.ConnectionContext
	d.odbcDriver = driver
	
	return d, nil
}

// Close closes the driver
func (d *ODBCDriver) Close(ctx context.Context) error {
	if d.conn != nil {
		C.odbc_disconnect(d.conn)
		d.conn = nil
	}
	return nil
}

// Ping pings the database
func (d *ODBCDriver) Ping(ctx context.Context) error {
	if d.conn == nil || d.conn.connected == 0 {
		return errors.New("not connected to database")
	}
	
	// Simple ping query
	cSQL := C.CString("SELECT 1 FROM systables WHERE tabid = 1")
	defer C.free(unsafe.Pointer(cSQL))
	
	result := C.odbc_query(d.conn, cSQL)
	defer C.free_query_result(result)
	
	if result.error[0] != 0 {
		errorMsg := C.GoString(&result.error[0])
		return errors.Errorf("ping failed: %s", errorMsg)
	}
	
	return nil
}

// GetDB returns nil as we're using ODBC, not sql.DB
func (d *ODBCDriver) GetDB() *sql.DB {
	return nil
}

// Execute executes a SQL statement
func (d *ODBCDriver) Execute(ctx context.Context, statement string, opts db.ExecuteOptions) (int64, error) {
	if d.conn == nil || d.conn.connected == 0 {
		return 0, errors.New("not connected to database")
	}
	
	if opts.CreateDatabase {
		// For CREATE DATABASE, use the database name from connection context
		dbName := d.connectionCtx.DatabaseName
		if dbName == "" {
			dbName = "testdb" // Default name
		}
		statement = fmt.Sprintf("CREATE DATABASE %s", dbName)
	}
	
	cSQL := C.CString(statement)
	defer C.free(unsafe.Pointer(cSQL))
	
	var cError [1024]C.char
	rowsAffected := C.odbc_execute(d.conn, cSQL, &cError[0])
	
	if rowsAffected < 0 {
		errorMsg := C.GoString(&cError[0])
		return 0, errors.Errorf("execute failed: %s", errorMsg)
	}
	
	return int64(rowsAffected), nil
}

// QueryConn queries a SQL statement
func (d *ODBCDriver) QueryConn(ctx context.Context, conn *sql.Conn, statement string, queryContext db.QueryContext) ([]*v1pb.QueryResult, error) {
	if d.conn == nil || d.conn.connected == 0 {
		return nil, errors.New("not connected to database")
	}
	
	cSQL := C.CString(statement)
	defer C.free(unsafe.Pointer(cSQL))
	
	queryResult := C.odbc_query(d.conn, cSQL)
	defer C.free_query_result(queryResult)
	
	if queryResult.error[0] != 0 {
		errorMsg := C.GoString(&queryResult.error[0])
		result := &v1pb.QueryResult{
			Statement: statement,
			Error:     errorMsg,
		}
		return []*v1pb.QueryResult{result}, nil
	}
	
	// For now, return basic result info
	// A complete implementation would fetch all data
	result := &v1pb.QueryResult{
		Statement: statement,
		// Note: Full result fetching would be implemented here
		Latency: nil, // Would measure actual latency as durationpb
	}
	
	slog.Debug("Query executed via ODBC", 
		"rows", queryResult.rows,
		"cols", queryResult.cols)
	
	return []*v1pb.QueryResult{result}, nil
}

// Driver methods that delegate to ODBCDriver

// Close closes the driver
func (d *Driver) Close(ctx context.Context) error {
	if d.odbcDriver != nil {
		return d.odbcDriver.Close(ctx)
	}
	return nil
}

// Ping pings the database
func (d *Driver) Ping(ctx context.Context) error {
	if d.odbcDriver != nil {
		return d.odbcDriver.Ping(ctx)
	}
	return errors.New("ODBC driver not initialized")
}

// GetDB returns nil as we're using ODBC, not sql.DB
func (d *Driver) GetDB() *sql.DB {
	if d.odbcDriver != nil {
		return d.odbcDriver.GetDB()
	}
	return nil
}

// Execute executes a SQL statement
func (d *Driver) Execute(ctx context.Context, statement string, opts db.ExecuteOptions) (int64, error) {
	if d.odbcDriver != nil {
		return d.odbcDriver.Execute(ctx, statement, opts)
	}
	return 0, errors.New("ODBC driver not initialized")
}

// QueryConn queries a SQL statement
func (d *Driver) QueryConn(ctx context.Context, conn *sql.Conn, statement string, queryContext db.QueryContext) ([]*v1pb.QueryResult, error) {
	if d.odbcDriver != nil {
		return d.odbcDriver.QueryConn(ctx, conn, statement, queryContext)
	}
	result := &v1pb.QueryResult{
		Statement: statement,
		Error:     "ODBC driver not initialized",
	}
	return []*v1pb.QueryResult{result}, nil
}