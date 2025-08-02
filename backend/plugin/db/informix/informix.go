// Package informix is the plugin for IBM Informix driver.
package informix

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/pkg/errors"

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
	connectionString string
	databaseName     string
	serverName       string
	connectionCtx    db.ConnectionContext
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

	// Build Informix connection string
	var connectionParams []string
	
	// Required parameters for Informix ODBC
	connectionParams = append(connectionParams, "Driver={IBM INFORMIX ODBC DRIVER}")
	
	// Server name from extra parameters or default
	serverName := config.DataSource.GetExtraConnectionParameters()["SERVER"]
	if serverName == "" {
		serverName = "informix"
	}
	connectionParams = append(connectionParams, fmt.Sprintf("Servername=%s", serverName))
	
	// Basic connection info
	connectionParams = append(connectionParams, fmt.Sprintf("Host=%s", config.DataSource.Host))
	connectionParams = append(connectionParams, fmt.Sprintf("Service=%d", port))
	connectionParams = append(connectionParams, fmt.Sprintf("UID=%s", config.DataSource.Username))
	connectionParams = append(connectionParams, fmt.Sprintf("PWD=%s", config.Password))
	
	// Database name
	if config.ConnectionContext.DatabaseName != "" {
		connectionParams = append(connectionParams, fmt.Sprintf("Database=%s", config.ConnectionContext.DatabaseName))
		d.databaseName = config.ConnectionContext.DatabaseName
	}
	
	// Protocol (default to onsoctcp)
	protocol := config.DataSource.GetExtraConnectionParameters()["PROTOCOL"]
	if protocol == "" {
		protocol = "onsoctcp"
	}
	connectionParams = append(connectionParams, fmt.Sprintf("Protocol=%s", protocol))
	
	// Additional connection parameters
	for key, value := range config.DataSource.GetExtraConnectionParameters() {
		if key != "SERVER" && key != "PROTOCOL" {
			connectionParams = append(connectionParams, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	connString := strings.Join(connectionParams, ";")
	d.connectionString = connString
	d.serverName = serverName
	d.connectionCtx = config.ConnectionContext
	
	slog.Info("Informix driver initialized", "server", serverName, "database", d.databaseName, "connectionString", connString)
	
	// For now, return success to test ENGINE_UNSPECIFIED fix
	// Real ODBC connection will be implemented with proper IBM SDK setup
	return d, nil
}

// Close closes the driver.
func (d *Driver) Close(ctx context.Context) error {
	return nil
}

// Ping pings the database.
func (d *Driver) Ping(ctx context.Context) error {
	// For now, return an informative error instead of actual ping
	return errors.New("Informix driver loaded successfully, but IBM Informix Client SDK is required for actual connections. Please run setup_informix_driver.sh")
}

// GetDB returns nil as we don't have a standard sql.DB connection.
func (d *Driver) GetDB() *sql.DB {
	return nil
}

// Execute executes a SQL statement.
func (d *Driver) Execute(ctx context.Context, statement string, opts db.ExecuteOptions) (int64, error) {
	if opts.CreateDatabase {
		return 0, errors.New("CREATE DATABASE is not supported for Informix via this driver")
	}

	return 0, errors.New("IBM Informix Client SDK is required for SQL execution. Please run setup_informix_driver.sh")
}

// QueryConn queries a SQL statement.
func (d *Driver) QueryConn(ctx context.Context, conn *sql.Conn, statement string, queryContext db.QueryContext) ([]*v1pb.QueryResult, error) {
	result := &v1pb.QueryResult{
		Statement: statement,
		Error:     "IBM Informix Client SDK is required for queries. Please run setup_informix_driver.sh",
	}
	
	return []*v1pb.QueryResult{result}, nil
}