package configs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/acudac-com/public-go/storage"
)

// Returns the blob key of the specified version and variation's config.
func key(version string, variation string) string {
	obj := fmt.Sprintf(".config/%s.cfg", version)
	if variation != "" {
		obj = fmt.Sprintf(".config/%s/%s.cfg", version, variation)
	}
	return obj
}

// Uses CONFIG_VERSION and CONFIG_VARIATION environment variables to load the
// config. Loads nothing if version is empty.
func Load[T any](storage storage.Storage, config T) T {
	Variation := os.Getenv("CONFIG_VARIATION")
	Version := os.Getenv("CONFIG_VERSION")
	if Version == "" {
		slog.Warn("no config version specified in CONFIG_VERSION env so loaded config will be empty")
		return config
	}

	ctx := context.Background()
	key := key(Version, Variation)
	reader, found := storage.Reader(ctx, key)
	if !found {
		panic(fmt.Errorf("configs.Configs.Load(): config not found for version %q and variation %q", Version, Variation))
	}
	if err := json.NewDecoder(reader).Decode(config); err != nil {
		panic(fmt.Errorf("configs.Config.Load() decoding config: %w", err))
	}
	return config
}

// Uploads the config alog with the specified variations to a new version and
// returns the version.
func Upload(storage storage.Storage, config any, variations map[string]any) string {
	ctx := context.Background()
	version := time.Now().UTC().Format("2006-01-02_15-04-05")
	configs := variations
	if configs == nil {
		configs = map[string]any{}
	}
	configs[""] = config
	for variation, conf := range configs {
		key := key(version, variation)
		writer := storage.Writer(ctx, key)
		if err := json.NewEncoder(writer).Encode(conf); err != nil {
			panic(fmt.Errorf("configs.Configs.Upload(): encoding config: %w", err))
		}
		if err := writer.Close(); err != nil {
			panic(fmt.Errorf("configs.Configs.Upload(): closing writer: %w", err))
		}
	}
	return version
}
