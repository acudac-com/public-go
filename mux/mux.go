package mux

import (
	"net/http"
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
		err := handleFunc(w, cxR)
		if m.EnableLogs {
			elapsedTime := time.Since(startT)
			alog.Infof(cxR.Context(), "[%s] [%s] [%s]\n", r.Method, r.URL.Path, elapsedTime)
		}
		if err != nil {
			alog.Errorf(cxR.Context(), "%s %s: %v", r.Method, r.URL.Path, err)
			if httpErr, ok := err.(*HttpErr); ok {
				http.Error(w, httpErr.msg, httpErr.code)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
	})
}

var DefaultMux = &Mux{http.DefaultServeMux, env.IsLocal()}
