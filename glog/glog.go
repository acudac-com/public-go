package glog

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type CtxKey string

const TraceCtxKey CtxKey = "trace"

type SlogHandler struct {
	slog.Handler
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
	var trace string
	traceVal := ctx.Value(TraceCtxKey)
	if traceVal != nil {
		trace = traceVal.(string)
	}
	if trace != "" {
		rec = rec.Clone()
		// Add trace ID	to the record so it is correlated with the Cloud Run request log
		// See https://cloud.google.com/trace/docs/trace-log-integration
		rec.Add("logging.googleapis.com/trace", slog.StringValue(trace))
	}
	return h.Handler.Handle(ctx, rec)
}

// Middleware that adds the Cloud Trace ID to the context
// This is used to correlate the structured logs with the Cloud Run
// request log.
func WithCloudTraceContext(projectId string, r *http.Request) {
	var trace string
	traceHeader := r.Header.Get("X-Cloud-Trace-Context")
	traceParts := strings.Split(traceHeader, "/")
	if len(traceParts) > 0 && len(traceParts[0]) > 0 {
		trace = fmt.Sprintf("projects/%s/traces/%s", projectId, traceParts[0])
	}
	*r = *r.WithContext(context.WithValue(r.Context(), TraceCtxKey, trace))
}
