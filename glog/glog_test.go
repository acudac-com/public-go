package glog_test

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/acudac-com/public-go/glog"
)

func init() {
	slogHandler := glog.NewSlogHandler(slog.LevelDebug)
	slog.SetDefault(slog.New(slogHandler))
}

func TestAny(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	ctx := glog.NewCtx("someid", "someproj", r, "userId", "12345", "accountType", 5)
	type SomeStruct struct {
		Id  string `json:"id"`
		Age int32  `json:"age"`
	}
	ss := &SomeStruct{"asdf", 32}
	slog.InfoContext(ctx, "testing", "stringval", "asdf", "intval", 35, slog.Any("some struct", ss))
}
