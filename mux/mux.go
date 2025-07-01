package mux

import (
	"context"
	"net/http"
)

type Mux[CxT context.Context] struct {
	mux        *http.ServeMux
	middleware func(w http.ResponseWriter, r *http.Request, handler func(cx CxT, w http.ResponseWriter, r *http.Request) error) error
}

func New[CxT context.Context](
	middleware func(w http.ResponseWriter, r *http.Request, handler func(cx CxT, w http.ResponseWriter, r *http.Request) error) error,
) *Mux[CxT] {
	return &Mux[CxT]{http.NewServeMux(), middleware}
}

func (m *Mux[CxT]) Get(pattern string, handleFunc func(cx CxT, w http.ResponseWriter, r *http.Request) error) {
	m.mux.HandleFunc("GET "+pattern, m.handler(handleFunc))
}

func (m *Mux[CxT]) Post(pattern string, handleFunc func(cx CxT, w http.ResponseWriter, r *http.Request) error) {
	m.mux.HandleFunc("POST "+pattern, m.handler(handleFunc))
}

func (m *Mux[CxT]) Patch(pattern string, handleFunc func(cx CxT, w http.ResponseWriter, r *http.Request) error) {
	m.mux.HandleFunc("PATCH "+pattern, m.handler(handleFunc))
}

func (m *Mux[CxT]) Put(pattern string, handleFunc func(cx CxT, w http.ResponseWriter, r *http.Request) error) {
	m.mux.HandleFunc("PUT "+pattern, m.handler(handleFunc))
}

func (m *Mux[CxT]) Delete(pattern string, handleFunc func(cx CxT, w http.ResponseWriter, r *http.Request) error) {
	m.mux.HandleFunc("DELETE "+pattern, m.handler(handleFunc))
}

func (m *Mux[CxT]) handler(handleFunc func(cx CxT, w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := m.middleware(w, r, handleFunc); err != nil {
			if httpErr, ok := err.(*Error); ok {
				http.Error(w, httpErr.msg, httpErr.code)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}
	}
}

func (m *Mux[CxT]) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, m.mux)
}
