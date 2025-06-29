/*
Package jsonx provides utility functions for encoding and decoding JSON data,
with additional support for base64 URL encoding and decoding. It wraps the
standard library's encoding/json package to simplify marshaling and unmarshaling
operations, and integrates with a custom base64 utility for encoding JSON data
to and from base64 URL-safe strings.
*/
package jsonx

import (
	"encoding/json"
	"fmt"

	"github.com/acudac-com/public-go/encoding/b64"
)

// Json encodes the payload.
func Marshal(payload any) []byte {
	marshalled, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Errorf("jsonx.Marshal(): %v", err))
	}
	return marshalled
}

// Json encodes the payload and base64 url encodes the result.
func MarshalB64[T any](payload T) []byte {
	marshalled := Marshal(payload)
	encoded := b64.UrlEncode(marshalled)
	return encoded
}

// Json decodes the data into the provided payload.
func Unmarshal[T any](data []byte, payload T) (T, error) {
	err := json.Unmarshal(data, payload)
	return payload, err
}

// Base64 url decodes the data and json decodes the result into the provided
// payload.
func UnmarshalB64[T any](data []byte, payload T) (T, error) {
	dec, err := b64.UrlDecode(data)
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(dec, payload)
	return payload, err
}
