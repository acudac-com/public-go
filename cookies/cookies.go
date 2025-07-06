package cookies

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/acudac-com/public-go/encoding/jsonx"
	"github.com/acudac-com/public-go/kms"
)

type Cookie struct {
	name   string
	path   string
	maxAge int
	opts   *Opts
}

type Opts struct {
	NotHttpOnly bool // default is false, i.e. HttpOnly
	SameSite    http.SameSite
	Insecure    bool // default is false, i.e. Secure
}

func New(name, path string, maxAge int, opts *Opts) *Cookie {
	if opts == nil {
		opts = &Opts{}
	}
	return &Cookie{name, path, maxAge, opts}
}

// Returns cookie value or empty string if not found
func (c *Cookie) Read(ctx context.Context, r *http.Request) string {
	cookie, err := r.Cookie(c.name)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}

func (c *Cookie) Delete(ctx context.Context, w http.ResponseWriter) {
	c.set(w, "", -1)
}

func (c *Cookie) Set(ctx context.Context, w http.ResponseWriter, value string) {
	c.set(w, value, c.maxAge)
}

func (c *Cookie) set(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.name,
		Value:    value,
		Path:     c.path,
		MaxAge:   maxAge,
		Secure:   !c.opts.Insecure,
		HttpOnly: !c.opts.NotHttpOnly,
		SameSite: c.opts.SameSite,
	})
}

type JsonCookie[T any] struct {
	*Cookie
}

func NewJson[T any](name, path string, maxAge int, opts *Opts) *JsonCookie[T] {
	c := New(name, path, maxAge, opts)
	return &JsonCookie[T]{c}
}

type JsonEncryptedCookie[T any] struct {
	*Cookie
	kms *kms.Kms
}

func (c *JsonCookie[T]) Read(ctx context.Context, r *http.Request, value T) T {
	cookie, err := r.Cookie(c.name)
	if err == nil && cookie.Value != "" {
		if _, err := jsonx.UnmarshalB64([]byte(cookie.Value), value); err != nil {
			slog.WarnContext(ctx, "unmarshalling json cookie", "name", c.name, "value", cookie.Value)
		}
	}
	return value
}

func (c *JsonCookie[T]) Set(ctx context.Context, w http.ResponseWriter, value T) {
	marshalledValue := jsonx.B64Marshal(value)
	c.set(w, string(marshalledValue), c.maxAge)
}

func NewJsonEncrypted[T any](name, path string, maxAge int, kms *kms.Kms, opts *Opts) *JsonEncryptedCookie[T] {
	c := New(name, path, maxAge, opts)
	return &JsonEncryptedCookie[T]{c, kms}
}

func (c *JsonEncryptedCookie[T]) Read(ctx context.Context, r *http.Request, value T) T {
	cookie, err := r.Cookie(c.name)
	if err == nil && cookie.Value != "" {
		if err := c.kms.JsonDecryptB64(ctx, []byte(cookie.Value), value); err != nil {
			slog.WarnContext(ctx, "decrypting cookie", "name", c.name, "value", cookie.Value)
		}
	}
	return value
}

func (c *JsonEncryptedCookie[T]) Set(ctx context.Context, w http.ResponseWriter, value T) {
	encryptedValue := c.kms.B64EncryptJson(ctx, value)
	c.set(w, string(encryptedValue), c.maxAge)
}
