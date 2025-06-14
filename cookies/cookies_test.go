package cookies

import (
	"net/http"
	"testing"
)

type Cookie struct {
	Path string
	Type string
}

func TestAll(t *testing.T) {
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		plain := &Cookie{r.URL.Path, "plain"}
		hashed := &Cookie{r.URL.Path, "hashed"}
		encrypted := &Cookie{r.URL.Path, "encrypted"}
		WritePlain(r.Context(), r.URL.Path, w, plain)
		WriteHashed(r.Context(), r.URL.Path, w, hashed)
		WriteEncrypted(r.Context(), r.URL.Path, w, encrypted)
	})
	http.HandleFunc("/about/read", func(w http.ResponseWriter, r *http.Request) {
		plain := ReadPlain(r.Context(), "/about", r, &Cookie{})
		hashed := ReadHashed(r.Context(), "/about", r, &Cookie{})
		encrypted := ReadEncrypted(r.Context(), "/about", r, &Cookie{})
		w.Write([]byte("Plain: " + plain.Path + ", " + plain.Type + "\n" +
			"Hashed: " + hashed.Path + ", " + hashed.Type + "\n" +
			"Encrypted: " + encrypted.Path + ", " + encrypted.Type + "\n"))
	})
	http.ListenAndServe(":8080", nil)
}

func BenchmarkName(b *testing.B) {
	for b.Loop() {
		name("plain", "/customers/123")
	}
}
