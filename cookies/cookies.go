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

	"github.com/acudac-com/public-go/b64"
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
		alog.Warnf(ctx, "marshalling %s cookie: %v", path, err)
		return
	}

	name := name("plain", path)
	setCookie(w, name, string(value), path)
}

// Read the uncrypted, unhashed, base64 json marshalled cookie (if any) for the
// specified path.
func ReadPlain[T any](ctx context.Context, path string, r *http.Request, v T) T {
	name := name("plain", path)
	cookie, err := r.Cookie(name)
	if err != nil {
		alog.Warnf(ctx, "reading %s cookie: %v", path, err)
		return v
	}

	if _, err := jsonx.UnmarshalB64(ctx, []byte(cookie.Value), v); err != nil {
		alog.Warnf(ctx, "unmarshalling %s cookie: %v", path, err)
	}
	return v
}

// Write the unencrypted, unhashed, base64 json marshalled cookie for the
// specified path.
func WriteHashed(ctx context.Context, path string, w http.ResponseWriter, v any) {
	value, err := jsonx.HashB64(ctx, v)
	if err != nil {
		alog.Warnf(ctx, "hashing %s cookie: %v", path, err)
		return
	}
	name := name("hashed", path)
	setCookie(w, name, string(value), path)
}

// Read the uncrypted, unhashed, base64 json marshalled cookie (if any) for the
// specified path.
func ReadHashed[T any](ctx context.Context, path string, r *http.Request, v T) T {
	name := name("hashed", path)
	cookie, err := r.Cookie(name)
	if err != nil {
		alog.Warnf(ctx, "reading %s cookie: %v", path, err)
		return v
	}

	if _, err := jsonx.UnhashB64(ctx, []byte(cookie.Value), v); err != nil {
		alog.Warnf(ctx, "unhashing %s cookie: %v", path, err)
	}
	return v
}

// Write the unencrypted, unhashed, base64 json marshalled cookie for the
// specified path.
func WriteEncrypted(ctx context.Context, path string, w http.ResponseWriter, v any) {
	value, err := jsonx.EncryptB64(ctx, v)
	if err != nil {
		alog.Warnf(ctx, "encrypting %s encrypted cookie: %v", path, err)
		return
	}
	name := name("encrypted", path)
	setCookie(w, name, string(value), path)
}

// Read the uncrypted, unhashed, base64 json marshalled cookie (if any) for the
// specified path.
func ReadEncrypted[T any](ctx context.Context, path string, r *http.Request, v T) T {
	name := name("encrypted", path)
	cookie, err := r.Cookie(name)
	if err != nil {
		alog.Warnf(ctx, "reading %s cookie: %v", path, err)
		return v
	}

	if _, err := jsonx.DecryptB64(ctx, []byte(cookie.Value), v); err != nil {
		alog.Warnf(ctx, "decrypting %s cookie: %v", path, err)
	}
	return v
}

// Returns the name of the cookie at the specified path.
func name(prefix string, path string) string {
	return prefix + "_" + string(b64.UrlEncode([]byte(path)))
}

func setCookie(w http.ResponseWriter, name, value, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HttpOnly: HttpOnly,
		SameSite: SameSite,
		Secure:   Secure,
		MaxAge:   MaxAge, // 400 days
	})
}
