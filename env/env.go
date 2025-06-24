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

// Returns true if "ENV" environment variables equals "local".
func IsLocal() bool {
	return Env == "local"
}

// Returns the string value of the environment variable at the specified key, or panics
// if its empty.
func RequiredString(key string) string {
	value := os.Getenv(key)
	if value == "" {
		alog.Fatalf(context.Background(), "%s environment variable is required", key)
	}
	return value
}

// Returns the string value of the environment variable at the specified key, or
// the fallback if its empty.
func OptionalString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}
	return value
}

// Returns the boolean value of the environment variable at the specified key,
// or panics if its not equal to "true" or "false".
func RequiredBool(key string) bool {
	value := os.Getenv(key)
	if value == "" {
		alog.Fatalf(context.Background(), "%s environment variable is required", key)
	}
	switch value {
	case "true":
		return true
	case "false":
		return false
	default:
		alog.Fatalf(context.Background(), "%s environment variable must be 'true' or 'false'", key)
		return false // will never execute
	}
}

// Returns the boolean value of the environment variable at the specified key,
// or the fallback if its empty.
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
