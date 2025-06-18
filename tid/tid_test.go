package tid_test

import (
	"testing"
	"time"

	"github.com/acudac-com/public-go/cx"
	"github.com/acudac-com/public-go/tid"
)

func TestSparse(t *testing.T) {
	cx := cx.New(t.Context())
	for range 2000 {
		println(tid.Sparse(cx))
		time.Sleep(50 * time.Nanosecond)
	}
}

func BenchmarkSparse(b *testing.B) {
	cx := cx.New(b.Context())
	for b.Loop() {
		tid.Sparse(cx)
	}
}

func TestLatestFirst(t *testing.T) {
	cx := cx.New(t.Context())
	for range 2000 {
		println(tid.LatestFirst(cx))
		time.Sleep(50 * time.Nanosecond)
	}
}

func BenchmarkLatestFirst(b *testing.B) {
	cx := cx.New(b.Context())
	for b.Loop() {
		tid.LatestFirst(cx)
	}
}
