package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/config"
	"github.com/MeViksry/Mautrade/backend/internal/domain/gasfee"
	"github.com/MeViksry/Mautrade/backend/internal/platform/queue"
	"github.com/MeViksry/Mautrade/backend/internal/platform/secrets"
	"github.com/MeViksry/Mautrade/backend/internal/store"
	"github.com/MeViksry/qdecimal"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	config              config.Config
	db                  *pgxpool.Pool
	queue               *queue.Client
	store               *store.DashboardStore
	gasFeeCalc          gasfee.Calculator
	credentialEncryptor *secrets.Encryptor
	logger              *slog.Logger
	mux                 *http.ServeMux
}

func NewServer(cfg config.Config, db *pgxpool.Pool, queueClient *queue.Client, logger *slog.Logger) (*Server, error) {
	shareRate, err := qdecimal.Parse(cfg.GasFeeShareRate)
	if err != nil {
		return nil, err
	}
	calculator, err := gasfee.NewCalculator(shareRate)
	if err != nil {
		return nil, err
	}
	credentialEncryptor, err := secrets.NewEncryptor(cfg.ExchangeCredentialKey, cfg.Environment)
	if err != nil {
		return nil, err
	}
	server := &Server{
		config:              cfg,
		db:                  db,
		queue:               queueClient,
		store:               store.NewDashboardStore(db),
		gasFeeCalc:          calculator,
		credentialEncryptor: credentialEncryptor,
		logger:              logger,
		mux:                 http.NewServeMux(),
	}
	if err := server.bootstrapAdmin(); err != nil {
		return nil, err
	}
	server.routes()
	return server, nil
}

func (s *Server) Handler() http.Handler {
	return s.cors(s.mux)
}

func (s *Server) bootstrapAdmin() error {
	if !s.store.Ready() || s.config.AdminBootstrapEmail == "" {
		return nil
	}
	result, err := s.store.BootstrapAdmin(context.Background(), store.BootstrapAdminParams{
		Email:       s.config.AdminBootstrapEmail,
		Password:    s.config.AdminBootstrapPassword,
		DisplayName: s.config.AdminBootstrapName,
		Role:        s.config.AdminBootstrapRole,
		Now:         time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	if result.Admin.ID != "" {
		s.logger.Info("admin bootstrap checked", "admin_id", result.Admin.ID, "created", result.Created)
	}
	return nil
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /readyz", s.handleReadyz)
	s.mux.HandleFunc("POST /api/v1/auth/register", s.handleRegister)
	s.mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("POST /api/v1/auth/verify-otp", s.handleVerifyEmailOTP)
	s.mux.HandleFunc("GET /api/v1/auth/me", s.handleMe)
	s.mux.HandleFunc("POST /api/v1/auth/logout", s.handleLogout)
	s.mux.HandleFunc("POST /api/v1/onboarding/complete", s.handleCompleteOnboarding)
	s.mux.HandleFunc("GET /api/v1/user/stats", s.handleUserStats)
	s.mux.HandleFunc("GET /api/v1/user/exchange-bindings", s.handleExchangeBindings)
	s.mux.HandleFunc("POST /api/v1/user/exchange-bindings", s.handleBindExchange)
	s.mux.HandleFunc("GET /api/v1/user/exchange-bindings/{exchange}/credentials", s.handleExchangeBindingCredentials)
	s.mux.HandleFunc("PATCH /api/v1/user/exchange-bindings/{exchange}/status", s.handleUpdateExchangeBindingStatus)
	s.mux.HandleFunc("DELETE /api/v1/user/exchange-bindings/{exchange}", s.handleDeleteExchangeBinding)
	s.mux.HandleFunc("GET /api/v1/user/gas-fee", s.handleUserGasFeeAccount)
	s.mux.HandleFunc("POST /api/v1/user/gas-fee/deposits", s.handleCreateGasFeeDeposit)
	s.mux.HandleFunc("GET /api/v1/user/layers", s.handleLayers)
	s.mux.HandleFunc("GET /api/v1/user/history/trades", s.handleTradeHistory)
	s.mux.HandleFunc("POST /api/v1/admin/auth/login", s.handleAdminLogin)
	s.mux.HandleFunc("GET /api/v1/admin/auth/me", s.handleAdminMe)
	s.mux.HandleFunc("POST /api/v1/admin/auth/logout", s.handleAdminLogout)
	s.mux.HandleFunc("POST /api/v1/admin/auth/2fa/setup", s.handleAdmin2FASetup)
	s.mux.HandleFunc("POST /api/v1/admin/auth/2fa/verify", s.handleAdmin2FAVerify)
	s.mux.HandleFunc("POST /api/v1/admin/auth/2fa/disable", s.handleAdmin2FADisable)
	s.mux.HandleFunc("GET /api/v1/admin/overview", s.handleAdminOverview)
	s.mux.HandleFunc("POST /api/v1/admin/gas-fee/preview", s.handleGasFeePreview)
	s.mux.HandleFunc("GET /api/v1/admin/gas-fee/deposits", s.handleAdminGasFeeDeposits)
	s.mux.HandleFunc("PATCH /api/v1/admin/gas-fee/deposits/{deposit_id}/status", s.handleUpdateGasFeeDepositStatus)
	s.mux.HandleFunc("POST /api/v1/admin/signals", s.handleCreateAdminSignal)
	s.mux.HandleFunc("POST /api/v1/admin/executions/{job_id}/retry", s.handleRetryAdminExecution)
	s.mux.HandleFunc("GET /api/v1/admin/signals/{signal_id}/executions", s.handleAdminSignalExecutions)
	s.mux.HandleFunc("GET /api/v1/admin/reconciliation-events", s.handleAdminReconciliationEvents)
	s.mux.HandleFunc("PATCH /api/v1/admin/reconciliation-events/{event_id}/resolve", s.handleResolveReconciliationEvent)
	s.mux.HandleFunc("POST /api/v1/internal/execution-results", s.handleExecutionResult)
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.config.AllowedCORSOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if s.db != nil {
		if err := s.db.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "degraded", "postgres": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
