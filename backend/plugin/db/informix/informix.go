// Package informix is the plugin for IBM Informix driver.
package informix

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	// Import OpenInformix IfxGo driver - specialized Informix ODBC driver
	_ "github.com/OpenInformix/IfxGo"
	
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/bytebase/bytebase/backend/common/log"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/plugin/db"
)

var (
	_ db.Driver = (*Driver)(nil)
)

func init() {
	db.Register(storepb.Engine_INFORMIX, newDriver)
}

// Driver is the Informix driver.
type Driver struct {
	db            *sql.DB
	databaseName  string
	serverName    string
	connectionCtx db.ConnectionContext
}

func newDriver() db.Driver {
	return &Driver{}
}

// Open opens an Informix driver.
func (d *Driver) Open(ctx context.Context, _ storepb.Engine, config db.ConnectionConfig) (db.Driver, error) {
	port, err := strconv.Atoi(config.DataSource.Port)
	if err != nil {
		return nil, errors.Errorf("invalid port %q", config.DataSource.Port)
	}

	// Build Informix ODBC connection string
	// Format: DSN=name;SERVER=server;HOST=host;SERVICE=port;DATABASE=db;UID=user;PWD=password
	var connectionParams []string
	
	// Required parameters
	connectionParams = append(connectionParams, fmt.Sprintf("HOST=%s", config.DataSource.Host))
	connectionParams = append(connectionParams, fmt.Sprintf("SERVICE=%d", port))
	connectionParams = append(connectionParams, fmt.Sprintf("UID=%s", config.DataSource.Username))
	connectionParams = append(connectionParams, fmt.Sprintf("PWD=%s", config.Password))
	
	// Server name from extra parameters or default
	serverName := config.DataSource.GetExtraConnectionParameters()["SERVER"]
	if serverName == "" {
		serverName = "informix"
	}
	connectionParams = append(connectionParams, fmt.Sprintf("SERVER=%s", serverName))
	
	// Database name
	if config.ConnectionContext.DatabaseName != "" {
		connectionParams = append(connectionParams, fmt.Sprintf("DATABASE=%s", config.ConnectionContext.DatabaseName))
		d.databaseName = config.ConnectionContext.DatabaseName
	}
	
	// Protocol (default to onsoctcp)
	protocol := config.DataSource.GetExtraConnectionParameters()["PROTOCOL"]
	if protocol == "" {
		protocol = "onsoctcp"
	}
	connectionParams = append(connectionParams, fmt.Sprintf("PROTOCOL=%s", protocol))
	
	// Additional connection parameters
	for key, value := range config.DataSource.GetExtraConnectionParameters() {
		if key != "SERVER" && key != "PROTOCOL" {
			connectionParams = append(connectionParams, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	dsn := "DRIVER={IBM INFORMIX ODBC DRIVER};" + strings.Join(connectionParams, ";")
	
	slog.Debug("Connecting to Informix", "dsn", dsn)
	
	// This is a placeholder implementation for Informix driver
	// To enable full Informix support, you need to:
	//
	// 1. Install ODBC development libraries:
	//    - Ubuntu/Debian: sudo apt-get install unixodbc-dev
	//    - macOS: brew install libiodbc
	//    - RHEL/CentOS: yum install unixODBC-devel
	//
	// 2. Install IBM Informix ODBC driver:
	//    - Download from IBM Informix downloads page
	//    - Install the driver and configure DSN
	//
	// 3. Update the import statement to include ODBC driver:
	//    - Uncomment: _ "github.com/alexbrainman/odbc"
	//
	// 4. Replace this error with actual ODBC connection:
	//    sqlDB, err := sql.Open("odbc", dsn)
	//    if err != nil {
	//        return nil, errors.Wrapf(err, "failed to open Informix connection")
	//    }
	//    sqlDB.SetConnMaxLifetime(10 * time.Minute)
	//    sqlDB.SetMaxOpenConns(10)
	//    sqlDB.SetMaxIdleConns(5)
	//    d.db = sqlDB
	//    d.serverName = serverName
	//    d.connectionCtx = config.ConnectionContext
	//    return d, nil
	
	// 使用 OpenInformix IfxGo 驱动连接
	slog.Debug("Connecting to Informix using IfxGo", "dsn", dsn)
	
	sqlDB, err := sql.Open("odbc", dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open Informix connection with DSN: %s", dsn)
	}
	
	// 设置连接池参数
	sqlDB.SetConnMaxLifetime(10 * time.Minute)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	
	// 测试连接
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, errors.Wrapf(err, "failed to ping Informix database")
	}
	
	d.db = sqlDB
	d.serverName = serverName
	d.connectionCtx = config.ConnectionContext
	
	slog.Info("Successfully connected to Informix", "server", serverName, "database", d.databaseName)
	return d, nil
}

// Close closes the driver.
func (d *Driver) Close(ctx context.Context) error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Ping pings the database.
func (d *Driver) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// GetDB returns the underlying sql.DB.
func (d *Driver) GetDB() *sql.DB {
	return d.db
}

// Execute executes a SQL statement.
func (d *Driver) Execute(ctx context.Context, statement string, opts db.ExecuteOptions) (int64, error) {
	if opts.CreateDatabase {
		return 0, errors.New("CREATE DATABASE is not supported for Informix via this driver")
	}

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	totalRowsAffected := int64(0)
	
	// Split statement by semicolon and execute each one
	// For now, use simple semicolon split since Informix parser is not implemented
	statements := splitInformixStatements(statement)
	for _, stmt := range statements {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		
		result, err := tx.ExecContext(ctx, stmt)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to execute statement: %s", stmt)
		}
		
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			slog.Warn("failed to get rows affected", log.BBError(err))
		} else {
			totalRowsAffected += rowsAffected
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.Wrapf(err, "failed to commit transaction")
	}

	return totalRowsAffected, nil
}

// QueryConn queries a SQL statement in a given connection.
func (d *Driver) QueryConn(ctx context.Context, conn *sql.Conn, statement string, queryContext db.QueryContext) ([]*v1pb.QueryResult, error) {
	// For now, use simple statement splitting since Informix parser is not implemented
	statements := splitInformixStatements(statement)
	var singleSQLs []struct {
		Text string
	}
	for _, stmt := range statements {
		if strings.TrimSpace(stmt) != "" {
			singleSQLs = append(singleSQLs, struct{ Text string }{Text: stmt})
		}
	}

	var results []*v1pb.QueryResult
	for _, singleSQL := range singleSQLs {
		statement := strings.TrimLeft(singleSQL.Text, " \n\t")
		if statement == "" {
			continue
		}

		result, err := d.querySingleSQL(ctx, conn, statement, queryContext)
		if err != nil {
			result = &v1pb.QueryResult{
				Error: err.Error(),
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func (d *Driver) querySingleSQL(ctx context.Context, conn *sql.Conn, statement string, queryContext db.QueryContext) (*v1pb.QueryResult, error) {
	startTime := time.Now()
	result := &v1pb.QueryResult{
		Statement: statement,
	}

	// Handle EXPLAIN queries
	if queryContext.Explain {
		statement = fmt.Sprintf("SET EXPLAIN ON; %s", statement)
	}

	rows, err := conn.QueryContext(ctx, statement)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to execute query")
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get column names")
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get column types")
	}

	// Build column metadata
	for _, name := range columnNames {
		result.ColumnNames = append(result.ColumnNames, name)
	}
	
	for _, columnType := range columnTypes {
		result.ColumnTypeNames = append(result.ColumnTypeNames, columnType.DatabaseTypeName())
	}

	// Read rows
	rowCount := 0
	for rows.Next() {
		if queryContext.Limit > 0 && rowCount >= queryContext.Limit {
			// Note: Warning field might not be available in this version
			// result.Warning = fmt.Sprintf("Output truncated to %d rows", queryContext.Limit)
			break
		}

		values := make([]any, len(columnNames))
		scanArgs := make([]any, len(values))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, errors.Wrapf(err, "failed to scan row")
		}

		var row v1pb.QueryRow
		for _, value := range values {
			var v *v1pb.RowValue
			if value == nil {
				v = &v1pb.RowValue{Kind: &v1pb.RowValue_NullValue{}}
			} else {
				switch val := value.(type) {
				case string:
					v = &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: val}}
				case []byte:
					v = &v1pb.RowValue{Kind: &v1pb.RowValue_BytesValue{BytesValue: val}}
				case int64:
					v = &v1pb.RowValue{Kind: &v1pb.RowValue_Int64Value{Int64Value: val}}
				case float64:
					v = &v1pb.RowValue{Kind: &v1pb.RowValue_DoubleValue{DoubleValue: val}}
				case bool:
					v = &v1pb.RowValue{Kind: &v1pb.RowValue_BoolValue{BoolValue: val}}
				case time.Time:
					v = &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: val.Format(time.RFC3339)}}
				default:
					v = &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: fmt.Sprintf("%v", val)}}
				}
			}
			row.Values = append(row.Values, v)
		}
		result.Rows = append(result.Rows, &row)
		rowCount++
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrapf(err, "failed to iterate rows")
	}

	result.Latency = durationpb.New(time.Since(startTime))
	return result, nil
}

// Note: SyncInstance and SyncDBSchema are implemented in sync.go

// Note: Dump is implemented in dump.go

// splitInformixStatements is a simple statement splitter for Informix.
// This is a temporary implementation until a proper Informix parser is added.
func splitInformixStatements(statement string) []string {
	// Simple semicolon-based splitting
	// TODO: Implement proper parsing that handles quotes, comments, etc.
	statements := strings.Split(statement, ";")
	var result []string
	
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}