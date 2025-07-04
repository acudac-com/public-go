package kms

import (
	"bytes"
	"testing"

	"github.com/acudac-com/public-go/storage"
	"github.com/acudac-com/public-go/timex"
)

var kms *Kms

func init() {
	storage := storage.NewFsStorage("")
	kms = New(storage, nil)
}

func BenchmarkKeyId(b *testing.B) {
	cx, _ := timex.Now(b.Context())
	expectedId := kms.currentKeyId(cx)
	b.Log(*expectedId)
	for b.Loop() {
		_ = kms.currentKeyId(cx)
	}
}

func BenchmarkGenerateKey(b *testing.B) {
	for b.Loop() {
		_ = generateKey()
	}
}

func BenchmarkKey(b *testing.B) {
	cx, _ := timex.Now(b.Context())
	id := kms.currentKeyId(cx)
	originalK, err := kms.key(cx, id, true)
	if err != nil {
		b.Fatal(err)
	}
	if len(originalK) != 96 {
		b.Fatalf("expected key length 96, got %d", len(originalK))
	}
	for b.Loop() {
		k, err := kms.key(cx, id, true)
		if err != nil {
			b.Fatal(err)
		}
		if !bytes.Equal(originalK, k) {
			b.Fatal("keys do not match")
		}
	}
}

func TestEncryption(t *testing.T) {
	data := []byte("Nobody needs to pay big cloud providers a fortunate to manage their keys. The logic is dead simple; DIY!")
	encr := kms.B64Encrypt(t.Context(), data)
	t.Log(string(encr))
	decr, err := kms.DecryptB64(t.Context(), encr)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if !bytes.Equal(data, decr) {
		t.Fatalf("decrypted data does not match original: got %s, want %s", decr, data)
	}
}

func BenchmarkEncryption(b *testing.B) {
	data := []byte("test data")
	cx, _ := timex.Now(b.Context())
	for b.Loop() {
		_ = kms.Encrypt(cx, data)
	}
}

func BenchmarkDecryption(b *testing.B) {
	data := []byte("test data")
	cx, _ := timex.Now(b.Context())
	encrypted := kms.Encrypt(cx, data)
	for b.Loop() {
		_, _ = kms.Decrypt(cx, encrypted)
	}
}

func TestHashing(t *testing.T) {
	data := []byte("test data")
	encr := kms.B64Hash(t.Context(), data)
	t.Log(string(encr))
	decr, err := kms.UnhashB64(t.Context(), encr)
	if err != nil {
		t.Fatalf("unhashing failed: %v", err)
	}
	if !bytes.Equal(data, decr) {
		t.Fatalf("unhashed data does not match original: got %s, want %s", decr, data)
	}
}

func BenchmarkHash(b *testing.B) {
	data := []byte("test data")
	cx, _ := timex.Now(b.Context())
	for b.Loop() {
		hashed := kms.Hash(cx, data)
		if len(hashed) == 0 {
			b.Fatal("hashing returned empty result")
		}
	}
}

func BenchmarkUnhash(b *testing.B) {
	data := []byte("test data")
	cx, _ := timex.Now(b.Context())
	hashed := kms.Hash(cx, data)
	for b.Loop() {
		_, _ = kms.Unhash(cx, hashed)
	}
}
