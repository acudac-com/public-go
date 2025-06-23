package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cloud.google.com/go/spanner"
	"github.com/acudac-com/public-go/cx"
	"github.com/acudac-com/public-go/env"
	spannerdriver "github.com/googleapis/go-sql-spanner"
	"go.alis.build/alog"
	"google.golang.org/api/option"
)

var Db *sql.DB

func init() {
	if env.OptionalString("SPANNER_PROJECT", "") != "" {
		UseSpanner()
	}
}

func UseSpanner() {
	ctx := context.Background()
	var err error
	cfg := spannerdriver.ConnectorConfig{
		Project:            env.RequiredString("SPANNER_PROJECT"),
		Instance:           env.RequiredString("SPANNER_INSTANCE"),
		Database:           env.OptionalString("SPANNER_DB", env.Env),
		AutoConfigEmulator: env.OptionalBool("SPANNER_EMULATOR", env.IsLocal()),
		Configurator: func(config *spanner.ClientConfig, opts *[]option.ClientOption) {
			if !env.IsLocal() {
				config.DatabaseRole = env.OptionalString("SPANNER_ROLE", env.Product)
			}
		},
	}

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

// Will not begin a new tx if ctx already has an existing tx. Remember to defer
// tx.Rollback() and err := tx.Commit(); err != nil{...} at the end of your
// function.
func BeginTx(ctx context.Context, opts *sql.TxOptions) *cx.SqlTx {
	return cx.Tx(ctx, Db)
}

// Executes the provided query. If the ctx contains a transaction, it will be
// used instead the default Db.ExecContext.
func Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx := cx.TxIfExists(ctx)
	if tx == nil {
		return Db.ExecContext(ctx, query, args...)
	} else {
		return tx.ExecContext(ctx, query, args...)
	}
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
