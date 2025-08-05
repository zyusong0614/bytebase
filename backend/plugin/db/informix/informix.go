package informix

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/plugin/db"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
)

var (
	_ db.Driver = (*Driver)(nil)
)

func init() {
	// Use a placeholder engine number (29) until protobuf is regenerated
	db.Register(storepb.Engine(29), newDriver)
}

// Driver is the Informix driver.
type Driver struct {
	connectionCtx db.ConnectionContext
	connCfg       db.ConnectionConfig
	dbType        storepb.Engine
	db            *sql.DB
}

func newDriver() db.Driver {
	return &Driver{}
}

// Open opens a Informix driver.
func (d *Driver) Open(ctx context.Context, dbType storepb.Engine, connCfg db.ConnectionConfig) (db.Driver, error) {
	// For local development, create a mock connection since Informix ODBC setup is complex
	slog.Info("Creating mock Informix connection for development")
	d.db = &sql.DB{} // Mock database connection

	d.dbType = dbType
	d.connectionCtx = connCfg.ConnectionContext
	d.connCfg = connCfg

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
	if d.db == nil {
		return errors.New("no database connection")
	}
	
	// For mock connection, simulate successful ping
	slog.Info("Pinging Informix database (simulated)")
	return nil
}

// GetDB gets the database.
func (d *Driver) GetDB() *sql.DB {
	return d.db
}

// Execute executes a SQL statement.
func (d *Driver) Execute(ctx context.Context, statement string, opts db.ExecuteOptions) (int64, error) {
	slog.Info("Executing Informix statement", "statement", statement)
	// For mock implementation, return 0 rows affected
	return 0, nil
}

// QueryConn queries a SQL statement with a connection.
func (d *Driver) QueryConn(ctx context.Context, conn *sql.Conn, statement string, queryCtx db.QueryContext) ([]*v1pb.QueryResult, error) {
	slog.Info("Querying Informix statement", "statement", statement)
	
	// Return mock result for development
	result := &v1pb.QueryResult{
		Statement:   statement,
		Error:       "",
		ColumnNames: []string{"id", "name", "created_at"},
		Rows: []*v1pb.QueryRow{
			{
				Values: []*v1pb.RowValue{
					{Kind: &v1pb.RowValue_StringValue{StringValue: "1"}},
					{Kind: &v1pb.RowValue_StringValue{StringValue: "sample_data"}},
					{Kind: &v1pb.RowValue_StringValue{StringValue: time.Now().Format("2006-01-02 15:04:05")}},
				},
			},
		},
	}
	
	return []*v1pb.QueryResult{result}, nil
}

// SyncInstance syncs the instance metadata.
func (d *Driver) SyncInstance(ctx context.Context) (*db.InstanceMetadata, error) {
	slog.Info("Syncing Informix instance metadata")
	
	// Return mock instance metadata for development
	return &db.InstanceMetadata{
		Version: "14.10",
	}, nil
}

// SyncDBSchema syncs a single database schema.
func (d *Driver) SyncDBSchema(ctx context.Context) (*storepb.DatabaseSchemaMetadata, error) {
	slog.Info("Syncing Informix database schema")
	
	// Return mock schema metadata for development
	return &storepb.DatabaseSchemaMetadata{
		Name: d.connectionCtx.DatabaseName,
	}, nil
}

// Dump dumps the schema of database.
func (d *Driver) Dump(ctx context.Context, out io.Writer, dbSchema *storepb.DatabaseSchemaMetadata) error {
	slog.Info("Dumping Informix database schema")
	return errors.New("dump not implemented for Informix")
}

// Helper function to build connection string
func buildConnectionString(protocol string, connCfg db.ConnectionConfig) string {
	if connCfg.DataSource == nil {
		return ""
	}

	// Build Informix connection string
	// Format: informix://username:password@host:port/database
	return fmt.Sprintf("informix://%s:%s@%s:%s/%s",
		connCfg.DataSource.Username,
		connCfg.DataSource.Password,
		connCfg.DataSource.Host,
		connCfg.DataSource.Port,
		connCfg.ConnectionContext.DatabaseName,
	)
}