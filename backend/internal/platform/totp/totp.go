package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultDigits = 6
	defaultPeriod = 30 * time.Second
	secretBytes   = 20
)

var base32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

func GenerateSecret() (string, error) {
	raw := make([]byte, secretBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("totp: generate secret: %w", err)
	}
	return base32NoPadding.EncodeToString(raw), nil
}

func BuildURI(issuer, account, secret string) string {
	issuer = strings.TrimSpace(issuer)
	account = strings.TrimSpace(account)
	secret = normalizeSecret(secret)
	label := account
	if issuer != "" {
		label = issuer + ":" + account
	}
	values := url.Values{}
	values.Set("secret", secret)
	if issuer != "" {
		values.Set("issuer", issuer)
	}
	values.Set("algorithm", "SHA1")
	values.Set("digits", strconv.Itoa(defaultDigits))
	values.Set("period", strconv.Itoa(int(defaultPeriod.Seconds())))
	return "otpauth://totp/" + url.PathEscape(label) + "?" + values.Encode()
}

func Validate(code, secret string, now time.Time) bool {
	code = normalizeCode(code)
	if len(code) != defaultDigits {
		return false
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	for offset := -1; offset <= 1; offset++ {
		expected, err := GenerateCode(secret, now.Add(time.Duration(offset)*defaultPeriod))
		if err != nil {
			return false
		}
		if subtle.ConstantTimeCompare([]byte(code), []byte(expected)) == 1 {
			return true
		}
	}
	return false
}

func GenerateCode(secret string, when time.Time) (string, error) {
	return generateCode(secret, when, defaultDigits)
}

func generateCode(secret string, when time.Time, digits int) (string, error) {
	if digits <= 0 || digits > 8 {
		return "", fmt.Errorf("totp: digits must be between 1 and 8")
	}
	key, err := decodeSecret(secret)
	if err != nil {
		return "", err
	}
	counter := uint64(when.UTC().Unix() / int64(defaultPeriod.Seconds()))
	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], counter)

	mac := hmac.New(sha1.New, key)
	if _, err := mac.Write(counterBytes[:]); err != nil {
		return "", fmt.Errorf("totp: write hmac: %w", err)
	}
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	binaryCode := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		(uint32(sum[offset+3]) & 0xff)

	modulo := uint32(1)
	for range digits {
		modulo *= 10
	}
	return fmt.Sprintf("%0*d", digits, binaryCode%modulo), nil
}

func decodeSecret(secret string) ([]byte, error) {
	normalized := normalizeSecret(secret)
	if normalized == "" {
		return nil, fmt.Errorf("totp: secret is required")
	}
	decoded, err := base32NoPadding.DecodeString(normalized)
	if err == nil {
		return decoded, nil
	}
	decoded, paddedErr := base32.StdEncoding.DecodeString(addPadding(normalized))
	if paddedErr != nil {
		return nil, fmt.Errorf("totp: decode secret: %w", err)
	}
	return decoded, nil
}

func normalizeSecret(secret string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secret), " ", ""))
}

func normalizeCode(code string) string {
	code = strings.TrimSpace(code)
	code = strings.ReplaceAll(code, " ", "")
	code = strings.ReplaceAll(code, "-", "")
	return code
}

func addPadding(value string) string {
	if remainder := len(value) % 8; remainder != 0 {
		value += strings.Repeat("=", 8-remainder)
	}
	return value
}
