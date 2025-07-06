package cookies

import "net/http"

type Cookie struct {
	Name     string
	Path     string
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

func (c *Cookie) Clear(w http.ResponseWriter, value string, maxAge int) {
	c.Set(w, "", -1)
}

func (c *Cookie) Set(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     c.Name,
		Value:    value,
		Path:     c.Path,
		MaxAge:   maxAge,
		Secure:   c.Secure,
		HttpOnly: c.HttpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}
