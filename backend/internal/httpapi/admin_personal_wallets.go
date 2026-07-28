package httpapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/store"
)

type updateAdminPersonalWalletRequest struct {
	WalletAddress string `json:"walletAddress"`
}

func (s *Server) handleAdminPersonalWallets(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}

	wallets, err := s.store.AdminPersonalWallets(r.Context())
	if err != nil {
		s.logger.Error("read admin personal wallets", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read personal wallets")
		return
	}
	writeJSON(w, http.StatusOK, wallets)
}

func (s *Server) handleUpdateAdminPersonalWallet(w http.ResponseWriter, r *http.Request) {
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}

	var req updateAdminPersonalWalletRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	wallet, err := s.store.AdminUpdatePersonalWalletAddress(r.Context(), store.UpdateAdminPersonalWalletParams{
		Code:          r.PathValue("code"),
		WalletAddress: req.WalletAddress,
		AdminID:       admin.ID,
		Now:           time.Now().UTC(),
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidPersonalWalletCode):
			writeError(w, http.StatusNotFound, "personal wallet not found")
		case errors.Is(err, store.ErrInvalidPersonalWalletAddress):
			writeError(w, http.StatusBadRequest, "wallet address must be a valid 0x EVM/BEP-20 address")
		default:
			s.logger.Error("update admin personal wallet", "error", err, "code", r.PathValue("code"))
			writeError(w, http.StatusInternalServerError, "failed to update personal wallet")
		}
		return
	}

	writeJSON(w, http.StatusOK, wallet)
}
