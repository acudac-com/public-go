package cx

import (
	"context"
	"database/sql"
	"time"
)

type Cx struct {
	context.Context
	now *time.Time
	tx  *SqlTx
}

// Returns Cx from context or creates a new one if it doesn't exist.
func New(ctx context.Context) *Cx {
	if cc, ok := ctx.(*Cx); ok {
		return cc
	}
	return &Cx{Context: ctx}
}

// Returns the current time either from the cached value or by calling
// time.Now().
func (cx *Cx) Now() time.Time {
	if cx.now != nil {
		return *cx.now
	}

	now := time.Now().UTC()
	cx.now = &now
	return now
}

// Returns the current time in UTC, using the context to create a Cx if
// necessary.
func Now(ctx context.Context) time.Time {
	cx := New(ctx)
	return cx.Now()
}

// Returns a SqlTx for the given database connection. Begins a new transaction
// if one does not already exist in cx.
func (cx *Cx) Tx(db *sql.DB) *SqlTx {
	if cx.tx != nil {
		return &SqlTx{Tx: cx.tx.Tx, cx: cx}
	}

	tx, err := db.BeginTx(cx.Context, nil)
	if err != nil {
		panic(err)
	}
	return &SqlTx{Tx: tx, cx: cx, created: true}
}

// Creates a new SqlTx for the given database connection. Begins a new
// transaction if one does not already exist in ctx.(*Cx).
func Tx(ctx context.Context, db *sql.DB) *SqlTx {
	cx := New(ctx)
	return cx.Tx(db)
}

// Returns the SqlTx if it exists in the context, or nil if it does not.
func (cx *Cx) TxIfExists() *SqlTx {
	return cx.tx
}

// Returns a SqlTx if it exists in the context, or nil if it does not.
func TxIfExists(ctx context.Context) *SqlTx {
	cx := New(ctx)
	return cx.TxIfExists()
}

type SqlTx struct {
	*sql.Tx
	created bool
	cx      *Cx
}

// Rolls back the transaction if it was created and not inherited.
func (s *SqlTx) Rollback() error {
	if s.created {
		s.cx.tx = nil
		return s.Tx.Rollback()
	}
	return nil
}

// Commits the transaction if it was created and not inherited.
func (s *SqlTx) Commit() error {
	if s.created {
		s.cx.tx = nil
		return s.Tx.Commit()
	}
	return nil
}
