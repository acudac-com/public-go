package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/acudac-com/public-go/cx"
	spannerdriver "github.com/googleapis/go-sql-spanner"
	"go.alis.build/alog"
)

var (
	Db *sql.DB
)

func UseSpanner(cfg spannerdriver.ConnectorConfig) {
	ctx := context.Background()
	var err error

	// Create a Connector for Spanner to create a DB with a custom configuration.
	c, err := spannerdriver.CreateConnector(cfg)
	if err != nil {
		alog.Fatalf(ctx, "creating spanner connector: %v", err)
	}

	// Create a DB using the Connector.
	sqlDb := sql.OpenDB(c)
	if err := sqlDb.Ping(); err != nil {
		alog.Fatalf(ctx, "pinging spanner database: %v", err)
	}
	Db = sqlDb
}

// Will not begin a new tx if ctx already has an existing ctx. Returns a handler
// that should be used to defer rollback and commit if the transaction was
// started here.
func BeginTx(ctx context.Context, opts *sql.TxOptions) (*cx.Cx, error) {
	cx, _, err := cx.Tx(ctx, Db)
	return cx, err
}

// Executes the provided query. If the ctx contains a transaction, it will be
// used instead the default Db.ExecContext.
func Exec(ctx context.Context, query string, args ...any) (int64, error) {
	var err error
	var result sql.Result
	tx := cx.TxIfExists(ctx)
	if tx == nil {
		result, err = Db.ExecContext(ctx, query, args...)
	} else {
		result, err = tx.ExecContext(ctx, query, args...)
	}
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, e("getting rows affected: %w", err)
	}
	return rowsAffected, nil
}

// Executes the provided query. If the ctx contains a transaction, it will be
// used instead the default Db.ExecContext.
func Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	tx := cx.TxIfExists(ctx)
	if tx == nil {
		return Db.QueryContext(ctx, query, args...)
	} else {
		return tx.QueryContext(ctx, query, args...)
	}
}

// Returns a string with a number of placeholders (?) for SQL queries.
func Placeholders(nr int) string {
	placeholders := make([]string, 0, nr)
	for range nr {
		placeholders = append(placeholders, "?")
	}
	return f("%s", strings.Join(placeholders, ", "))
}

var (
	f = fmt.Sprintf
	e = fmt.Errorf
)
