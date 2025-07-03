package configs_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/acudac-com/public-go/configs"
	"github.com/acudac-com/public-go/storage"
)

type TestConfig struct {
	MyKey    string
	LogLevel slog.Level
}

func defaultConfig() *TestConfig {
	return &TestConfig{
		MyKey: "asdfadsf",
	}
}

func Test_Configs(t *testing.T) {
	storage := storage.NewFsStorage(".storage")
	defer storage.RemoveFolder(t.Context(), "")
	loadedConfig := configs.Load(storage, &TestConfig{})
	t.Logf("empty config: %+v", loadedConfig)
	debugConfig := defaultConfig()
	debugConfig.LogLevel = slog.LevelDebug
	version := configs.Upload(storage, defaultConfig(), map[string]any{"debug": debugConfig})
	t.Log(version)
	os.Setenv("CONFIG_VERSION", version)
	os.Setenv("CONFIG_VARIATION", "")
	configs.Load(storage, loadedConfig)
	t.Logf("default config: %+v", loadedConfig)
	os.Setenv("CONFIG_VARIATION", "debug")
	configs.Load(storage, loadedConfig)
	t.Logf("debug config: %+v", loadedConfig)
}
