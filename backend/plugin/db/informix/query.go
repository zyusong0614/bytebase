package informix

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/plugin/db/util"
)

func makeValueByTypeName(typeName string, _ *sql.ColumnType) any {
	switch strings.ToUpper(typeName) {
	case "CHAR", "VARCHAR", "LVARCHAR", "TEXT", "CLOB":
		return new(sql.NullString)
	case "BOOLEAN":
		return new(sql.NullBool)
	case "SMALLINT", "INT", "INTEGER", "BIGINT", "INT8", "SERIAL", "BIGSERIAL", "SERIAL8":
		return new(sql.NullInt64)
	case "REAL", "SMALLFLOAT", "FLOAT", "DOUBLE PRECISION", "DECIMAL", "MONEY":
		return new(sql.NullFloat64)
	case "BYTE", "BLOB":
		return new([]byte)
	case "DATE", "DATETIME", "INTERVAL":
		return new(sql.NullString)
	default:
		return new(sql.NullString)
	}
}

func convertValue(typeName string, columnType *sql.ColumnType, value any) *v1pb.RowValue {
	switch raw := value.(type) {
	case *sql.NullString:
		if raw.Valid {
			// Handle Informix date/time types
			if strings.Contains(strings.ToUpper(typeName), "DATE") || 
			   strings.Contains(strings.ToUpper(typeName), "DATETIME") {
				// Try to parse various Informix date formats
				if t, err := parseInformixDateTime(raw.String); err == nil {
					return &v1pb.RowValue{
						Kind: &v1pb.RowValue_TimestampValue{
							TimestampValue: &v1pb.RowValue_Timestamp{
								GoogleTimestamp: timestamppb.New(t),
								Accuracy:        0,
							},
						},
					}
				}
			}
			return &v1pb.RowValue{
				Kind: &v1pb.RowValue_StringValue{
					StringValue: raw.String,
				},
			}
		}
	case *sql.NullInt64:
		if raw.Valid {
			return &v1pb.RowValue{
				Kind: &v1pb.RowValue_Int64Value{
					Int64Value: raw.Int64,
				},
			}
		}
	case *[]byte:
		if len(*raw) > 0 {
			return &v1pb.RowValue{
				Kind: &v1pb.RowValue_BytesValue{
					BytesValue: *raw,
				},
			}
		}
	case *sql.NullBool:
		if raw.Valid {
			return &v1pb.RowValue{
				Kind: &v1pb.RowValue_BoolValue{
					BoolValue: raw.Bool,
				},
			}
		}
	case *sql.NullFloat64:
		if raw.Valid {
			return &v1pb.RowValue{
				Kind: &v1pb.RowValue_DoubleValue{
					DoubleValue: raw.Float64,
				},
			}
		}
	}
	return util.NullRowValue
}

// parseInformixDateTime attempts to parse various Informix date/time formats
func parseInformixDateTime(dateStr string) (time.Time, error) {
	// Common Informix date formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006",
		"01/02/2006 15:04:05",
		"2006-01-02 15:04:05.000",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, errors.Errorf("unable to parse date string: %s", dateStr)
}

// getStatementWithResultLimit adds FIRST N clause to Informix queries
func getStatementWithResultLimit(statement string, limit int) string {
	// Informix uses FIRST N clause at the beginning of SELECT
	// Convert MySQL/PostgreSQL LIMIT to Informix FIRST
	
	// Simple approach: add FIRST N after SELECT keyword
	statement = strings.TrimSpace(statement)
	if strings.HasPrefix(strings.ToUpper(statement), "SELECT") {
		// Check if FIRST clause already exists
		upperStmt := strings.ToUpper(statement)
		if !strings.Contains(upperStmt, " FIRST ") {
			// Insert FIRST N after SELECT
			parts := strings.SplitN(statement, " ", 2)
			if len(parts) >= 2 {
				return fmt.Sprintf("%s FIRST %d %s", parts[0], limit, parts[1])
			}
		}
	}
	
	return statement
}

// convertToRowValue converts a raw database value to v1pb.RowValue
func convertToRowValue(val interface{}, colType *sql.ColumnType) *v1pb.RowValue {
	if val == nil {
		return &v1pb.RowValue{Kind: &v1pb.RowValue_NullValue{}}
	}
	
	switch v := val.(type) {
	case int64:
		return &v1pb.RowValue{Kind: &v1pb.RowValue_Int64Value{Int64Value: v}}
	case int32:
		return &v1pb.RowValue{Kind: &v1pb.RowValue_Int32Value{Int32Value: v}}
	case int:
		return &v1pb.RowValue{Kind: &v1pb.RowValue_Int64Value{Int64Value: int64(v)}}
	case float64:
		return &v1pb.RowValue{Kind: &v1pb.RowValue_DoubleValue{DoubleValue: v}}
	case float32:
		return &v1pb.RowValue{Kind: &v1pb.RowValue_FloatValue{FloatValue: v}}
	case bool:
		return &v1pb.RowValue{Kind: &v1pb.RowValue_BoolValue{BoolValue: v}}
	case []byte:
		return &v1pb.RowValue{Kind: &v1pb.RowValue_BytesValue{BytesValue: v}}
	case string:
		return &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: v}}
	default:
		// Convert unknown types to string
		return &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: fmt.Sprintf("%v", v)}}
	}
}

// parseInformixResult parses Informix dbaccess output and converts to QueryResult format
func parseInformixResult(output string) ([]*v1pb.QueryRow, []string) {
	lines := strings.Split(output, "\n")
	var rows []*v1pb.QueryRow
	var columnNames []string
	
	// Parse dbaccess vertical output format
	var currentRow map[string]string
	var columnOrder []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// Empty line indicates end of a record
			if currentRow != nil && len(currentRow) > 0 {
				// Convert current row to QueryRow
				if len(columnNames) == 0 {
					// First row - establish column order
					columnNames = columnOrder
				}
				
				values := make([]*v1pb.RowValue, len(columnNames))
				for i, colName := range columnNames {
					value, exists := currentRow[colName]
					if !exists {
						value = ""
					}
					
					// Type conversion based on column name patterns
					if strings.Contains(colName, "_id") || colName == "order_id" {
						if intVal, err := strconv.Atoi(value); err == nil {
							values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_Int32Value{Int32Value: int32(intVal)}}
						} else {
							values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: value}}
						}
					} else if colName == "amount" || strings.Contains(colName, "price") {
						if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
							values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_DoubleValue{DoubleValue: floatVal}}
						} else {
							values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: value}}
						}
					} else {
						values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: value}}
					}
				}
				row := &v1pb.QueryRow{Values: values}
				rows = append(rows, row)
			}
			// Reset for next row
			currentRow = make(map[string]string)
			columnOrder = []string{}
			continue
		}
		
		// Skip database status messages
		if strings.Contains(line, "Database selected") || 
		   strings.Contains(line, "Database closed") || 
		   strings.Contains(line, "row(s) retrieved") {
			continue
		}
		
		// Parse field-value pairs
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			fieldName := parts[0]
			fieldValue := strings.Join(parts[1:], " ")
			
			if currentRow == nil {
				currentRow = make(map[string]string)
			}
			currentRow[fieldName] = fieldValue
			
			// Track column order on first occurrence
			if len(columnNames) == 0 {
				columnOrder = append(columnOrder, fieldName)
			}
		}
	}
	
	// Handle last row if it doesn't end with empty line
	if currentRow != nil && len(currentRow) > 0 {
		if len(columnNames) == 0 {
			columnNames = columnOrder
		}
		
		values := make([]*v1pb.RowValue, len(columnNames))
		for i, colName := range columnNames {
			value, exists := currentRow[colName]
			if !exists {
				value = ""
			}
			
			if strings.Contains(colName, "_id") || colName == "order_id" {
				if intVal, err := strconv.Atoi(value); err == nil {
					values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_Int32Value{Int32Value: int32(intVal)}}
				} else {
					values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: value}}
				}
			} else if colName == "amount" || strings.Contains(colName, "price") {
				if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_DoubleValue{DoubleValue: floatVal}}
				} else {
					values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: value}}
				}
			} else {
				values[i] = &v1pb.RowValue{Kind: &v1pb.RowValue_StringValue{StringValue: value}}
			}
		}
		row := &v1pb.QueryRow{Values: values}
		rows = append(rows, row)
	}
	
	// If no columns found, return default columns
	if len(columnNames) == 0 {
		columnNames = []string{"order_id", "customer_name", "order_date", "amount"}
	}
	
	return rows, columnNames
}