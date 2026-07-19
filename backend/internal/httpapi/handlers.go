package httpapi

import (
	"net/http"
	"time"

	"github.com/MeViksry/qdecimal"
)

func (s *Server) handleUserStats(w http.ResponseWriter, r *http.Request) {
	if s.store.Ready() {
		user, err := s.authUserFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		stats, err := s.store.UserStats(r.Context(), user.ID, s.config.DefaultCurrency, s.config.GasFeeShareRate)
		if err != nil {
			s.logger.Error("read user stats", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to read user stats")
			return
		}
		writeJSON(w, http.StatusOK, stats)
		return
	}

	writeJSON(w, http.StatusOK, mockUserStats(s.config.DefaultCurrency, s.config.GasFeeShareRate))
}

func mockUserStats(defaultCurrency, gasFeeShareRate string) map[string]any {
	return map[string]any{
		"totalBalance":      "12450.75",
		"realizedProfit":    "3240.50",
		"totalGasFeePaid":   "1620.25",
		"activeLayersCount": 18,
		"precisionPolicy":   "qdecimal",
		"defaultCurrency":   defaultCurrency,
		"gasFeeShareRate":   gasFeeShareRate,
	}
}

func (s *Server) handleExchangeBindings(w http.ResponseWriter, r *http.Request) {
	if s.store.Ready() {
		user, err := s.authUserFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		bindings, err := s.store.UserExchangeBindings(r.Context(), user.ID, s.config.DefaultCurrency)
		if err != nil {
			s.logger.Error("read exchange bindings", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to read exchange bindings")
			return
		}
		writeJSON(w, http.StatusOK, bindings)
		return
	}

	writeJSON(w, http.StatusOK, mockExchangeBindings())
}

func mockExchangeBindings() []map[string]any {
	return []map[string]any{
		{"id": "binance-binding", "name": "Binance", "status": "connected", "lastSynced": "2026-07-18T10:30:00Z", "balance": "8450.75", "hasApi": true},
		{"id": "okx-binding", "name": "OKX", "status": "connected", "lastSynced": "2026-07-18T10:30:00Z", "balance": "4000.00", "hasApi": true},
		{"id": "bybit-binding", "name": "Bybit", "status": "disconnected", "lastSynced": "2026-07-01T08:00:00Z", "balance": "0.00", "hasApi": true},
		{"id": "tokocrypto-binding", "name": "Tokocrypto", "status": "disconnected", "lastSynced": nil, "balance": "0.00", "hasApi": true},
	}
}

func (s *Server) handleLayers(w http.ResponseWriter, r *http.Request) {
	if s.store.Ready() {
		user, err := s.authUserFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		layers, err := s.store.ActiveLayers(r.Context(), user.ID)
		if err != nil {
			s.logger.Error("read active layers", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to read active layers")
			return
		}
		writeJSON(w, http.StatusOK, layers)
		return
	}

	writeJSON(w, http.StatusOK, mockLayers())
}

func mockLayers() []map[string]any {
	return []map[string]any{
		layer("L-101", "BTC/USDT", "62450.00", "63100.50", "10", "845.07", "8.79", "1.04", "open", "2026-07-18T08:15:00Z"),
		layer("L-102", "ETH/USDT", "3450.25", "3410.00", "5", "422.53", "-4.93", "-1.16", "open", "2026-07-18T09:00:00Z"),
		layer("L-103", "SOL/USDT", "145.50", "151.20", "15", "1267.61", "49.65", "3.91", "open", "2026-07-17T14:20:00Z"),
	}
}

func (s *Server) handleTradeHistory(w http.ResponseWriter, r *http.Request) {
	if s.store.Ready() {
		user, err := s.authUserFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}
		history, err := s.store.TradeHistory(r.Context(), user.ID)
		if err != nil {
			s.logger.Error("read trade history", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to read trade history")
			return
		}
		writeJSON(w, http.StatusOK, history)
		return
	}

	writeJSON(w, http.StatusOK, mockTradeHistory())
}

func mockTradeHistory() []map[string]any {
	return []map[string]any{
		{"id": "L-099", "pair": "BNB/USDT", "exitPrice": "580.00", "pnl": "120.50", "gasFee": "60.25", "closedAt": "2026-07-16T11:00:00Z"},
		{"id": "L-098", "pair": "ADA/USDT", "exitPrice": "0.42", "pnl": "-20.00", "gasFee": "-10.00", "closedAt": "2026-07-15T09:30:00Z"},
	}
}

func (s *Server) handleAdminOverview(w http.ResponseWriter, r *http.Request) {
	if s.store.Ready() {
		if _, ok := s.requireAdmin(w, r); !ok {
			return
		}
		overview, err := s.store.AdminOverview(r.Context(), s.config.DefaultCurrency)
		if err != nil {
			s.logger.Error("read admin overview", "error", err)
			writeError(w, http.StatusInternalServerError, "failed to read admin overview")
			return
		}
		writeJSON(w, http.StatusOK, overview)
		return
	}

	writeJSON(w, http.StatusOK, mockAdminOverview())
}

func mockAdminOverview() map[string]any {
	return map[string]any{
		"registeredUsers":     12840,
		"activeUsers":         8420,
		"openLayers":          18420,
		"estimatedAUM":        "8450192.250000000000000000",
		"gasFeeRevenueToday":  "4820.500000000000000000",
		"orphanedLayers":      12,
		"executionQueueState": "ready",
	}
}

type gasFeePreviewRequest struct {
	EntryValue string `json:"entry_value"`
	ExitValue  string `json:"exit_value"`
}

func (s *Server) handleGasFeePreview(w http.ResponseWriter, r *http.Request) {
	if s.store.Ready() {
		if _, ok := s.requireAdmin(w, r); !ok {
			return
		}
	}
	var req gasFeePreviewRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	entry, err := qdecimal.Parse(req.EntryValue)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid entry_value"})
		return
	}
	exit, err := qdecimal.Parse(req.ExitValue)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid exit_value"})
		return
	}
	writeJSON(w, http.StatusOK, s.gasFeeCalc.CalculateFromValues(entry, exit))
}

func layer(id, pair, entry, current, allocation, allocated, pnl, pnlPct, status, openedAt string) map[string]any {
	return map[string]any{
		"id":               id,
		"pair":             pair,
		"entryPrice":       entry,
		"currentPrice":     current,
		"allocationPct":    allocation,
		"allocatedUsdt":    allocated,
		"unrealizedPnl":    pnl,
		"unrealizedPnlPct": pnlPct,
		"status":           status,
		"openedAt":         openedAt,
		"serverTime":       time.Now().UTC(),
	}
}
