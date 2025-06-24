package mux_test

import (
	"net/http"
	"testing"

	"github.com/acudac-com/public-go/mux"
)

func TestHandle(t *testing.T) {
	mux.Handle("GET /hello", hello)
	mux.Handle("GET /missing", missing)
	mux.Handle("GET /nilptr", nilptr)
	mux.ListenAndServe(":8080")
}

func hello(w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("hello"))
	return nil
}

func missing(w http.ResponseWriter, r *http.Request) error {
	return mux.NotFoundErr("not here")
}

func nilptr(w http.ResponseWriter, r *http.Request) error {
	var p *string
	*p = "hi"
	w.Write([]byte(*p))
	return nil
}
