package sqlx_test

import (
	"testing"

	"github.com/acudac-com/public-go/sqlx"
)

func TestExec(t *testing.T) {
	if _, err := sqlx.Exec(t.Context(), "create schema gm"); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
}
