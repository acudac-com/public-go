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

func New(ctx context.Context) *Cx {
	if cc, ok := ctx.(*Cx); ok {
		return cc
	}
	return &Cx{Context: ctx}
}

func (cx *Cx) Now() time.Time {
	if cx.now != nil {
		return *cx.now
	}

	now := time.Now().UTC()
	cx.now = &now
	return now
}

func Now(ctx context.Context) time.Time {
	cx := New(ctx)
	return cx.Now()
}

type SqlTx struct {
	*sql.Tx
	created bool
	cx      *Cx
}

func (cx *Cx) Tx(db *sql.DB) *SqlTx {
	if cx.tx != nil {
		return &SqlTx{Tx: cx.tx.Tx, cx: cx}
	}

	tx, err := db.BeginTx(cx.Context, nil)
	if err != nil {
		panic(err)
	}
	return &SqlTx{Tx: tx, cx: cx, created: true}, nil
}

func Tx(ctx context.Context, db *sql.DB) (*SqlTx, error) {
	cx := New(ctx)
	tx, err := cx.Tx(db)
	return cx, tx, err
}

func (cx *Cx) TxIfExists() *SqlTx {
	return cx.tx
}

func TxIfExists(ctx context.Context) *SqlTx {
	cx := New(ctx)
	return cx.tx
}

func (s *SqlTx) Rollback() error {
	if s.created {
		s.cx.tx = nil
		return s.Tx.Rollback()
	}
	return nil
}

func (s *SqlTx) Commit() error {
	if s.created {
		s.cx.tx = nil
		return s.Tx.Commit()
	}
	return nil
}
