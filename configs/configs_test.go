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
	c := configs.New(storage)
	defer storage.RemoveFolder(t.Context(), "")
	loadedConfig := &TestConfig{}
	c.Load(loadedConfig)
	t.Logf("empty config: %+v", loadedConfig)
	debugConfig := defaultConfig()
	debugConfig.LogLevel = slog.LevelDebug
	version := c.Upload(defaultConfig(), map[string]any{"debug": debugConfig})
	t.Log(version)
	os.Setenv("CONFIG_VERSION", version)
	os.Setenv("CONFIG_VARIATION", "")
	c.Load(loadedConfig)
	t.Logf("default config: %+v", loadedConfig)
	os.Setenv("CONFIG_VARIATION", "debug")
	c.Load(loadedConfig)
	t.Logf("debug config: %+v", loadedConfig)
}
