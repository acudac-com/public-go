package kms

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"sync"
	"time"

	"github.com/acudac-com/public-go/b64"
	"github.com/acudac-com/public-go/cx"
	"github.com/acudac-com/public-go/storage"
)

const KeyFormat string = "20060102_1504"

var (
	// How often key is rotated. Default is 30 days.
	RotationFreq = 30 * 24 * time.Hour
	// How long a key is valid for. Default is 90 days.
	MaxAge = 90 * 24 * time.Hour
	// Concurrency proof cache of keys.
	keys = &sync.Map{}

	// aliases
	e = fmt.Errorf
)

var keyLength = len(KeyFormat)

// Determines the current key id to use for hashing/encrypting data
func keyId(ctx context.Context) *string {
	now := cx.Now(ctx)
	t := now.Truncate(RotationFreq)
	id := t.Format(KeyFormat) // Format as YYYYMMDD_HHMMSS
	return &id
}

// Determines if the provided key is valid and not expired.
func validateKeyId(ctx context.Context, id string) error {
	now := cx.Now(ctx)
	t, err := time.Parse(KeyFormat, id)
	if err != nil {
		return e("invalid key id format: %s, expected format: %s", id, KeyFormat)
	}
	if t.After(now) {
		return e("key id %s is in the future", id)
	}
	if t.Add(MaxAge).Before(now) {
		return e("key id %s is expired, max age is %s", id, MaxAge)
	}
	return nil
}

// Generates a new key 96 byte key: 32 for encryption, 64 for HMAC"
func generateKey() ([]byte, error) {
	key := make([]byte, 96)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// Returns the storage key based on the provided key id
func blobKey(keyId string) string {
	return ".keys/" + keyId
}

// Writes a newly generated key to storage for the specified storage id (not key
// id). Will first attempt write, but if fails due to already existing key, it
// will simply read and return the existing key. Does NOT consult the keys
// sync.Map cache.
func writeOrReadExisting(ctx context.Context, blobId *string, data []byte) ([]byte, error) {
	if err := storage.WriteIfMissing(ctx, *blobId, data); err != nil {
		if _, ok := err.(*storage.AlreadyExistsError); !ok {
			return nil, e("writing generated key to storage: %w", err)
		}
		if data, err = storage.Read(ctx, *blobId); err != nil {
			return nil, e("reading already existing key from storage: %w", err)
		}
	}
	return data, nil
}

// Returns the key with the specified id if it exists
func key(ctx context.Context, id *string, createIfMissing bool) ([]byte, error) {
	if !createIfMissing {
		if err := validateKeyId(ctx, *id); err != nil {
			return nil, err
		}
	}
	var key []byte
	val, ok := keys.Load(*id)
	if !ok {
		var err error
		blobKey := blobKey(*id)
		if createIfMissing {
			if key, err = generateKey(); err != nil {
				return nil, err
			}
			if key, err = writeOrReadExisting(ctx, &blobKey, key); err != nil {
				return nil, err
			}
		} else {
			if key, err = storage.Read(ctx, blobKey); err != nil {
				if _, ok := err.(*storage.NotFoundError); ok {
					return nil, e("key not found: " + *id)
				}
				return nil, err
			}
		}
		keys.Store(*id, key)
	} else {
		if createIfMissing {
			if err := validateKeyId(ctx, *id); err != nil {
				return nil, err
			}
		}
		key = val.([]byte)
	}
	return key, nil
}

// Uses first 32 bytes of the current key, as aes requires a 32-byte key for AES-256.
func encryptionKey(ctx context.Context, id *string, createIfMissing bool) ([]byte, error) {
	key, err := key(ctx, id, createIfMissing)
	if err != nil {
		return nil, err
	}
	return key[:32], nil
}

// Use AES-GCM encryption to encrypt data with key used for encryption added as first bytes.
func Encrypt(ctx context.Context, data []byte) ([]byte, error) {
	keyId := keyId(ctx)
	key, err := encryptionKey(ctx, keyId, true)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, e("creating cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, e("creating GCM: %w", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, e("generating nonce: %w", err)
	}
	ciphertext := aead.Seal(nil, nonce, data, nil)

	// Prepend the key id and nonce to the ciphertext
	ciphertext = append(nonce, ciphertext...)
	ciphertext = append([]byte(*keyId), ciphertext...)
	return ciphertext, nil
}

// Use AES-GCM encryption to encrypt data with key used for encryption added as first bytes.
// Returns base64 url encoding of encrypted data.
func EncryptB64(ctx context.Context, data []byte) ([]byte, error) {
	encr, err := Encrypt(ctx, data)
	if err != nil {
		return nil, err
	}
	return b64.UrlEncode(encr), nil
}

// Use AES-GCM decryption to decrypt data with key used for encryption extracted from the first bytes.
func Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	keyId := string(data[:keyLength])
	key, err := encryptionKey(ctx, &keyId, false)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, e("creating cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, e("creating GCM: %w", err)
	}
	nonceSize := aead.NonceSize()
	nonce := data[keyLength : keyLength+nonceSize]
	ciphertext := data[keyLength+nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, e("decrypting: %w", err)
	}
	return plaintext, nil
}

// Use AES-GCM decryption to decrypt base64 url encoded data with key used for encryption extracted from the first bytes.
func DecryptB64(ctx context.Context, data []byte) ([]byte, error) {
	dec, err := b64.UrlDecode(data)
	if err != nil {
		return nil, err
	}
	return Decrypt(ctx, dec)
}

func hashKey(ctx context.Context, keyId *string, createIfMissing bool) ([]byte, error) {
	key, err := key(ctx, keyId, createIfMissing)
	if err != nil {
		return nil, err
	}
	return key[32:], nil
}

// Prepends key id and SHA512 HMAC to data to sign it.
func Hash(ctx context.Context, data []byte) ([]byte, error) {
	keyId := keyId(ctx)
	key, err := hashKey(ctx, keyId, true)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha512.New, key)
	mac.Write(data)
	result := mac.Sum(nil)
	result = append(result, data...)
	result = append([]byte(*keyId), result...)
	return result, nil
}

// Prepends key id and SHA512 HMAC to data to sign it. Returns base64 url encoding of final hashed data.
func HashB64(ctx context.Context, data []byte) ([]byte, error) {
	hashed, err := Hash(ctx, data)
	if err != nil {
		return nil, err
	}
	return b64.UrlEncode(hashed), nil
}

// Verifies the HMAC SHA512 signature of the data and returns original payload.
func Unhash(ctx context.Context, data []byte) ([]byte, error) {
	keyId := string(data[:keyLength])
	macStart := keyLength
	macEnd := macStart + sha512.Size
	macBytes := data[macStart:macEnd]
	payload := data[macEnd:]

	key, err := hashKey(ctx, &keyId, false)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha512.New, key)
	mac.Write(payload)
	expectedMac := mac.Sum(nil)
	if !hmac.Equal(macBytes, expectedMac) {
		return nil, e("invalid HMAC signature for key id %s", keyId)
	}
	return payload, nil
}

// Verifies the HMAC SHA512 signature of the base64 url encoded data and returns original payload.
func UnhashB64(ctx context.Context, data []byte) ([]byte, error) {
	dec, err := b64.UrlDecode(data)
	if err != nil {
		return nil, err
	}
	return Unhash(ctx, dec)
}
