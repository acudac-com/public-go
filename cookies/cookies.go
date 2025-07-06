package cookies

import "net/http"

type Cookie struct {
	name     string
	path     string
	secure   bool
	httpOnly bool
	sameSite http.SameSite
}

func New(name, path string, secure, httpOnly bool, sameSite http.SameSite) *Cookie {
	return &Cookie{name, path, secure, httpOnly, sameSite}
}

func (c *Cookie) Name() string {
	return c.name
}

func (c *Cookie) Clear(w http.ResponseWriter) {
	c.Set(w, "", -1)
}

func (c *Cookie) Set(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.name,
		Value:    value,
		Path:     c.path,
		MaxAge:   maxAge,
		Secure:   c.secure,
		HttpOnly: c.httpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}
