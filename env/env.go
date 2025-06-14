package env

import (
	"context"
	"os"

	"go.alis.build/alog"
)

var (
	Env     = RequiredString("ENV")
	Product = RequiredString("PRODUCT")
)

func IsLocal() bool {
	return Env == "local"
}

func RequiredString(key string) string {
	value := os.Getenv(key)
	if value == "" {
		alog.Fatalf(context.Background(), "%s environment variable is required", key)
	}
	return value
}

func OptionalString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}
	return value
}

func RequiredBool(key string) bool {
	value := os.Getenv(key)
	if value == "" {
		alog.Fatalf(context.Background(), "%s environment variable is required", key)
	}
	if value == "true" {
		return true
	} else if value == "false" {
		return false
	} else {
		alog.Fatalf(context.Background(), "%s environment variable must be 'true' or 'false'", key)
		return false // will never execute
	}
}

func OptionalBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	if value == "true" {
		return true
	} else {
		return false
	}
}
