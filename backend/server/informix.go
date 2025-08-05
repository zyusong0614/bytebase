//go:build informix
// +build informix

package server

import (
	// Informix driver
	_ "github.com/bytebase/bytebase/backend/plugin/db/informix"
)