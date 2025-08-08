//go:build !informix
// +build !informix

package informix

import (
	"context"

	"github.com/pkg/errors"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/plugin/db"
)

// openODBC stub for when informix build tag is not present
func (d *Driver) openODBC(ctx context.Context, dbType storepb.Engine, config db.ConnectionConfig) (*ODBCDriver, error) {
	return nil, errors.New("Informix ODBC driver not available - compile with -tags informix")
}