package env

import (
	"fmt"
	"os"
)

// Returns the string value of the environment variable at the specified key, or panics
// if its empty.
func RequiredString(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("%s environment variable is required", key))
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
		panic(fmt.Sprintf("%s environment variable is required", key))
	}
	switch value {
	case "true":
		return true
	case "false":
		return false
	default:
		panic(fmt.Sprintf("%s environment variable must be 'true' or 'false'", key))
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
