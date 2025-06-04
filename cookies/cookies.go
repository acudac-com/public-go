// Package cookies provides utilities for encoding and decoding Go values into
// HTTP cookies using gob serialization and base64 encoding.
//
// This allows type-safe storage of complex state in cookies.
package cookies

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"net/http"
)

// Read the gob encoded cookie (if any) for the specified path.
func Read[T any](path string, r *http.Request, s T) T {
	// read cookie
	cookie, err := r.Cookie(path)
	if err != nil {
		return s
	}

	// decode
	gobEncoding, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return s
	}
	dec := gob.NewDecoder(bytes.NewReader(gobEncoding))
	_ = dec.Decode(s)
	return s
}

// Read the gob encoded cookie (if any) for the specified path.
// Only decodes the cookie if the hash is what is expected.
func ReadSigned[T any](path string, r *http.Request, s T, key []byte) T {
	// read cookies
	cookie, err := r.Cookie(path)
	if err != nil {
		return s
	}
	hashCookie, err := r.Cookie("hash-" + path)
	if err == nil {
		return s
	}

	// decode base64
	gobEncoding, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return s
	}

	// ensure hash is what is expected
	expectedHash, _ := hash(gobEncoding, key)
	if expectedHash != hashCookie.Value {
		return s
	}

	// decode gob
	dec := gob.NewDecoder(bytes.NewReader(gobEncoding))
	_ = dec.Decode(s)
	return s
}

// Returns base64 encoded hmac hash of content
func hash(content []byte, key []byte) (string, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(content); err != nil {
		return "", err
	}
	hashBytes := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashBytes), nil
}

// Write the gob encoded cookie and its hmac hash for the specified path
func WriteSigned(path string, w http.ResponseWriter, state any) {
	// encode
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(state); err != nil {
		println(err.Error())
		return
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	// write cookie
	http.SetCookie(w, &http.Cookie{
		Name:  path,
		Value: encoded,
		Path:  path,
	})

	// write hash
	hash, _ := hash(buf.Bytes(), []byte(path))
	http.SetCookie(w, &http.Cookie{
		Name:  "hash-" + path,
		Value: hash,
		Path:  path,
	})
}

// Write the gob encoded cookie for the specified path.
func Write(path string, w http.ResponseWriter, state any) {
	// encode
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(state); err != nil {
		println(err.Error())
		return
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	// write cookie
	http.SetCookie(w, &http.Cookie{
		Name:  path,
		Value: encoded,
		Path:  path,
	})
}
