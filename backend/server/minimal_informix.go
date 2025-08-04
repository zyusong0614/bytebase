//go:build informix

package server

import (
	// Core drivers only: PostgreSQL (for metadata) + Informix (target database)
	_ "github.com/bytebase/bytebase/backend/plugin/db/pg"
	_ "github.com/bytebase/bytebase/backend/plugin/db/informix"

	// Core parsers
	_ "github.com/bytebase/bytebase/backend/plugin/parser/pg"
)