package cx

import (
	"context"
	"database/sql"
	"time"
)

type Cx struct {
	context.Context
	now *time.Time
	tx  *sqlTx
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

func Now(ctx context.Context) (*Cx, time.Time) {
	cx := New(ctx)
	return cx, cx.Now()
}

type sqlTx struct {
	*sql.Tx
	created bool
	cx      *Cx
}

func (cx *Cx) Tx(db *sql.DB) (*sqlTx, error) {
	if cx.tx != nil {
		return &sqlTx{Tx: cx.tx.Tx, cx: cx}, nil
	}

	tx, err := db.BeginTx(cx.Context, nil)
	if err != nil {
		return nil, err
	}
	return &sqlTx{Tx: tx, cx: cx, created: true}, nil
}

func Tx(ctx context.Context, db *sql.DB) (*Cx, *sqlTx, error) {
	cx := New(ctx)
	tx, err := cx.Tx(db)
	return cx, tx, err
}

func (s *sqlTx) Rollback() error {
	if s.created {
		s.cx.tx = nil
		return s.Tx.Rollback()
	}
	return nil
}

func (s *sqlTx) Commit() error {
	if s.created {
		s.cx.tx = nil
		return s.Tx.Commit()
	}
	return nil
}
