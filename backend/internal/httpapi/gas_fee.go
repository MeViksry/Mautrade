package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/domain/id"
	"github.com/MeViksry/Mautrade/backend/internal/store"
)

type createGasFeeDepositRequest struct {
	Amount string `json:"amount"`
	Asset  string `json:"asset"`
	TxID   string `json:"tx_id"`
	TxId   string `json:"txId"`
}

type updateGasFeeDepositStatusRequest struct {
	AdminID        string `json:"admin_id"`
	ActorID        string `json:"actor_id"`
	Status         string `json:"status"`
	ResolutionNote string `json:"resolution_note"`
}

func (s *Server) handleUserGasFeeAccount(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to read gas fee account")
		return
	}
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	account, err := s.store.UserGasFeeAccount(
		r.Context(),
		user.ID,
		firstNonEmpty(r.URL.Query().Get("asset"), s.config.DefaultCurrency),
		positiveIntQuery(r, "limit", 100, 200),
	)
	if err != nil {
		s.logger.Error("read user gas fee account", "user_id", user.ID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read gas fee account")
		return
	}
	writeJSON(w, http.StatusOK, account)
}

func (s *Server) handleCreateGasFeeDeposit(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to create gas fee deposit")
		return
	}
	user, err := s.authUserFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid or expired session")
		return
	}

	var req createGasFeeDepositRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	deposit, err := s.store.CreateGasFeeDeposit(r.Context(), store.CreateGasFeeDepositParams{
		UserID:         user.ID,
		Amount:         req.Amount,
		Asset:          firstNonEmpty(req.Asset, s.config.DefaultCurrency),
		DepositAddress: s.config.GasFeeDepositAddress,
		TxID:           firstNonEmpty(req.TxID, req.TxId),
		Now:            time.Now().UTC(),
	})
	if err != nil {
		writeGasFeeDepositError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, deposit)
}

func (s *Server) handleAdminGasFeeDeposits(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to read gas fee deposits")
		return
	}
	if _, ok := s.requireAdmin(w, r); !ok {
		return
	}
	status, err := normalizeGasFeeDepositStatus(r.URL.Query().Get("status"), true)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	deposits, err := s.store.AdminGasFeeDeposits(r.Context(), store.AdminGasFeeDepositsParams{
		Status: status,
		Limit:  positiveIntQuery(r, "limit", 100, 500),
		Offset: nonNegativeIntQuery(r, "offset", 0),
	})
	if err != nil {
		s.logger.Error("read admin gas fee deposits", "status", status, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read gas fee deposits")
		return
	}
	writeJSON(w, http.StatusOK, deposits)
}

func (s *Server) handleUpdateGasFeeDepositStatus(w http.ResponseWriter, r *http.Request) {
	if !s.store.Ready() {
		writeError(w, http.StatusServiceUnavailable, "postgres is required to update gas fee deposit")
		return
	}
	admin, ok := s.requireAdmin(w, r)
	if !ok {
		return
	}
	depositID := strings.TrimSpace(r.PathValue("deposit_id"))
	if _, err := id.Parse(depositID); err != nil {
		writeError(w, http.StatusBadRequest, "deposit_id must be a canonical UUID")
		return
	}

	var req updateGasFeeDepositStatusRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	status, err := normalizeGasFeeDepositStatus(req.Status, false)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	adminID := strings.TrimSpace(firstNonEmpty(req.AdminID, req.ActorID))
	if !adminBodyMustMatchSession(w, adminID, admin) {
		return
	}

	deposit, err := s.store.UpdateGasFeeDepositStatus(r.Context(), store.UpdateGasFeeDepositStatusParams{
		DepositID:      depositID,
		AdminID:        admin.ID,
		Status:         status,
		ResolutionNote: req.ResolutionNote,
		Now:            time.Now().UTC(),
	})
	if err != nil {
		writeGasFeeDepositError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, deposit)
}

func normalizeGasFeeDepositStatus(value string, allowPending bool) (string, error) {
	status := strings.ToLower(strings.TrimSpace(value))
	if status == "" {
		if allowPending {
			return "pending", nil
		}
		return "", errors.New("status is required")
	}
	if allowPending && status == "all" {
		return "", nil
	}
	switch status {
	case "pending":
		if allowPending {
			return status, nil
		}
	case "confirmed", "rejected":
		return status, nil
	}
	if allowPending {
		return "", errors.New("status must be pending, confirmed, rejected, or all")
	}
	return "", errors.New("status must be confirmed or rejected")
}

func writeGasFeeDepositError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrGasFeeDepositAmount):
		writeError(w, http.StatusBadRequest, "amount must be at least 500 USDT")
	case errors.Is(err, store.ErrGasFeeDepositTxIDRequired):
		writeError(w, http.StatusBadRequest, "tx_id is required")
	case errors.Is(err, store.ErrGasFeeDepositStatus):
		writeError(w, http.StatusBadRequest, "invalid gas fee deposit status")
	case errors.Is(err, store.ErrGasFeeDepositTransition):
		writeError(w, http.StatusConflict, "gas fee deposit is no longer pending")
	case errors.Is(err, store.ErrGasFeeDepositNotFound):
		writeError(w, http.StatusNotFound, "gas fee deposit not found")
	default:
		writeError(w, http.StatusInternalServerError, "gas fee deposit operation failed")
	}
}
