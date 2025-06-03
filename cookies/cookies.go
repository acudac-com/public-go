// Package cookies provides utilities for encoding and decoding Go values into
// HTTP cookies using gob serialization and base64 encoding.
//
// This allows type-safe storage of complex state in cookies.
package cookies

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"net/http"
)

// Read the gob encoded cookie (if any) for the specified path.
func Read[T any](path string, r *http.Request, s T) T {
	cookie, err := r.Cookie(path)
	if err == nil {
		decoded, err := base64.StdEncoding.DecodeString(cookie.Value)
		if err == nil {
			dec := gob.NewDecoder(bytes.NewReader(decoded))
			_ = dec.Decode(s)
		}
	}
	return s
}

// Write the gob encoded cookie for the specified path.
func Write(path string, w http.ResponseWriter, state any) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(state); err != nil {
		println(err.Error())
		return
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	http.SetCookie(w, &http.Cookie{
		Name:  path,
		Value: encoded,
		Path:  path,
	})
}
