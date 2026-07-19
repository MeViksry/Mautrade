package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

const envelopeVersion = "v1:"

type Encryptor struct {
	aead cipher.AEAD
}

func NewEncryptor(rawKey, environment string) (*Encryptor, error) {
	key, err := normalizeKey(rawKey, environment)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("secrets: create aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("secrets: create aes-gcm: %w", err)
	}
	return &Encryptor{aead: aead}, nil
}

func (e *Encryptor) SealString(value string) ([]byte, error) {
	if e == nil {
		return nil, fmt.Errorf("secrets: encryptor is nil")
	}
	nonce := make([]byte, e.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("secrets: generate nonce: %w", err)
	}
	ciphertext := e.aead.Seal(nil, nonce, []byte(value), nil)
	payload := append(nonce, ciphertext...)
	encoded := envelopeVersion + base64.RawStdEncoding.EncodeToString(payload)
	return []byte(encoded), nil
}

func (e *Encryptor) OpenString(ciphertext []byte) (string, error) {
	if e == nil {
		return "", fmt.Errorf("secrets: encryptor is nil")
	}
	encoded := strings.TrimSpace(string(ciphertext))
	if !strings.HasPrefix(encoded, envelopeVersion) {
		return "", fmt.Errorf("secrets: unsupported ciphertext envelope")
	}
	payload, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(encoded, envelopeVersion))
	if err != nil {
		return "", fmt.Errorf("secrets: decode ciphertext: %w", err)
	}
	nonceSize := e.aead.NonceSize()
	if len(payload) <= nonceSize {
		return "", fmt.Errorf("secrets: ciphertext too short")
	}
	plaintext, err := e.aead.Open(nil, payload[:nonceSize], payload[nonceSize:], nil)
	if err != nil {
		return "", fmt.Errorf("secrets: decrypt ciphertext: %w", err)
	}
	return string(plaintext), nil
}

func Mask(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}
	return value[:4] + strings.Repeat("*", 8) + value[len(value)-4:]
}

func normalizeKey(rawKey, environment string) ([]byte, error) {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		if strings.EqualFold(environment, "production") {
			return nil, fmt.Errorf("secrets: EXCHANGE_CREDENTIAL_KEY is required in production")
		}
		sum := sha256.Sum256([]byte("mautrade-local-development-exchange-credential-key"))
		return sum[:], nil
	}
	if strings.HasPrefix(rawKey, "base64:") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(rawKey, "base64:"))
		if err != nil {
			return nil, fmt.Errorf("secrets: decode base64 key: %w", err)
		}
		if len(decoded) != 32 {
			return nil, fmt.Errorf("secrets: base64 key must decode to 32 bytes")
		}
		return decoded, nil
	}
	if strings.HasPrefix(rawKey, "hex:") {
		decoded, err := hex.DecodeString(strings.TrimPrefix(rawKey, "hex:"))
		if err != nil {
			return nil, fmt.Errorf("secrets: decode hex key: %w", err)
		}
		if len(decoded) != 32 {
			return nil, fmt.Errorf("secrets: hex key must decode to 32 bytes")
		}
		return decoded, nil
	}
	if len(rawKey) == 32 {
		return []byte(rawKey), nil
	}
	sum := sha256.Sum256([]byte(rawKey))
	return sum[:], nil
}
