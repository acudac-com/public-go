// Package cookies provides utilities for encoding and decoding Go values into
// HTTP cookies using base64, json, SHA512 hashing and AES-GCM encryption. Each
// path must only have one plain, one hashed and one encrypted cookie. It does
// not need all or any, but not more than one of the same serialization type.
//
// This allows type-safe storage of complex state in cookies.
package cookies

import (
	"context"
	"net/http"
	"strings"

	"github.com/acudac-com/public-go/jsonx"
	"go.alis.build/alog"
)

var (
	// Edit this on init to your prefered same site settings. Default is lax mode.
	SameSite = http.SameSiteLaxMode
	// Edit this on init to your prefered same site settings. Default is true.
	Secure = true
	// The max age of cookies, default is 400 days.
	MaxAge = 400 * 24 * 60 * 60
	// Whether cookies are http only, which is the default.
	HttpOnly = true
)

// Write the unencrypted, unhashed, base64 json marshalled cookie for the
// specified path.
func WritePlain(ctx context.Context, path string, w http.ResponseWriter, v any) {
	value, err := jsonx.MarshalB64(ctx, v)
	if err != nil {
		return
	}

	name := name("plain", path)
	setCookie(ctx, w, name, string(value), path)
}

// Read the uncrypted, unhashed, base64 json marshalled cookie (if any) for the
// specified path.
func ReadPlain[T any](ctx context.Context, path string, r *http.Request, v T) T {
	name := name("plain", path)
	cookie, err := r.Cookie(name)
	if err != nil {
		return v
	}
	jsonx.UnmarshalB64(ctx, []byte(cookie.Value), v)
	return v
}

// Deletes the encrypted cookie for the specified path.
func DeletePlain(ctx context.Context, path string, w http.ResponseWriter) {
	name := name("plain", path)
	setCookie(ctx, w, name, "", path)
}

// Write the unencrypted, unhashed, base64 json marshalled cookie for the
// specified path.
func WriteHashed(ctx context.Context, path string, w http.ResponseWriter, v any) {
	value, err := jsonx.HashB64(ctx, v)
	if err != nil {
		return
	}
	name := name("hashed", path)
	setCookie(ctx, w, name, string(value), path)
}

// Read the uncrypted, unhashed, base64 json marshalled cookie (if any) for the
// specified path.
func ReadHashed[T any](ctx context.Context, path string, r *http.Request, v T) T {
	name := name("hashed", path)
	cookie, err := r.Cookie(name)
	if err != nil {
		return v
	}
	jsonx.UnhashB64(ctx, []byte(cookie.Value), v)
	return v
}

// Deletes the encrypted cookie for the specified path.
func DeleteHashed(ctx context.Context, path string, w http.ResponseWriter) {
	name := name("hashed", path)
	setCookie(ctx, w, name, "", path)
}

// Write the unencrypted, unhashed, base64 json marshalled cookie for the
// specified path.
func WriteEncrypted(ctx context.Context, path string, w http.ResponseWriter, v any) {
	value, err := jsonx.EncryptB64(ctx, v)
	if err != nil {
		return
	}
	name := name("encrypted", path)
	setCookie(ctx, w, name, string(value), path)
}

// Read the uncrypted, unhashed, base64 json marshalled cookie (if any) for the
// specified path.
func ReadEncrypted[T any](ctx context.Context, path string, r *http.Request, v T) T {
	name := name("encrypted", path)
	cookie, err := r.Cookie(name)
	if err != nil {
		return v
	}

	jsonx.DecryptB64(ctx, []byte(cookie.Value), v)
	return v
}

// Deletes the encrypted cookie for the specified path.
func DeleteEncrypted(ctx context.Context, path string, w http.ResponseWriter) {
	name := name("encrypted", path)
	setCookie(ctx, w, name, "", path)
}

// Returns the name of the cookie at the specified path.
func name(prefix string, path string) string {
	name := prefix + strings.ReplaceAll(path, "/", "_")
	return strings.TrimSuffix(name, "_")
}

func setCookie(ctx context.Context, w http.ResponseWriter, name, value, path string) {
	maxAge := MaxAge
	if value == "" {
		maxAge = -1 // Delete the cookie by setting MaxAge to -1
	}
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: HttpOnly,
		SameSite: SameSite,
		Secure:   Secure,
		MaxAge:   maxAge,
	}
	if err := cookie.Valid(); err != nil {
		alog.Errorf(ctx, "invalid cookie: %v", err)
		return
	}
	alog.Debugf(ctx, "setting cookie: %v", cookie)
	println("done with cookie")
	http.SetCookie(w, cookie)
}
