/*
Package gobx provides utility functions for encoding and decoding gob data,
with additional support for base64 URL encoding and decoding. It wraps the
standard library's encoding/gob package to simplify marshaling and unmarshaling
operations, and integrates with a custom base64 utility for encoding gob data
to and from base64 URL-safe strings.
*/
package gobx

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"

	"github.com/acudac-com/public-go/encoding/b64"
)

// Gob encodes the payload.
func Marshal(ctx context.Context, payload any) []byte {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(payload); err != nil {
		panic(fmt.Errorf("gobx.Marshal(): %v", err))
	}
	return buf.Bytes()
}

// Gob encodes the payload and base64 url encodes the result.
func MarshalB64(ctx context.Context, payload any) []byte {
	marshalled := Marshal(ctx, payload)
	return b64.UrlEncode(marshalled)
}

// Gob decodes the data into the provided payload.
func Unmarshal[T any](ctx context.Context, data []byte, payload T) (T, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&payload)
	return payload, err
}

// Base64 url decodes the data and gob decodes the result into the provided
// payload.
func UnmarshalB64[T any](ctx context.Context, data []byte, payload T) (T, error) {
	dec, err := b64.UrlDecode(data)
	if err != nil {
		return payload, err
	}
	return Unmarshal(ctx, dec, payload)
}
