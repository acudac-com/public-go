package glog_test

import (
	"log/slog"
	"testing"

	"github.com/acudac-com/public-go/glog"
)

func init() {
	slogHandler := glog.NewSlogHandler(slog.LevelDebug)
	slog.SetDefault(slog.New(slogHandler))
}

func TestAny(t *testing.T) {
	type SomeStruct struct {
		Id  string `json:"id"`
		Age int32  `json:"age"`
	}
	ss := &SomeStruct{"asdf", 32}
	slog.Info("testing", "stringval", "asdf", "intval", 35, slog.Any("some struct", ss))
}
