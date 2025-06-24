package mux

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/acudac-com/public-go/cx"
	"github.com/acudac-com/public-go/env"
	"go.alis.build/alog"
)

type Mux struct {
	mux        *http.ServeMux
	EnableLogs bool
}

func New(enableLogs bool) *Mux {
	return &Mux{mux: http.NewServeMux()}
}

func (m *Mux) Handle(pattern string, handleFunc func(w http.ResponseWriter, r *http.Request) error) {
	m.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		newCtx := cx.New(r.Context())
		cxR := r.WithContext(newCtx)
		var startT time.Time
		if m.EnableLogs {
			startT = cx.Now(cxR.Context())
		}
		defer func() {
			if rec := recover(); rec != nil {
				elapsedTime := time.Since(startT)
				details := ""
				stack := make([]byte, 4096)
				length := runtime.Stack(stack, false)
				stackStr := string(stack[:length])
				if err, ok := rec.(error); ok {
					details = err.Error() + "\n" + stackStr
				} else {
					details = fmt.Sprintf("%v\n%s", rec, stackStr)
				}
				alog.Errorf(cxR.Context(), "[%s] [%s] [%s]\n%v", r.Method, r.URL.Path, elapsedTime, details)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		err := handleFunc(w, cxR)
		if err != nil {
			elapsedTime := time.Since(startT)
			alog.Errorf(cxR.Context(), "[%s] [%s] [%s]\n%s", r.Method, r.URL.Path, elapsedTime, err.Error())
			if httpErr, ok := err.(*HttpErr); ok {
				http.Error(w, httpErr.msg, httpErr.code)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		} else if m.EnableLogs {
			elapsedTime := time.Since(startT)
			alog.Infof(cxR.Context(), "[%s] [%s] [%s]\n", r.Method, r.URL.Path, elapsedTime)
		}
	})
}

func (m *Mux) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, m.mux)
}

var DefaultMux = &Mux{http.DefaultServeMux, env.IsLocal()}

func Handle(pattern string, handleFunc func(w http.ResponseWriter, r *http.Request) error) {
	DefaultMux.Handle(pattern, handleFunc)
}

func ListenAndServe(addr string) error {
	return DefaultMux.ListenAndServe(addr)
}
