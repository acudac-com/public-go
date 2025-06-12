package kms

import (
	"bytes"
	"testing"

	"github.com/acudac-com/public-go/cx"
)

func BenchmarkKeyId(b *testing.B) {
	cx, _ := cx.Now(b.Context())
	expectedId := keyId(cx)
	b.Log(*expectedId)
	for b.Loop() {
		id := keyId(cx)
		if *id != *expectedId {
			b.Fatalf("expected key id %s, got %s", *expectedId, *id)
		}
	}
}

func BenchmarkGenerateKey(b *testing.B) {
	for b.Loop() {
		key, err := generateKey()
		if err != nil {
			b.Fatal(err)
		}
		if len(key) != 96 {
			b.Fatalf("expected key length 96, got %d", len(key))
		}
		_ = key
	}
}

func BenchmarkKey(b *testing.B) {
	cx, _ := cx.Now(b.Context())
	id := keyId(cx)
	originalK, err := key(cx, id, true)
	if err != nil {
		b.Fatal(err)
	}
	if len(originalK) != 96 {
		b.Fatalf("expected key length 96, got %d", len(originalK))
	}
	for b.Loop() {
		k, err := key(cx, id, true)
		if err != nil {
			b.Fatal(err)
		}
		if !bytes.Equal(originalK, k) {
			b.Fatal("keys do not match")
		}
	}
}

func TestEncryption(t *testing.T) {
	data := []byte("test data")
	encr, err := Encrypt(t.Context(), data)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}
	decr, err := Decrypt(t.Context(), encr)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if !bytes.Equal(data, decr) {
		t.Fatalf("decrypted data does not match original: got %s, want %s", decr, data)
	}
}

func BenchmarkEncryption(b *testing.B) {
	data := []byte("test data")
	cx := cx.New(b.Context())
	for b.Loop() {
		encrypted, err := Encrypt(cx, data)
		if err != nil {
			b.Fatal(err)
		}
		if len(encrypted) == 0 {
			b.Fatal("encryption returned empty result")
		}
	}
}

func BenchmarkDecryption(b *testing.B) {
	data := []byte("test data")
	cx := cx.New(b.Context())
	encrypted, err := Encrypt(cx, data)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		decrypted, err := Decrypt(cx, encrypted)
		if err != nil {
			b.Fatal(err)
		}
		_ = decrypted
		// if !bytes.Equal(decrypted, data) {
		// 	b.Fatalf("decrypted data does not match original: got %s, want %s", decrypted, data)
		// }
	}
}

func TestHashing(t *testing.T) {
	data := []byte("test data")
	encr, err := Hash(t.Context(), data)
	if err != nil {
		t.Fatalf("hashing failed: %v", err)
	}
	decr, err := Unhash(t.Context(), encr)
	if err != nil {
		t.Fatalf("unhashing failed: %v", err)
	}
	if !bytes.Equal(data, decr) {
		t.Fatalf("unhashed data does not match original: got %s, want %s", decr, data)
	}
}

func BenchmarkHash(b *testing.B) {
	data := []byte("test data")
	cx := cx.New(b.Context())
	for b.Loop() {
		hashed, err := Hash(cx, data)
		if err != nil {
			b.Fatal(err)
		}
		if len(hashed) == 0 {
			b.Fatal("hashing returned empty result")
		}
	}
}

func BenchmarkUnhash(b *testing.B) {
	data := []byte("test data")
	cx := cx.New(b.Context())
	hashed, err := Hash(cx, data)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		unhashed, err := Unhash(cx, hashed)
		if err != nil {
			b.Fatal(err)
		}
		_ = unhashed
		// if !bytes.Equal(unhashed, data) {
		// 	b.Fatalf("unhashed data does not match original: got %s, want %s", unhashed, data)
		// }
	}
}
