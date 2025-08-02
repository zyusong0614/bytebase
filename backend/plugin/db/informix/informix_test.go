package informix

import (
	"context"
	"testing"

	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	"github.com/bytebase/bytebase/backend/plugin/db"
)

func TestDriver_Register(t *testing.T) {
	// Test that the driver is properly registered
	driver, err := db.Open(
		context.Background(),
		storepb.Engine_INFORMIX,
		db.ConnectionConfig{
			DataSource: &storepb.DataSource{
				Host:     "localhost",
				Port:     "9088",
				Username: "informix",
			},
			ConnectionContext: db.ConnectionContext{
				DatabaseName: "test",
			},
			Password: "informix",
		},
	)

	// Since this is a placeholder implementation, we expect it to fail with our specific error
	if err == nil {
		t.Error("Expected error due to placeholder implementation, but got nil")
		if driver != nil {
			driver.Close(context.Background())
		}
		return
	}

	expectedError := "Informix driver implementation requires a proper database driver library (e.g., ODBC driver). This is a placeholder implementation for demonstration purposes"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestSplitInformixStatements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single statement",
			input:    "SELECT * FROM users",
			expected: []string{"SELECT * FROM users"},
		},
		{
			name:     "Multiple statements",
			input:    "SELECT * FROM users; INSERT INTO logs VALUES (1, 'test'); UPDATE users SET name = 'John'",
			expected: []string{"SELECT * FROM users", "INSERT INTO logs VALUES (1, 'test')", "UPDATE users SET name = 'John'"},
		},
		{
			name:     "Empty statements",
			input:    "SELECT * FROM users;; INSERT INTO logs VALUES (1, 'test');",
			expected: []string{"SELECT * FROM users", "INSERT INTO logs VALUES (1, 'test')"},
		},
		{
			name:     "Whitespace handling",
			input:    "  SELECT * FROM users  ;  INSERT INTO logs VALUES (1, 'test')  ;  ",
			expected: []string{"SELECT * FROM users", "INSERT INTO logs VALUES (1, 'test')"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitInformixStatements(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d statements, got %d", len(tt.expected), len(result))
				return
			}
			for i, stmt := range result {
				if stmt != tt.expected[i] {
					t.Errorf("Statement %d: expected '%s', got '%s'", i, tt.expected[i], stmt)
				}
			}
		})
	}
}

func TestConvertInformixType(t *testing.T) {
	tests := []struct {
		name     string
		colType  int
		expected string
	}{
		{
			name:     "CHAR",
			colType:  0,
			expected: "CHAR",
		},
		{
			name:     "SMALLINT",
			colType:  1,
			expected: "SMALLINT",
		},
		{
			name:     "INTEGER",
			colType:  2,
			expected: "INTEGER",
		},
		{
			name:     "VARCHAR",
			colType:  12,
			expected: "VARCHAR",
		},
		{
			name:     "NULLABLE CHAR",
			colType:  256, // 0 + 256 (nullable flag)
			expected: "CHAR",
		},
		{
			name:     "NULLABLE INTEGER",
			colType:  258, // 2 + 256 (nullable flag)
			expected: "INTEGER",
		},
		{
			name:     "Unknown type",
			colType:  999,
			expected: "UNKNOWN(231)", // 999 % 256 = 231
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertInformixType(tt.colType)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNewDriver(t *testing.T) {
	driver := newDriver()
	if driver == nil {
		t.Error("newDriver() returned nil")
	}

	// Verify it implements the Driver interface
	var _ db.Driver = driver
}

// TestDriverInterface verifies that our Driver implements all required methods
func TestDriverInterface(t *testing.T) {
	driver := &Driver{}

	// Test that all interface methods exist and can be called
	// (Even though they might return errors due to placeholder implementation)

	ctx := context.Background()

	// Test Close with nil db (should not panic)
	err := driver.Close(ctx)
	if err != nil {
		// Expected for nil db
	}

	// Test GetDB with nil db
	db := driver.GetDB()
	if db != nil {
		t.Error("Expected nil db for uninitialized driver")
	}

	// Test SyncInstance
	_, err = driver.SyncInstance(ctx)
	if err != nil {
		// Expected for placeholder implementation
	}

	// Test SyncDBSchema
	_, err = driver.SyncDBSchema(ctx)
	if err != nil {
		// Expected for placeholder implementation without database name
	}
}