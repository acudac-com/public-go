package tid_test

import (
	"testing"
	"time"

	"github.com/acudac-com/public-go/tid"
)

func Test_UnixSparse(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	randomNr := 0
	id := tid.UnixSparse(now, &randomNr)
	expected := "0o242d"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}
}

func Benchmark_UnixSparse(b *testing.B) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for b.Loop() {
		tid.UnixSparse(now, nil)
	}
}

func Test_MilliSparse(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	randomNr := 0
	id := tid.MilliSparse(now, &randomNr)
	expected := "0o26pq2a"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}
}

func Benchmark_MilliSparse(b *testing.B) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for b.Loop() {
		tid.MilliSparse(now, nil)
	}
}

func Test_MicroSparse(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	randomNr := 0
	id := tid.MicroSparse(now, &randomNr)
	expected := "0o2q4n5wr7"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}
}

func Benchmark_MicroSparse(b *testing.B) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for b.Loop() {
		tid.MicroSparse(now, nil)
	}
}

func Test_NanoSparse(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	randomNr := 0
	id := tid.NanoSparse(now, &randomNr)
	expected := "0o2a8jq8tyz5"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}
}

func Benchmark_NanoSparse(b *testing.B) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for b.Loop() {
		tid.NanoSparse(now, nil)
	}
}

func Test_UnixLatestFirst(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	id := tid.UnixLatestFirst(now)
	expected := "1954175r"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}

	now = time.Date(5025, 1, 1, 0, 0, 0, 0, time.UTC)
	id = tid.UnixLatestFirst(now)
	expected = "01nfh4hr"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}
}

func Benchmark_UnixLatestFirst(b *testing.B) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for b.Loop() {
		tid.UnixLatestFirst(now)
	}
}

func Test_MilliLatestFirst(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	id := tid.MilliLatestFirst(now)
	expected := "yty01avq0"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}

	now = time.Date(5025, 1, 1, 0, 0, 0, 0, time.UTC)
	id = tid.MilliLatestFirst(now)
	expected = "19utvot20"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}
}

func Benchmark_MilliLatestFirst(b *testing.B) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for b.Loop() {
		tid.MilliLatestFirst(now)
	}
}

func Test_MicroLatestFirst(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	id := tid.MicroLatestFirst(now)
	expected := "qvjsh069680"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}

	now = time.Date(5025, 1, 1, 0, 0, 0, 0, time.UTC)
	id = tid.MicroLatestFirst(now)
	expected = "0zdse0933k0"
	if *id != expected {
		t.Errorf("expected %s, got %s", expected, *id)
	}
}

func Benchmark_MicroLatestFirst(b *testing.B) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for b.Loop() {
		tid.MicroLatestFirst(now)
	}
}
