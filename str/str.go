// Package str provides utilities related to strings.
package str

import (
	"context"
	"crypto/rand"
	"math/big"

	"go.alis.build/alog"
)

// Generates a secure code consisting of capital letters and numbers, excluding
// 'I', 'O' and '0' for readability. The code consists of one or more segments
// seperated by hyphens.
func SecureCode(segmentL int, segments int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ123456789"
	L := big.NewInt(int64(len(charset)))
	code := make([]byte, 0, segmentL*segments+(segments-1))
	for s := range segments {
		if s > 0 {
			code = append(code, '-')
		}
		for range segmentL {
			num, err := rand.Int(rand.Reader, L)
			if err != nil {
				alog.Fatalf(context.Background(), "generating random integer: %v", err)
			}
			code = append(code, charset[num.Int64()])
		}
	}
	return string(code)
}
