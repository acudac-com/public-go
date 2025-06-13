package gobx

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/acudac-com/public-go/b64"
	"github.com/acudac-com/public-go/kms"
)

type Gobx[T any] struct {
	Marshal   func(ctx context.Context, payload T) ([]byte, error)
	Unmarshal func(ctx context.Context, data []byte, payload T) (T, error)
}

func Marshaller[T any]() *Gobx[T] {
	return &Gobx[T]{Marshal[T], Unmarshal[T]}
}

func B64Marshaller[T any]() *Gobx[T] {
	return &Gobx[T]{MarshalB64[T], UnmarshalB64[T]}
}

func Encrypter[T any]() *Gobx[T] {
	return &Gobx[T]{Encrypt[T], Decrypt[T]}
}

func B64Encrypter[T any]() *Gobx[T] {
	return &Gobx[T]{EncryptB64[T], DecryptB64[T]}
}

func Hasher[T any]() *Gobx[T] {
	return &Gobx[T]{Hash[T], Unhash[T]}
}

func B64Hasher[T any]() *Gobx[T] {
	return &Gobx[T]{HashB64[T], UnhashB64[T]}
}

func Marshal[T any](ctx context.Context, payload T) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(payload)
	return buf.Bytes(), err
}

// Returns url encoded base64 of payload's marshalled gob
func MarshalB64[T any](ctx context.Context, payload T) ([]byte, error) {
	enc, err := Marshal(ctx, payload)
	if err != nil {
		return nil, err
	}
	return b64.UrlEncode(enc), nil
}

func Unmarshal[T any](ctx context.Context, data []byte, payload T) (T, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&payload)
	return payload, err
}

// Returns payload of url encoded base64 gob data
func UnmarshalB64[T any](ctx context.Context, data []byte, payload T) (T, error) {
	dec, err := b64.UrlDecode(data)
	if err != nil {
		return payload, err
	}
	return Unmarshal(ctx, dec, payload)
}

func Encrypt[T any](ctx context.Context, payload T) ([]byte, error) {
	data, err := Marshal(ctx, payload)
	if err != nil {
		return nil, err
	}
	return kms.Encrypt(ctx, data)
}

func EncryptB64[T any](ctx context.Context, payload T) ([]byte, error) {
	data, err := Marshal(ctx, payload)
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
	return Unmarshal(ctx, decr, payload)
}

func DecryptB64[T any](ctx context.Context, data []byte, payload T) (T, error) {
	decr, err := kms.DecryptB64(ctx, data)
	if err != nil {
		return payload, err
	}
	return Unmarshal(ctx, decr, payload)
}

func Hash[T any](ctx context.Context, payload T) ([]byte, error) {
	data, err := Marshal(ctx, payload)
	if err != nil {
		return nil, err
	}
	return kms.Hash(ctx, data)
}

func HashB64[T any](ctx context.Context, payload T) ([]byte, error) {
	data, err := Marshal(ctx, payload)
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
	return Unmarshal(ctx, decr, payload)
}

func UnhashB64[T any](ctx context.Context, data []byte, payload T) (T, error) {
	decr, err := kms.UnhashB64(ctx, data)
	if err != nil {
		return payload, err
	}
	return Unmarshal(ctx, decr, payload)
}
