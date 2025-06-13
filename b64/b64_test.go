package b64_test

import (
	"testing"

	"github.com/acudac-com/public-go/b64"
)

var data = []byte("Hello, World!")

func TestUrlEncoding(t *testing.T) {
	encoded := b64.UrlEncode(data)
	decoded, err := b64.UrlDecode(encoded)
	if err != nil {
		t.Fatalf("UrlDecode failed: %v", err)
	}
	if string(decoded) != string(data) {
		t.Fatalf("Expected %s, got %s", data, decoded)
	}
}

func BenchmarkUrlEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		encoded := b64.UrlEncode(data)
		if len(encoded) == 0 {
			b.Fatal("UrlEncode returned empty result")
		}
	}
}

func BenchmarkUrlDecode(b *testing.B) {
	encoded := b64.UrlEncode(data)
	for i := 0; i < b.N; i++ {
		decoded, err := b64.UrlDecode(encoded)
		if err != nil {
			b.Fatal(err)
		}
		if string(decoded) != string(data) {
			b.Fatalf("Expected %s, got %s", data, decoded)
		}
	}
}

func BenchmarkStdEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		encoded := b64.StdEncode(data)
		if len(encoded) == 0 {
			b.Fatal("StdEncode returned empty result")
		}
	}
}

func BenchmarkStdDecode(b *testing.B) {
	encoded := b64.StdEncode(data)
	for i := 0; i < b.N; i++ {
		decoded, err := b64.StdDecode(encoded)
		if err != nil {
			b.Fatal(err)
		}
		if string(decoded) != string(data) {
			b.Fatalf("Expected %s, got %s", data, decoded)
		}
	}
}
