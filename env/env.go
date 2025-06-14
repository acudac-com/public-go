package env

import (
	"context"
	"os"

	"go.alis.build/alog"
)

type environment string

const (
	local environment = "local"
)

var (
	Env     = loadEnvironment()
	Product = loadProduct()
)

func IsLocal() bool {
	return Env == local
}

// reads and validates ENV environment variable.
func loadEnvironment() environment {
	key := "ENV"
	value := environment(os.Getenv(key))
	if value == "" {
		alog.Fatalf(context.Background(), "%s environment variable is required", key)
	}
	return value
}

// reads and validates ENV environment variable.
func loadProduct() string {
	key := "PRODUCT"
	value := os.Getenv(key)
	if value == "" {
		alog.Fatalf(context.Background(), "%s environment variable is required", key)
	}
	return value
}
