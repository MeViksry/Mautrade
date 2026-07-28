package httpapi

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestValidateUpdateProfileRequest(t *testing.T) {
	t.Parallel()

	req := updateProfileRequest{
		DisplayName:         "Meviksry",
		Timezone:            "Asia/Jakarta",
		ProfilePhotoDataURL: "data:image/png;base64,aGVsbG8=",
	}
	if err := validateUpdateProfileRequest(req); err != nil {
		t.Fatalf("expected valid profile request, got %v", err)
	}
}

func TestValidateUpdateProfileRequestRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tooLargePhoto := "data:image/png;base64," + base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, maxProfilePhotoBytes+1))
	tests := []struct {
		name string
		req  updateProfileRequest
	}{
		{
			name: "short display name",
			req: updateProfileRequest{
				DisplayName: "M",
				Timezone:    "Asia/Jakarta",
			},
		},
		{
			name: "invalid timezone",
			req: updateProfileRequest{
				DisplayName: "Meviksry",
				Timezone:    "Asia/Not_A_Zone",
			},
		},
		{
			name: "svg rejected",
			req: updateProfileRequest{
				DisplayName:         "Meviksry",
				Timezone:            "Asia/Jakarta",
				ProfilePhotoDataURL: "data:image/svg+xml;base64,PHN2Zy8+",
			},
		},
		{
			name: "malformed base64",
			req: updateProfileRequest{
				DisplayName:         "Meviksry",
				Timezone:            "Asia/Jakarta",
				ProfilePhotoDataURL: "data:image/png;base64,not-base64",
			},
		},
		{
			name: "too large",
			req: updateProfileRequest{
				DisplayName:         "Meviksry",
				Timezone:            "Asia/Jakarta",
				ProfilePhotoDataURL: tooLargePhoto,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateUpdateProfileRequest(tt.req); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
