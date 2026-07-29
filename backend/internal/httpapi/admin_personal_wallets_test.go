package httpapi

import (
	"testing"

	"github.com/MeViksry/Mautrade/backend/internal/config"
	"github.com/MeViksry/Mautrade/backend/internal/store"
)

func TestAdminCanManagePersonalWallet(t *testing.T) {
	t.Parallel()

	server := &Server{config: config.Config{
		AdminTwoEmail: "admin@mautrade.com",
		AdminTwoName:  "ARYANTO HONG",
	}}

	vikri := store.AdminUserView{
		Email:       "kingvikvik25@gmail.com",
		DisplayName: "VIKRI AHPAD TANTOWI",
	}
	aryanto := store.AdminUserView{
		Email:       "admin@mautrade.com",
		DisplayName: "ARYANTO HONG",
	}

	tests := []struct {
		name  string
		admin store.AdminUserView
		code  string
		want  bool
	}{
		{name: "vikri can manage viksry wallet", admin: vikri, code: "viksry", want: true},
		{name: "vikri can manage aryanto wallet", admin: vikri, code: "aryanto_hong", want: true},
		{name: "aryanto cannot manage viksry wallet", admin: aryanto, code: "viksry", want: false},
		{name: "aryanto can manage own wallet", admin: aryanto, code: "aryanto-hong", want: true},
		{name: "unknown wallet code falls through to store validation", admin: aryanto, code: "unknown", want: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := server.adminCanManagePersonalWallet(tt.admin, tt.code); got != tt.want {
				t.Fatalf("adminCanManagePersonalWallet() = %v, want %v", got, tt.want)
			}
		})
	}
}
