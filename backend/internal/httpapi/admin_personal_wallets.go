package httpapi

import (
	"errors"
	"net/http"
	"strings"
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
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}

	wallets, err := s.store.AdminPersonalWallets(r.Context())
	if err != nil {
		s.logger.Error("read admin personal wallets", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read personal wallets")
		return
	}
	for index := range wallets {
		wallets[index].CanManage = s.adminCanManagePersonalWallet(admin, wallets[index].Code)
	}
	writeJSON(w, http.StatusOK, wallets)
}

func (s *Server) handleUpdateAdminPersonalWallet(w http.ResponseWriter, r *http.Request) {
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	if !s.adminCanManagePersonalWallet(admin, r.PathValue("code")) {
		writeError(w, http.StatusForbidden, "admin can only manage their assigned personal wallet")
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

	wallet.CanManage = true
	writeJSON(w, http.StatusOK, wallet)
}

func (s *Server) handleCreateAdminPersonalWalletWithdrawal(w http.ResponseWriter, r *http.Request) {
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	if !s.adminCanManagePersonalWallet(admin, r.PathValue("code")) {
		writeError(w, http.StatusForbidden, "admin can only manage their assigned personal wallet")
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

	if s.gasFeeWithdrawer != nil {
		broadcast, err := s.gasFeeWithdrawer.SendUSDTTransfer(r.Context(), withdrawal.DestinationAddress, withdrawal.Amount)
		if err != nil {
			failed, markErr := s.store.AdminMarkPersonalWalletWithdrawalFailed(r.Context(), store.UpdateAdminPersonalWalletWithdrawalStatusParams{
				WithdrawalID: withdrawal.ID,
				AdminID:      admin.ID,
				Reason:       err.Error(),
				Now:          time.Now().UTC(),
			})
			if markErr != nil {
				s.logger.Error("mark personal wallet withdrawal failed", "error", markErr, "withdrawal_id", withdrawal.ID)
			} else {
				withdrawal = failed
			}
			s.logger.Error("broadcast personal wallet withdrawal", "error", err, "withdrawal_id", withdrawal.ID, "wallet_code", withdrawal.WalletCode)
			writeError(w, http.StatusBadGateway, "withdrawal saved but on-chain broadcast failed")
			return
		}

		withdrawal, err = s.store.AdminMarkPersonalWalletWithdrawalBroadcast(r.Context(), store.UpdateAdminPersonalWalletWithdrawalStatusParams{
			WithdrawalID: withdrawal.ID,
			AdminID:      admin.ID,
			TxID:         broadcast.TxHash,
			Now:          time.Now().UTC(),
		})
		if err != nil {
			s.logger.Error("mark personal wallet withdrawal broadcast", "error", err, "withdrawal_id", withdrawal.ID, "tx_id", broadcast.TxHash)
			writeError(w, http.StatusInternalServerError, "withdrawal broadcasted but failed to save transaction hash")
			return
		}
	}

	writeJSON(w, http.StatusAccepted, withdrawal)
}

func (s *Server) adminCanManagePersonalWallet(admin store.AdminUserView, code string) bool {
	normalizedCode := normalizePersonalWalletRouteCode(code)
	if normalizedCode == "" {
		return true
	}
	if !s.isAssignedToAryantoPersonalWallet(admin) {
		return true
	}
	return normalizedCode == "aryanto_hong"
}

func (s *Server) isAssignedToAryantoPersonalWallet(admin store.AdminUserView) bool {
	adminEmail := normalizeAdminIdentity(admin.Email)
	adminName := normalizeAdminIdentity(admin.DisplayName)
	configEmail := normalizeAdminIdentity(s.config.AdminTwoEmail)
	configName := normalizeAdminIdentity(s.config.AdminTwoName)

	return (configEmail != "" && adminEmail == configEmail) ||
		(configName != "" && adminName == configName) ||
		adminEmail == "admin@mautrade.com" ||
		adminName == "aryanto hong"
}

func normalizePersonalWalletRouteCode(value string) string {
	code := strings.ToLower(strings.TrimSpace(value))
	code = strings.ReplaceAll(code, "-", "_")
	switch code {
	case "viksry", "aryanto_hong":
		return code
	default:
		return ""
	}
}

func normalizeAdminIdentity(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}
