package timex_test

import (
	"testing"
	"time"

	"github.com/acudac-com/public-go/timex"
)

func BenchmarkNow(b *testing.B) {
	ctx, _ := timex.Now(b.Context())
	for b.Loop() {
		timex.Now(ctx)
	}
}

func TestSparse(t *testing.T) {
	ctx, _ := timex.Now(t.Context())
	for range 2000 {
		println(timex.SparseId(ctx))
		time.Sleep(50 * time.Nanosecond)
	}
}

func BenchmarkSparse(b *testing.B) {
	ctx, _ := timex.Now(b.Context())
	for b.Loop() {
		timex.SparseId(ctx)
	}
}

func TestLatestFirst(t *testing.T) {
	for range 2000 {
		println(timex.LatestFirstId(t.Context()))
		time.Sleep(100 * time.Microsecond)
	}
}

func BenchmarkLatestFirst(b *testing.B) {
	ctx, _ := timex.Now(b.Context())
	for b.Loop() {
		timex.LatestFirstId(ctx)
	}
}
