package mux

import (
	"context"
	"net/http"
)

type (
	Mux[CxT context.Context, OptsT any] struct {
		mux     *http.ServeMux
		gateway Gateway[CxT, OptsT]
	}
	Handler[CxT context.Context, OptsT any] func(cx CxT, w http.ResponseWriter, r *http.Request, opts OptsT) error
	Gateway[CxT context.Context, OptsT any] func(w http.ResponseWriter, r *http.Request, handler Handler[CxT, OptsT], opts OptsT) error
)

func New[CxT context.Context, OptsT any](gateway Gateway[CxT, OptsT]) *Mux[CxT, OptsT] {
	return &Mux[CxT, OptsT]{http.NewServeMux(), gateway}
}

func (m *Mux[CxT, OptsT]) Get(pattern string, handleFunc Handler[CxT, OptsT], opts OptsT) {
	m.mux.HandleFunc("GET "+pattern, m.handler(handleFunc, opts))
}

func (m *Mux[CxT, OptsT]) Post(pattern string, handleFunc Handler[CxT, OptsT], opts OptsT) {
	m.mux.HandleFunc("POST "+pattern, m.handler(handleFunc, opts))
}

func (m *Mux[CxT, OptsT]) Patch(pattern string, handleFunc Handler[CxT, OptsT], opts OptsT) {
	m.mux.HandleFunc("PATCH "+pattern, m.handler(handleFunc, opts))
}

func (m *Mux[CxT, OptsT]) Put(pattern string, handleFunc Handler[CxT, OptsT], opts OptsT) {
	m.mux.HandleFunc("PUT "+pattern, m.handler(handleFunc, opts))
}

func (m *Mux[CxT, OptsT]) Delete(pattern string, handleFunc Handler[CxT, OptsT], opts OptsT) {
	m.mux.HandleFunc("DELETE "+pattern, m.handler(handleFunc, opts))
}

func (m *Mux[CxT, OptsT]) handler(handleFunc Handler[CxT, OptsT], opts OptsT) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := m.gateway(w, r, handleFunc, opts); err != nil {
			if httpErr, ok := err.(*Error); ok {
				http.Error(w, httpErr.msg, httpErr.code)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}
	}
}

func (m *Mux[CxT, OptsT]) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, m.mux)
}
