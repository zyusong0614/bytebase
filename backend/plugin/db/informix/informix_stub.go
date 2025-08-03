//go:build !informix
// +build !informix

package informix

import (
	"context"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/pkg/errors"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/plugin/db"
)

// Open opens an Informix driver (stub implementation)
func (d *Driver) Open(ctx context.Context, _ storepb.Engine, config db.ConnectionConfig) (db.Driver, error) {
	port, err := strconv.Atoi(config.DataSource.Port)
	if err != nil {
		return nil, errors.Errorf("invalid port %q", config.DataSource.Port)
	}

	// Log connection attempt
	slog.Info("Informix driver stub - connection attempted",
		"host", config.DataSource.Host,
		"port", port,
		"database", config.ConnectionContext.DatabaseName,
		"note", "This is a stub implementation. Use Docker deployment with -tags informix for full support")

	return nil, errors.New("Informix support requires Docker deployment with IBM Informix Client SDK. Please use docker-compose.informix.yml for deployment")
}

// Close closes the driver
func (d *Driver) Close(ctx context.Context) error {
	return nil
}

// Ping pings the database
func (d *Driver) Ping(ctx context.Context) error {
	return errors.New("Informix driver not available in this build. Use Docker deployment for Informix support")
}

// GetDB returns nil
func (d *Driver) GetDB() *sql.DB {
	return nil
}

// Execute executes a SQL statement
func (d *Driver) Execute(ctx context.Context, statement string, opts db.ExecuteOptions) (int64, error) {
	return 0, errors.New("Informix driver not available in this build. Use Docker deployment for Informix support")
}

// QueryConn queries a SQL statement
func (d *Driver) QueryConn(ctx context.Context, conn *sql.Conn, statement string, queryContext db.QueryContext) ([]*v1pb.QueryResult, error) {
	result := &v1pb.QueryResult{
		Statement: statement,
		Error:     "Informix driver not available in this build. Use Docker deployment for Informix support",
	}
	return []*v1pb.QueryResult{result}, nil
}