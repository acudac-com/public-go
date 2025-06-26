package sqlx

import (
	"context"
	"database/sql"
	"fmt"

	spannerdriver "github.com/googleapis/go-sql-spanner"
	"go.alis.build/alog"
)

type Db struct {
	db *sql.DB
}

// Returns new spanner db instance using the provided connector config.
func NewSpannerDb(cfg spannerdriver.ConnectorConfig) *Db {
	// create spanner connector
	c, err := spannerdriver.CreateConnector(cfg)
	if err != nil {
		alog.Fatalf(context.Background(), "creating spanner connector: %v", err)
	}

	// create a DB using the Connector and ping it to test connectivity
	db := sql.OpenDB(c)
	if err := db.Ping(); err != nil {
		alog.Fatalf(context.Background(), "pinging spanner database: %v", err)
	}
	return &Db{db}
}

type CtxKey string

const TxCtxKey CtxKey = "sqlx.Tx"

type SqlTx struct {
	*sql.Tx
	created bool
}

// Begins a new transaction and adds it to context.
func (d *Db) BeginTx(ctx context.Context, opts *sql.TxOptions) (context.Context, *SqlTx) {
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		panic(fmt.Errorf("sqlx.NewTx(): %v", err))
	}
	sqlTx := &SqlTx{tx, true}
	return context.WithValue(ctx, TxCtxKey, sqlTx), sqlTx
}

// Returns any existing transaction in ctx if it exists or begins a new one
// otherwise. If existing tx is returned, its rollback and commit functions are
// ignored.
func (d *Db) ResumeOrBeginTx(ctx context.Context, opts *sql.TxOptions) (context.Context, *SqlTx) {
	if tx, ok := ctx.Value(TxCtxKey).(*SqlTx); ok && tx.Tx != nil {
		return ctx, &SqlTx{tx.Tx, false}
	}
	return d.BeginTx(ctx, opts)
}

// Rolls back the transaction if it was created and not inherited.
func (s *SqlTx) Rollback(ctx context.Context) error {
	if s.created && s.Tx != nil {
		return s.Tx.Rollback()
	}
	return nil
}

// Commits the transaction if it was created and not inherited.
func (s *SqlTx) Commit(ctx context.Context) error {
	if s.created && s.Tx != nil {
		if err := s.Tx.Commit(); err != nil {
			return fmt.Errorf("sqlx.SqlTx.Commit(): %v", err)
		}
		s.Tx = nil
		return nil
	}
	return nil
}

// Executes the provided query. If the ctx contains a transaction, it will be
// used instead the default Db.ExecContext.
func (d *Db) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	tx, ok := ctx.Value(TxCtxKey).(*SqlTx)
	if ok && tx.Tx != nil {
		return tx.ExecContext(ctx, query, args...)
	} else {
		return d.db.ExecContext(ctx, query, args...)
	}
}

// Executes the provided query. If the ctx contains a transaction, it will be
// used instead the default Db.ExecContext.
func (d *Db) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	tx, ok := ctx.Value(TxCtxKey).(*SqlTx)
	if ok && tx.Tx != nil {
		return tx.QueryContext(ctx, query, args...)
	} else {
		return d.db.QueryContext(ctx, query, args...)
	}
}
