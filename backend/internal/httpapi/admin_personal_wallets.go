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

type createAdminPersonalWalletWithdrawalRequest struct {
	Amount             string `json:"amount"`
	WalletAddress      string `json:"walletAddress"`
	DestinationAddress string `json:"destinationAddress"`
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

func (s *Server) handleCreateAdminPersonalWalletWithdrawal(w http.ResponseWriter, r *http.Request) {
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}

	var req createAdminPersonalWalletWithdrawalRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	destinationAddress := req.DestinationAddress
	if destinationAddress == "" {
		destinationAddress = req.WalletAddress
	}

	withdrawal, err := s.store.AdminCreatePersonalWalletWithdrawal(r.Context(), store.CreateAdminPersonalWalletWithdrawalParams{
		Code:          r.PathValue("code"),
		WalletAddress: destinationAddress,
		Amount:        req.Amount,
		AdminID:       admin.ID,
		Now:           time.Now().UTC(),
	})
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidPersonalWalletCode):
			writeError(w, http.StatusNotFound, "personal wallet not found")
		case errors.Is(err, store.ErrInvalidPersonalWalletAddress):
			writeError(w, http.StatusBadRequest, "wallet address must be a valid 0x EVM/BEP-20 address")
		case errors.Is(err, store.ErrPersonalWalletWithdrawAddressRequired):
			writeError(w, http.StatusBadRequest, "wallet address is required before withdrawing")
		case errors.Is(err, store.ErrInvalidPersonalWalletWithdrawAmount):
			writeError(w, http.StatusBadRequest, "withdraw amount must be greater than 0 and use up to 18 decimals")
		case errors.Is(err, store.ErrPersonalWalletWithdrawInsufficient):
			writeError(w, http.StatusConflict, "withdraw amount exceeds available wallet balance")
		default:
			s.logger.Error("create admin personal wallet withdrawal", "error", err, "code", r.PathValue("code"))
			writeError(w, http.StatusInternalServerError, "failed to create personal wallet withdrawal")
		}
		return
	}

	writeJSON(w, http.StatusAccepted, withdrawal)
}
