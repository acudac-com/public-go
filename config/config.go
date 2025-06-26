package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/acudac-com/public-go/env"
	"github.com/acudac-com/public-go/storage"
	"go.alis.build/alog"
)

var (
	Variation = os.Getenv("CONFIG_VARIATION")
	Version   = os.Getenv("CONFIG_VERSION")
)

func init() {
	ctx := context.Background()
	if Version == "" {
		alog.Warnf(ctx, "no config version specified yet in %s environment so loaded config will be empty", env.Env)
		return
	}
}

// Returns the blob key of the specified version and variation's config.
func key(version string, variation string) string {
	obj := f(".config/%s.cfg", version)
	if variation != "" {
		obj = f(".config/%s/%s.cfg", version, variation)
	}
	return obj
}

// Reads the json encoded config for the given version and variation from blob storage.
func Load[T any](config T) T {
	if Version == "" {
		return config
	}

	ctx := context.Background()
	key := key(Version, Variation)
	reader, err := storage.Reader(ctx, key)
	if err != nil {
		alog.Fatalf(ctx, "creating config reader: %v", err)
	}
	if err := json.NewDecoder(reader).Decode(config); err != nil {
		alog.Fatalf(ctx, "decoding config: %v", err)
	}
	return config
}

// Encodes and writes the given config and its optional variations to blob
// storage. Prints out the version so you can copy and paste it into your
// version constant to use the new config.
func Upload[T any](config T, variations map[string]T) error {
	ctx := context.Background()
	version := time.Now().UTC().Format("2006-01-02_15-04-05")
	configs := variations
	if configs == nil {
		configs = map[string]T{}
	}
	configs[""] = config
	for variation, conf := range configs {
		key := key(version, variation)
		writer, err := storage.Writer(ctx, key)
		if err != nil {
			return e("creating config writer: %w", err)
		}
		if err := json.NewEncoder(writer).Encode(conf); err != nil {
			return e("encoding config: %w", err)
		}
		if err := writer.Close(); err != nil {
			return e("closing config writer: %w", err)
		}
	}
	println("Version:")
	println(version)
	return nil
}

var (
	f = fmt.Sprintf
	e = fmt.Errorf
)
