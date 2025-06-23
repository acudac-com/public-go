package cx_test

import (
	"database/sql"
	"testing"

	"github.com/acudac-com/public-go/cx"
)

func BenchmarkNow(t *testing.B) {
	cx := cx.New(t.Context())
	for i := 0; i < t.N; i++ {
		_ = cx.Now()
	}
}

func TextTx(t *testing.B) {
	cx := cx.New(t.Context())
	db := &sql.DB{} // Mock or actual database connection
	tx := cx.Tx(db)
	defer tx.Rollback()
	if tx == nil {
		t.Fatal("Expected non-nil transaction")
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
}
