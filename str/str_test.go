package str_test

import (
	"testing"

	"github.com/acudac-com/public-go/str"
)

func TestSecureCode(t *testing.T) {
	println(str.SecureCode(4, 2))
}

func BenchmarkSecureCode(b *testing.B) {
	for b.Loop() {
		_ = str.SecureCode(4, 2)
	}
}
