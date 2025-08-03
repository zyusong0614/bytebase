// Package informix is the plugin for IBM Informix driver.
package informix

import (
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
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

