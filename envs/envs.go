// Package envs provides convenience functions for accessing environment variables.
// It also provides a few commonly used environment variables.
package envs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Dev returns true if the DEV environment variable is set to true.
func Dev() bool {
	return OptionalBool("DEV", false)
}

// Debug returns true if the DEBUG environment variable is set to true.
func Debug() bool {
	return OptionalBool("DEBUG", false)
}

// Host returns the hostname of the machine.
// Note that if no hostname is set, the function will panic.
func Host() string {
	return RequiredStringFromFunc(os.Hostname)
}

// Home returns the home directory of the user.
// Note that if no home directory is set, the function will panic.
func Home() string {
	return RequiredStringFromFunc(os.UserHomeDir)
}

// Version returns the value of the VERSION environment variable, or the given default value if it does not exist.
func Version(defaultValue string) string {
	return OptionalString("VERSION", defaultValue)
}

// RequiredString returns the string value of the given env, or panics if it does not exist.
func RequiredString(key string) string {
	value, found := os.LookupEnv(key)
	if !found {
		panic(fmt.Sprintf("No %s string environment variable found", key))
	}
	return value
}

// OptionalString returns the string value of the given env, or defaults to the given fallback if it does not exist.
func OptionalString(key string, fallback string) string {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback
	}
	return value
}

// RequiredBool returns the string value of the given env, or panics if it does not exist.
func RequiredBool(key string) bool {
	value, found := os.LookupEnv(key)
	if !found {
		panic(fmt.Sprintf("No %s bool environment variable found", key))
	}
	switch value {
	case "true":
		return true
	case "false":
		return false
	default:
		panic(fmt.Sprintf("Invalid bool environment variable %s=%s", key, value))
	}
}

// OptionalBool returns the string value of the given env, or defaults to the given fallback if it does not exist.
func OptionalBool(key string, fallback bool) bool {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback
	}
	switch value {
	case "true":
		return true
	case "false":
		return false
	default:
		panic(fmt.Sprintf("Invalid bool environment variable %s=%s", key, value))
	}
}

// RequiredStringFromFunc returns the string from the given function and panics if an error is returned
func RequiredStringFromFunc(f func() (string, error)) string {
	value, err := f()
	if err != nil {
		panic(fmt.Sprintf("%T: %v", f, err))
	}
	return value
}

// OptionalStringList returns a slice of strings from the given comma-seperated string env.
func OptionalStringList(key string) []string {
	strValue := OptionalString(key, "")
	if strValue == "" {
		return nil
	}
	return strings.Split(strValue, ",")
}

// RequiredInt returns the integer value of the given env, or panics if it does not exist.
func RequiredInt(key string) int {
	value, found := os.LookupEnv(key)
	if !found {
		panic(fmt.Sprintf("No %s int environment variable found", key))
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("Invalid int environment variable %s=%s", key, value))
	}
	return intValue
}

// OptionalInt returns the integer value of the given env, or defaults to the given fallback if it does not exist.
func OptionalInt(key string, fallback int) int {
	value, found := os.LookupEnv(key)
	if !found {
		return fallback
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("Invalid int environment variable %s=%s", key, value))
	}
	return intValue
}
