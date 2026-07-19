package totp

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateCodeRFC6238SHA1Vectors(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	tests := []struct {
		unix int64
		want string
	}{
		{unix: 59, want: "94287082"},
		{unix: 1111111109, want: "07081804"},
		{unix: 1111111111, want: "14050471"},
		{unix: 1234567890, want: "89005924"},
		{unix: 2000000000, want: "69279037"},
		{unix: 20000000000, want: "65353130"},
	}

	for _, tt := range tests {
		got, err := generateCode(secret, time.Unix(tt.unix, 0).UTC(), 8)
		if err != nil {
			t.Fatalf("generateCode(%d): %v", tt.unix, err)
		}
		if got != tt.want {
			t.Fatalf("generateCode(%d) = %s, want %s", tt.unix, got, tt.want)
		}
	}
}

func TestValidateAcceptsCurrentAndAdjacentWindow(t *testing.T) {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	now := time.Unix(2000000000, 0).UTC()
	code, err := GenerateCode(secret, now)
	if err != nil {
		t.Fatalf("GenerateCode: %v", err)
	}

	if !Validate(code, secret, now) {
		t.Fatal("Validate rejected current window code")
	}
	if !Validate(code, secret, now.Add(30*time.Second)) {
		t.Fatal("Validate rejected adjacent window code")
	}
	if Validate(code, secret, now.Add(90*time.Second)) {
		t.Fatal("Validate accepted expired code outside drift window")
	}
}

func TestGenerateSecretAndBuildURI(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret: %v", err)
	}
	if len(secret) < 16 {
		t.Fatalf("secret length = %d, want at least 16", len(secret))
	}
	if strings.Contains(secret, "=") {
		t.Fatalf("secret should not contain padding: %q", secret)
	}

	uri := BuildURI("Mautrade", "admin@example.com", secret)
	for _, part := range []string{"otpauth://totp/", "secret=" + secret, "issuer=Mautrade", "digits=6", "period=30"} {
		if !strings.Contains(uri, part) {
			t.Fatalf("uri %q missing %q", uri, part)
		}
	}
}
