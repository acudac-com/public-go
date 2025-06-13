package jsonx

import (
	"context"
	"encoding/json"

	"github.com/acudac-com/public-go/b64"
	"github.com/acudac-com/public-go/kms"
)

type Jsonx[T any] struct {
	Marshal   func(ctx context.Context, payload T) ([]byte, error)
	Unmarshal func(ctx context.Context, data []byte, payload T) (T, error)
}

func Marshaller[T any]() *Jsonx[T] {
	return &Jsonx[T]{Marshal[T], Unmarshal[T]}
}

func B64Marshaller[T any]() *Jsonx[T] {
	return &Jsonx[T]{MarshalB64[T], UnmarshalB64[T]}
}

func Encrypter[T any]() *Jsonx[T] {
	return &Jsonx[T]{Encrypt[T], Decrypt[T]}
}

func B64Encrypter[T any]() *Jsonx[T] {
	return &Jsonx[T]{EncryptB64[T], DecryptB64[T]}
}

func Hasher[T any]() *Jsonx[T] {
	return &Jsonx[T]{Hash[T], Unhash[T]}
}

func B64Hasher[T any]() *Jsonx[T] {
	return &Jsonx[T]{HashB64[T], UnhashB64[T]}
}

func Marshal[T any](ctx context.Context, payload T) ([]byte, error) {
	return json.Marshal(payload)
}

// Returns url encoded base64 of payload's marshalled json
func MarshalB64[T any](ctx context.Context, payload T) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return b64.UrlEncode(data), nil
}

func Unmarshal[T any](ctx context.Context, data []byte, payload T) (T, error) {
	err := json.Unmarshal(data, payload)
	return payload, err
}

// Returns payload of url encoded base64 json data
func UnmarshalB64[T any](ctx context.Context, data []byte, payload T) (T, error) {
	dec, err := b64.UrlDecode(data)
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(dec, payload)
	return payload, err
}

func Encrypt[T any](ctx context.Context, payload T) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return kms.Encrypt(ctx, data)
}

func EncryptB64[T any](ctx context.Context, payload T) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return kms.EncryptB64(ctx, data)
}

func Decrypt[T any](ctx context.Context, data []byte, payload T) (T, error) {
	decr, err := kms.Decrypt(ctx, data)
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(decr, payload)
	return payload, err
}

func DecryptB64[T any](ctx context.Context, data []byte, payload T) (T, error) {
	decr, err := kms.DecryptB64(ctx, data)
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(decr, payload)
	return payload, err
}

func Hash[T any](ctx context.Context, payload T) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return kms.Hash(ctx, data)
}

func HashB64[T any](ctx context.Context, payload T) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return kms.HashB64(ctx, data)
}

func Unhash[T any](ctx context.Context, data []byte, payload T) (T, error) {
	decr, err := kms.Unhash(ctx, data)
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(decr, payload)
	return payload, err
}

func UnhashB64[T any](ctx context.Context, data []byte, payload T) (T, error) {
	decr, err := kms.UnhashB64(ctx, data)
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(decr, payload)
	return payload, err
}
