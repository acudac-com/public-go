package glog

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type ctxKey string

const CtxKey ctxKey = "glog"

type ctxVal struct {
	requestId string
	trace     string
}

type SlogHandler struct {
	slog.Handler
}

// Adds request id to ctx so all logs will have "requestId" in their json payload.
// Extracts the "X-Cloud-Trace-Context" header from the request and adds it to
// the request's ctx.
func NewCtx(requestId string, projectId string, r *http.Request) context.Context {
	traceHeader := r.Header.Get("X-Cloud-Trace-Context")
	traceParts := strings.Split(traceHeader, "/")
	var ctx = r.Context()
	ctxVal := &ctxVal{requestId, ""}
	if len(traceParts) > 0 && len(traceParts[0]) > 0 {
		ctxVal.trace = fmt.Sprintf("projects/%s/traces/%s", projectId, traceParts[0])
	}
	ctx = context.WithValue(ctx, CtxKey, ctxVal)
	return ctx
}

func NewSlogHandler(level slog.Level) *SlogHandler {
	return &SlogHandler{slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Only replace top-level attributes
			if groups != nil {
				return a
			}
			switch a.Key {
			case slog.MessageKey:
				a.Key = "message"
			case slog.SourceKey:
				a.Key = "logging.googleapis.com/sourceLocation"
			}
			return a
		},
	})}
}

func (h *SlogHandler) Handle(ctx context.Context, rec slog.Record) error {
	val := ctx.Value(CtxKey)
	if val != nil {
		ctxVal := val.(*ctxVal)
		rec.Add("requestId", ctxVal.requestId)
		if ctxVal.trace != "" {
			rec.Add("logging.googleapis.com/trace", ctxVal.trace)
		}
	}
	return h.Handler.Handle(ctx, rec)
}
