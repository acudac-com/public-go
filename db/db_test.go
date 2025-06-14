package db_test

import (
	"testing"

	"github.com/acudac-com/public-go/db"
)

func TestExec(t *testing.T) {
	if _, err := db.Exec(t.Context(), "create schema gm"); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}
}
