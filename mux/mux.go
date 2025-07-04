package mux

import (
	"context"
	"net/http"
)

type (
	Mux[CxT context.Context] struct {
		mux     *http.ServeMux
		gateway Gateway[CxT]
	}
	Handler[CxT context.Context] func(cx CxT, w http.ResponseWriter, r *http.Request) error
	Gateway[CxT context.Context] func(w http.ResponseWriter, r *http.Request, handler Handler[CxT], mw ...Handler[CxT]) error
)

func New[CxT context.Context](gateway Gateway[CxT]) *Mux[CxT] {
	return &Mux[CxT]{http.NewServeMux(), gateway}
}

func (m *Mux[CxT]) Get(pattern string, handleFunc Handler[CxT], middleware ...Handler[CxT]) {
	m.mux.HandleFunc("GET "+pattern, m.handler(handleFunc, middleware))
}

func (m *Mux[CxT]) Post(pattern string, handleFunc Handler[CxT], middleware ...Handler[CxT]) {
	m.mux.HandleFunc("POST "+pattern, m.handler(handleFunc, middleware))
}

func (m *Mux[CxT]) Patch(pattern string, handleFunc Handler[CxT], middleware ...Handler[CxT]) {
	m.mux.HandleFunc("PATCH "+pattern, m.handler(handleFunc, middleware))
}

func (m *Mux[CxT]) Put(pattern string, handleFunc Handler[CxT], middleware ...Handler[CxT]) {
	m.mux.HandleFunc("PUT "+pattern, m.handler(handleFunc, middleware))
}

func (m *Mux[CxT]) Delete(pattern string, handleFunc Handler[CxT], middleware ...Handler[CxT]) {
	m.mux.HandleFunc("DELETE "+pattern, m.handler(handleFunc, middleware))
}

func (m *Mux[CxT]) handler(handleFunc Handler[CxT], middleware []Handler[CxT]) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := m.gateway(w, r, handleFunc, middleware...); err != nil {
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
