package kms

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/acudac-com/public-go/encoding/b64"
	"github.com/acudac-com/public-go/encoding/jsonx"
	"github.com/acudac-com/public-go/storage"
	"github.com/acudac-com/public-go/timex"
)

type Kms struct {
	storage       storage.Storage
	keys          *sync.Map
	keyLength     int
	rotationFreq  time.Duration
	maxAge        time.Duration
	keyTimeFormat string
	keysFolder    string
}

type Opts struct {
	RotationFreq  time.Duration // default is 30 days
	MaxAge        time.Duration // default is 90 days
	KeyTimeFormat string        // default is "20060102_1504"
	KeysFolder    string        // default is ".keys"
}

// Opts may be empty
func NewKms(storage storage.Storage, opts *Opts) *Kms {
	if opts == nil {
		opts = &Opts{}
	}
	if opts.RotationFreq < 5*time.Minute {
		opts.RotationFreq = 30 * 24 * time.Hour // 30 day default
	}
	if opts.MaxAge < 5*time.Minute {
		opts.MaxAge = 90 * 24 * time.Hour // 90 day default
	}
	if opts.KeyTimeFormat == "" {
		opts.KeyTimeFormat = "20060102_1504" // default key format
	}
	if opts.KeysFolder == "" {
		opts.KeysFolder = ".keys"
	}
	return &Kms{
		storage, &sync.Map{}, len(opts.KeyTimeFormat),
		opts.RotationFreq, opts.MaxAge, opts.KeyTimeFormat, opts.KeysFolder,
	}
}

// Determines the current key id to use for hashing/encrypting data
func (k *Kms) currentKeyId(ctx context.Context) *string {
	_, now := timex.Now(ctx)
	t := now.Truncate(k.rotationFreq)
	id := t.Format(k.keyTimeFormat)
	return &id
}

// Determines if the provided key is valid and not expired.
func (k *Kms) validateKeyId(ctx context.Context, id string) error {
	_, now := timex.Now(ctx)
	t, err := time.Parse(k.keyTimeFormat, id)
	if err != nil {
		return fmt.Errorf("invalid key id format: %s, expected format: %s", id, k.keyTimeFormat)
	}
	if t.After(now) {
		return fmt.Errorf("key id %s is in the future", id)
	}
	if t.Add(k.maxAge).Before(now) {
		return fmt.Errorf("key id %s is expired, max age is %s", id, k.maxAge)
	}
	return nil
}

// Generates a new key 96 byte key: 32 for encryption, 64 for HMAC"
func generateKey() []byte {
	key := make([]byte, 96)
	if _, err := rand.Read(key); err != nil {
		panic(fmt.Errorf("kms.generateKey(): rand.Read returned %w", err))
	}
	return key
}

// Returns the storage path based on the provided key id
func (k *Kms) storagePath(keyId string) string {
	return path.Join(k.keysFolder, keyId)
}

// Writes a newly generated key to storage for the specified storage id (not key
// id). Will first attempt write, but if fails due to already existing key, it
// will simply read and return the existing key. Does NOT consult the keys
// sync.Map cache.
func (k *Kms) writeOrReadExisting(ctx context.Context, blobId *string, data []byte) []byte {
	if written := k.storage.WriteIfMissing(ctx, *blobId, data); !written {
		var found bool
		if data, found = k.storage.Read(ctx, *blobId); !found {
			panic(fmt.Errorf("kms.Kms.writeOrReadExisting(): failed to write key %s, and could not read existing", *blobId))
		}
	}
	return data
}

// Returns the key with the specified id if it exists
func (k *Kms) key(ctx context.Context, id *string, createIfMissing bool) ([]byte, error) {
	if !createIfMissing {
		if err := k.validateKeyId(ctx, *id); err != nil {
			return nil, err
		}
	}
	var key []byte
	val, ok := k.keys.Load(*id)
	if !ok {
		storagePath := k.storagePath(*id)
		if createIfMissing {
			key = generateKey()
			key = k.writeOrReadExisting(ctx, &storagePath, key)
		} else {
			var found bool
			if key, found = k.storage.Read(ctx, storagePath); !found {
				return nil, fmt.Errorf("key not found: %s", *id)
			}
		}
		k.keys.Store(*id, key)
	} else {
		if createIfMissing {
			if err := k.validateKeyId(ctx, *id); err != nil {
				return nil, err
			}
		}
		key = val.([]byte)
	}
	return key, nil
}

// Uses first 32 bytes of the current key, as aes requires a 32-byte key for
// AES-256.
func (k *Kms) encryptionKey(ctx context.Context, id *string, createIfMissing bool) ([]byte, error) {
	key, err := k.key(ctx, id, createIfMissing)
	if err != nil {
		return nil, err
	}
	return key[:32], nil
}

// Uses the last 64 bytes of the current key, as HMAC SHA512 requires a 64-byte
// key.
func (k *Kms) hashKey(ctx context.Context, keyId *string, createIfMissing bool) ([]byte, error) {
	key, err := k.key(ctx, keyId, createIfMissing)
	if err != nil {
		return nil, err
	}
	return key[32:], nil
}

// Use AES-GCM encryption to encrypt data with key used for encryption added as
// first bytes.
func (k *Kms) Encrypt(ctx context.Context, data []byte) []byte {
	// get key
	keyId := k.currentKeyId(ctx)
	key, err := k.encryptionKey(ctx, keyId, true)
	if err != nil {
		panic(fmt.Errorf("kms.Kms.Encrypt(): getting encryption key: %w", err))
	}

	// generate ciphertext
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Errorf("creating cipher: %w", err))
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(fmt.Errorf("creating GCM: %w", err))
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		panic(fmt.Errorf("generating nonce: %w", err))
	}
	ciphertext := aead.Seal(nil, nonce, data, nil)
	ciphertext = append(nonce, ciphertext...)

	// Prepend the key id and nonce to the ciphertext
	ciphertext = append([]byte(*keyId), ciphertext...)
	return ciphertext
}

// Prepends key id and SHA512 HMAC to data to sign it.
func (k *Kms) Hash(ctx context.Context, data []byte) []byte {
	// get key
	keyId := k.currentKeyId(ctx)
	key, err := k.hashKey(ctx, keyId, true)
	if err != nil {
		panic(fmt.Errorf("kms.Kms.Hash(): getting hash key: %w", err))
	}

	// generate hashed data
	mac := hmac.New(sha512.New, key)
	mac.Write(data)
	result := mac.Sum(nil)
	result = append(result, data...)

	// prepend the key id to the cipher text
	result = append([]byte(*keyId), result...)
	return result
}

func (k *Kms) B64Encrypt(ctx context.Context, data []byte) []byte {
	encr := k.Encrypt(ctx, data)
	return b64.UrlEncode(encr)
}

func (k *Kms) B64Hash(ctx context.Context, data []byte) []byte {
	hashed := k.Hash(ctx, data)
	return b64.UrlEncode(hashed)
}

func (k *Kms) B64EncryptJson(ctx context.Context, payload any) []byte {
	jsonBytes := jsonx.Marshal(payload)
	return k.B64Encrypt(ctx, jsonBytes)
}

func (k *Kms) B64HashJson(ctx context.Context, payload any) []byte {
	jsonBytes := jsonx.Marshal(payload)
	return k.B64Hash(ctx, jsonBytes)
}

// Use AES-GCM decryption to decrypt data with key used for encryption extracted
// from the first bytes.
func (k *Kms) Decrypt(ctx context.Context, data []byte) ([]byte, error) {
	// get key
	keyId := string(data[:k.keyLength])
	key, err := k.encryptionKey(ctx, &keyId, false)
	if err != nil {
		return nil, err
	}

	// decrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Errorf("creating cipher: %w", err))
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(fmt.Errorf("creating GCM: %w", err))
	}
	nonceSize := aead.NonceSize()
	nonce := data[k.keyLength : k.keyLength+nonceSize]
	ciphertext := data[k.keyLength+nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting: %w", err)
	}
	return plaintext, nil
}

// Verifies the HMAC SHA512 signature of the data and returns original payload.
func (k *Kms) Unhash(ctx context.Context, data []byte) ([]byte, error) {
	keyId := string(data[:k.keyLength])
	macStart := k.keyLength
	macEnd := macStart + sha512.Size
	macBytes := data[macStart:macEnd]
	payload := data[macEnd:]

	key, err := k.hashKey(ctx, &keyId, false)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha512.New, key)
	mac.Write(payload)
	expectedMac := mac.Sum(nil)
	if !hmac.Equal(macBytes, expectedMac) {
		return nil, fmt.Errorf("invalid HMAC signature for key id %s", keyId)
	}
	return payload, nil
}

// AES-GCM decrypts base64-encoded data.
func (k *Kms) DecryptB64(ctx context.Context, data []byte) ([]byte, error) {
	dec, err := b64.UrlDecode(data)
	if err != nil {
		return nil, err
	}
	return k.Decrypt(ctx, dec)
}

// Verifies the HMAC SHA512 signature of the base64-encoded data and returns
// original payload.
func (k *Kms) UnhashB64(ctx context.Context, data []byte) ([]byte, error) {
	dec, err := b64.UrlDecode(data)
	if err != nil {
		return nil, err
	}
	return k.Unhash(ctx, dec)
}

// AES-GCM decrypts base64-encoded, JSON data and unmarshals it into the provided
// payload.
func (k *Kms) JsonDecryptB64(ctx context.Context, data []byte, payload any) error {
	payloadBytes, err := k.DecryptB64(ctx, data)
	if err != nil {
		return err
	}
	return json.Unmarshal(payloadBytes, payload)
}

// Verifies the HMAC SHA512 signature of base64-encoded, JSON data and
// unmarshals it into the provided payload.
func (k *Kms) JsonUnhashB64(ctx context.Context, data []byte, payload any) error {
	payloadBytes, err := k.UnhashB64(ctx, data)
	if err != nil {
		return err
	}
	return json.Unmarshal(payloadBytes, payload)
}
