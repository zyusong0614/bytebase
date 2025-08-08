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

	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/bytebase/bytebase/backend/common"
	"github.com/bytebase/bytebase/backend/common/log"
	storepb "github.com/bytebase/bytebase/backend/generated-go/store"
	v1pb "github.com/bytebase/bytebase/backend/generated-go/v1"
	"github.com/bytebase/bytebase/backend/plugin/db"
	"github.com/bytebase/bytebase/backend/plugin/db/util"
	"github.com/bytebase/bytebase/backend/plugin/parser/base"
)

var (
	_ db.Driver = (*Driver)(nil)
)

func init() {
	db.Register(storepb.Engine_INFORMIX, newDriver)
}

// Driver is the Informix driver.
type Driver struct {
	db               *sql.DB
	databaseName     string
	connectionString string
	connectionCtx    db.ConnectionContext
	
	// ODBC-specific fields (only available with informix build tag)
	odbcDriver   *ODBCDriver
}

func newDriver() db.Driver {
	return &Driver{}
}

// Open opens an Informix driver.
func (d *Driver) Open(ctx context.Context, dbType storepb.Engine, config db.ConnectionConfig) (db.Driver, error) {
	// Delegate to ODBC driver if available (with build tag)
	if odbcDriver, err := d.openODBC(ctx, dbType, config); err == nil {
		d.odbcDriver = odbcDriver
		return d, nil
	}
	
	// For now, return error since we don't have native Go driver for Informix
	// In the future, this could use a native Informix driver
	return nil, errors.New("Informix native driver not implemented, requires ODBC with informix build tag")
}

// Close closes the driver.
func (d *Driver) Close(ctx context.Context) error {
	var err error
	if d.db != nil {
		err = multierr.Append(err, d.db.Close())
	}
	if d.odbcDriver != nil {
		err = multierr.Append(err, d.odbcDriver.Close(ctx))
	}
	return err
}

// Ping pings the database.
func (d *Driver) Ping(ctx context.Context) error {
	if d.db != nil {
		return d.db.PingContext(ctx)
	}
	if d.odbcDriver != nil {
		return d.odbcDriver.Ping(ctx)
	}
	return errors.New("not connected to database")
}

// GetDB gets the database.
func (d *Driver) GetDB() *sql.DB {
	if d.db != nil {
		return d.db
	}
	if d.odbcDriver != nil {
		return d.odbcDriver.GetDB()
	}
	return nil
}

// Execute executes a SQL statement.
func (d *Driver) Execute(ctx context.Context, statement string, opts db.ExecuteOptions) (int64, error) {
	// For ODBC driver, delegate to the ODBC-specific implementation
	if d.odbcDriver != nil {
		return d.odbcDriver.Execute(ctx, statement, opts)
	}
	
	// For native implementation, use standard sql.DB
	if d.db == nil {
		return 0, errors.New("not connected to database")
	}
	
	// Parse transaction mode from the script
	transactionMode, cleanedStatement := base.ParseTransactionMode(statement)
	statement = cleanedStatement
	
	// Apply default when transaction mode is not specified
	if transactionMode == common.TransactionModeUnspecified {
		transactionMode = common.GetDefaultTransactionMode()
	}
	
	// Split SQL statements for Informix
	singleSQLs, err := base.SplitMultiSQL(storepb.Engine_INFORMIX, statement)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to split sql")
	}
	singleSQLs = base.FilterEmptySQL(singleSQLs)
	if len(singleSQLs) == 0 {
		return 0, nil
	}
	
	conn, err := d.db.Conn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	
	// Execute based on transaction mode
	if transactionMode == common.TransactionModeOff {
		return d.executeInAutoCommitMode(ctx, conn, singleSQLs, opts)
	}
	return d.executeInTransactionMode(ctx, conn, singleSQLs, opts)
}

// executeInTransactionMode executes statements within a single transaction
func (d *Driver) executeInTransactionMode(ctx context.Context, conn *sql.Conn, commands []base.SingleSQL, opts db.ExecuteOptions) (int64, error) {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		opts.LogTransactionControl(storepb.TaskRunLog_TransactionControl_BEGIN, err.Error())
		return 0, err
	}
	opts.LogTransactionControl(storepb.TaskRunLog_TransactionControl_BEGIN, "")
	
	committed := false
	defer func() {
		if !committed {
			err := tx.Rollback()
			var rerr string
			if err != nil {
				rerr = err.Error()
			}
			opts.LogTransactionControl(storepb.TaskRunLog_TransactionControl_ROLLBACK, rerr)
		}
	}()
	
	var totalRowsAffected int64
	for i, command := range commands {
		indexes := []int32{int32(i)}
		opts.LogCommandExecute(indexes)
		
		sqlResult, err := tx.ExecContext(ctx, command.Text)
		if err != nil {
			opts.LogCommandResponse(indexes, 0, nil, err.Error())
			return 0, &db.ErrorWithPosition{
				Err:   errors.Wrapf(err, "failed to execute context in a transaction"),
				Start: command.Start,
				End:   command.End,
			}
		}
		
		rowsAffected, err := sqlResult.RowsAffected()
		if err != nil {
			slog.Info("rowsAffected returns error", log.BBError(err))
		}
		totalRowsAffected += rowsAffected
		
		opts.LogCommandResponse(indexes, int32(rowsAffected), []int32{int32(rowsAffected)}, "")
	}
	
	if err := tx.Commit(); err != nil {
		opts.LogTransactionControl(storepb.TaskRunLog_TransactionControl_COMMIT, err.Error())
		return 0, errors.Wrapf(err, "failed to commit execute transaction")
	}
	opts.LogTransactionControl(storepb.TaskRunLog_TransactionControl_COMMIT, "")
	committed = true
	
	return totalRowsAffected, nil
}

// executeInAutoCommitMode executes statements sequentially in auto-commit mode
func (d *Driver) executeInAutoCommitMode(ctx context.Context, conn *sql.Conn, commands []base.SingleSQL, opts db.ExecuteOptions) (int64, error) {
	var totalRowsAffected int64
	
	for i, command := range commands {
		indexes := []int32{int32(i)}
		opts.LogCommandExecute(indexes)
		
		sqlResult, err := conn.ExecContext(ctx, command.Text)
		if err != nil {
			opts.LogCommandResponse(indexes, 0, nil, err.Error())
			return 0, &db.ErrorWithPosition{
				Err:   errors.Wrapf(err, "failed to execute statement %d in auto-commit mode", i+1),
				Start: command.Start,
				End:   command.End,
			}
		}
		
		rowsAffected, err := sqlResult.RowsAffected()
		if err != nil {
			slog.Info("rowsAffected returns error", log.BBError(err))
		}
		totalRowsAffected += rowsAffected
		
		opts.LogCommandResponse(indexes, int32(rowsAffected), []int32{int32(rowsAffected)}, "")
	}
	
	return totalRowsAffected, nil
}

// QueryConn queries a SQL statement in a given connection.
func (d *Driver) QueryConn(ctx context.Context, conn *sql.Conn, statement string, queryContext db.QueryContext) ([]*v1pb.QueryResult, error) {
	// For ODBC driver, delegate to the ODBC-specific implementation
	if d.odbcDriver != nil {
		return d.odbcDriver.QueryConn(ctx, conn, statement, queryContext)
	}
	
	// For native implementation, use standard sql.DB
	if d.db == nil {
		result := &v1pb.QueryResult{
			Statement: statement,
			Error:     "not connected to database",
		}
		return []*v1pb.QueryResult{result}, nil
	}
	
	singleSQLs, err := base.SplitMultiSQL(storepb.Engine_INFORMIX, statement)
	if err != nil {
		return nil, err
	}
	singleSQLs = base.FilterEmptySQL(singleSQLs)
	if len(singleSQLs) == 0 {
		return nil, nil
	}
	
	var results []*v1pb.QueryResult
	for _, singleSQL := range singleSQLs {
		statement := singleSQL.Text
		if queryContext.Explain {
			statement = fmt.Sprintf("SET EXPLAIN ON; %s", statement)
		} else if queryContext.Limit > 0 {
			statement = getStatementWithResultLimit(statement, queryContext.Limit)
		}
		
		_, allQuery, err := base.ValidateSQLForEditor(storepb.Engine_INFORMIX, statement)
		if err != nil {
			slog.Error("failed to validate sql", slog.String("statement", statement), log.BBError(err))
			allQuery = true
		}
		
		startTime := time.Now()
		queryResult, err := func() (*v1pb.QueryResult, error) {
			if allQuery {
				rows, err := conn.QueryContext(ctx, statement)
				if err != nil {
					return nil, err
				}
				defer rows.Close()
				r, err := util.RowsToQueryResult(rows, makeValueByTypeName, convertValue, queryContext.MaximumSQLResultSize)
				if err != nil {
					return nil, err
				}
				if err := rows.Err(); err != nil {
					return nil, err
				}
				return r, nil
			}
			
			sqlResult, err := conn.ExecContext(ctx, statement)
			if err != nil {
				return nil, err
			}
			affectedRows, err := sqlResult.RowsAffected()
			if err != nil {
				slog.Info("rowsAffected returns error", log.BBError(err))
			}
			return util.BuildAffectedRowsResult(affectedRows, nil), nil
		}()
		
		stop := false
		if err != nil {
			queryResult = &v1pb.QueryResult{
				Error: err.Error(),
			}
			stop = true
		}
		queryResult.Statement = statement
		queryResult.Latency = durationpb.New(time.Since(startTime))
		if queryResult.Rows != nil {
			queryResult.RowsCount = int64(len(queryResult.Rows))
		}
		results = append(results, queryResult)
		if stop {
			break
		}
	}
	
	return results, nil
}

// getVersion gets the version.
func (d *Driver) getVersion(ctx context.Context) (string, string, error) {
	query := "SELECT FIRST 1 version FROM sysmaster:sysvinfo WHERE name = 'DBSERVERTYPE'"
	var version string
	if err := d.db.QueryRowContext(ctx, query).Scan(&version); err != nil {
		if err == sql.ErrNoRows {
			return "", "", common.FormatDBErrorEmptyRowWithQuery(query)
		}
		return "", "", util.FormatErrorWithQuery(err, query)
	}
	
	return parseInformixVersion(version)
}

func parseInformixVersion(version string) (string, string, error) {
	// Informix version parsing logic
	parts := strings.Fields(version)
	if len(parts) > 0 {
		return parts[0], strings.Join(parts[1:], " "), nil
	}
	return version, "", nil
}

// generateMockQueryResult generates mock query results when ODBC is not available
func (d *Driver) generateMockQueryResult(statement string) []*v1pb.QueryResult {
	result := &v1pb.QueryResult{
		Statement: statement,
		ColumnNames: []string{"status"},
		Rows: []*v1pb.QueryRow{
			{
				Values: []*v1pb.RowValue{
					{Kind: &v1pb.RowValue_StringValue{StringValue: "ODBC driver not available - using mock results"}},
				},
			},
		},
		Error: "ODBC connection unavailable",
	}
	return []*v1pb.QueryResult{result}
}

