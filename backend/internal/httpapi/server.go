package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/config"
	"github.com/MeViksry/Mautrade/backend/internal/domain/gasfee"
	"github.com/MeViksry/Mautrade/backend/internal/mailer"
	"github.com/MeViksry/Mautrade/backend/internal/platform/bscwallet"
	"github.com/MeViksry/Mautrade/backend/internal/platform/exchangebalance"
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
	credentialEncryptor *secrets.Encryptor
	exchangeBalance     *exchangebalance.Client
	gasFeeWithdrawer    *bscwallet.Withdrawer
	mailer              *mailer.Mailer
	logger              *slog.Logger
	mux                 *http.ServeMux
}

func NewServer(cfg config.Config, db *pgxpool.Pool, queueClient *queue.Client, mailer *mailer.Mailer, logger *slog.Logger) (*Server, error) {
	credentialEncryptor, err := secrets.NewEncryptor(cfg.ExchangeCredentialKey, cfg.Environment)
	if err != nil {
		return nil, err
	}
	gasFeeWithdrawer, err := newGasFeeWithdrawer(cfg)
	if err != nil {
		return nil, err
	}
	server := &Server{
		config:              cfg,
		db:                  db,
		queue:               queueClient,
		store:               store.NewDashboardStore(db),
		credentialEncryptor: credentialEncryptor,
		exchangeBalance:     exchangebalance.NewClient(),
		gasFeeWithdrawer:    gasFeeWithdrawer,
		mailer:              mailer,
		logger:              logger,
		mux:                 http.NewServeMux(),
	}
	if err := server.bootstrapAdmin(); err != nil {
		return nil, err
	}
	server.routes()
	return server, nil
}

func newGasFeeWithdrawer(cfg config.Config) (*bscwallet.Withdrawer, error) {
	if strings.TrimSpace(cfg.GasFeeWithdrawKey) == "" {
		return nil, nil
	}

	withdrawer, err := bscwallet.NewWithdrawer(bscwallet.Config{
		RPCURLs:       cfg.GasFeeRPCURLs,
		ChainID:       cfg.GasFeeChainID,
		TokenContract: cfg.GasFeeUSDTContract,
		TokenDecimals: cfg.GasFeeTokenDecimals,
		PrivateKey:    cfg.GasFeeWithdrawKey,
	})
	if err != nil {
		return nil, err
	}

	depositAddress, err := bscwallet.ParseAddress(cfg.GasFeeDepositAddress)
	if err != nil {
		return nil, fmt.Errorf("GAS_FEE_DEPOSIT_ADDRESS must be a valid EVM address when GAS_FEE_WITHDRAW_PRIVATE_KEY is configured: %w", err)
	}
	if !strings.EqualFold(depositAddress.Hex(), withdrawer.SignerAddress()) {
		return nil, fmt.Errorf("GAS_FEE_WITHDRAW_PRIVATE_KEY does not match GAS_FEE_DEPOSIT_ADDRESS")
	}
	return withdrawer, nil
}

func (s *Server) getGasFeeCalculator(ctx context.Context) (gasfee.Calculator, error) {
	settings, err := s.store.GlobalSettings(ctx)
	if err != nil {
		return gasfee.Calculator{}, err
	}
	hundredth, _ := qdecimal.Parse("0.01")
	return gasfee.NewCalculator(settings.GasFeePercentage.Mul(hundredth))
}

func (s *Server) Handler() http.Handler {
	return s.cors(s.mux)
}

func (s *Server) bootstrapAdmin() error {
	// Admin Account One
	if s.store.Ready() && s.config.AdminOneEmail != "" {
		result, err := s.store.BootstrapAdmin(context.Background(), store.BootstrapAdminParams{
			Email:       strings.TrimSpace(s.config.AdminOneEmail),
			Password:    s.config.AdminOnePassword,
			DisplayName: s.config.AdminOneName,
			Role:        "super_admin",
			Now:         time.Now().UTC(),
		})
		if err != nil {
			return err
		}
		if result.Admin.ID != "" {
			s.logger.Info("admin one bootstrap checked", "admin_id", result.Admin.ID, "created", result.Created, "email", s.config.AdminOneEmail)
		}
	}

	// Admin Account Two
	if s.store.Ready() && s.config.AdminTwoEmail != "" {
		result, err := s.store.BootstrapAdmin(context.Background(), store.BootstrapAdminParams{
			Email:       strings.TrimSpace(s.config.AdminTwoEmail),
			Password:    s.config.AdminTwoPassword,
			DisplayName: s.config.AdminTwoName,
			Role:        "super_admin",
			Now:         time.Now().UTC(),
		})
		if err != nil {
			return err
		}
		if result.Admin.ID != "" {
			s.logger.Info("admin two bootstrap checked", "admin_id", result.Admin.ID, "created", result.Created, "email", s.config.AdminTwoEmail)
		}
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
	s.mux.HandleFunc("PUT /api/v1/user/profile", s.handleUpdateUserProfile)
	s.mux.HandleFunc("GET /api/v1/user/stats", s.handleUserStats)
	s.mux.HandleFunc("GET /api/v1/user/exchange-bindings", s.handleExchangeBindings)
	s.mux.HandleFunc("POST /api/v1/user/exchange-bindings", s.handleBindExchange)
	s.mux.HandleFunc("GET /api/v1/user/exchange-bindings/{exchange}/credentials", s.handleExchangeBindingCredentials)
	s.mux.HandleFunc("PATCH /api/v1/user/exchange-bindings/{exchange}/status", s.handleUpdateExchangeBindingStatus)
	s.mux.HandleFunc("PATCH /api/v1/user/exchange-bindings/{exchange}/account-mode", s.handleUpdateExchangeBindingAccountMode)
	s.mux.HandleFunc("DELETE /api/v1/user/exchange-bindings/{exchange}", s.handleDeleteExchangeBinding)
	s.mux.HandleFunc("GET /api/v1/user/gas-fee", s.handleUserGasFeeAccount)
	s.mux.HandleFunc("POST /api/v1/user/gas-fee/deposits", s.handleCreateGasFeeDeposit)
	s.mux.HandleFunc("GET /api/v1/user/layers", s.handleLayers)
	s.mux.HandleFunc("GET /api/v1/user/history", s.handleTradeHistory)
	s.mux.HandleFunc("GET /api/v1/user/history/trades", s.handleTradeHistory)
	s.mux.HandleFunc("POST /api/v1/admin/auth/login", s.handleAdminLogin)
	s.mux.HandleFunc("GET /api/v1/admin/auth/me", s.handleAdminMe)
	s.mux.HandleFunc("POST /api/v1/admin/auth/logout", s.handleAdminLogout)
	s.mux.HandleFunc("POST /api/v1/admin/auth/2fa/setup", s.handleAdmin2FASetup)
	s.mux.HandleFunc("POST /api/v1/admin/auth/2fa/verify", s.handleAdmin2FAVerify)
	s.mux.HandleFunc("POST /api/v1/admin/auth/2fa/disable", s.handleAdmin2FADisable)
	s.mux.HandleFunc("GET /api/v1/admin/users", s.handleAdminListUsers)
	s.mux.HandleFunc("DELETE /api/v1/admin/users/{id}", s.handleAdminDeleteUser)
	s.mux.HandleFunc("GET /api/v1/admin/overview", s.handleAdminOverview)
	s.mux.HandleFunc("GET /api/v1/admin/analytics", s.handleAdminGetAnalytics)
	s.mux.HandleFunc("GET /api/v1/admin/personal-wallets", s.handleAdminPersonalWallets)
	s.mux.HandleFunc("PATCH /api/v1/admin/personal-wallets/{code}", s.handleUpdateAdminPersonalWallet)
	s.mux.HandleFunc("POST /api/v1/admin/personal-wallets/{code}/withdrawals", s.handleCreateAdminPersonalWalletWithdrawal)
	s.mux.HandleFunc("POST /api/v1/admin/gas-fee/preview", s.handleGasFeePreview)
	s.mux.HandleFunc("GET /api/v1/admin/gas-fee/deposits", s.handleAdminGasFeeDeposits)
	s.mux.HandleFunc("PATCH /api/v1/admin/gas-fee/deposits/{deposit_id}/status", s.handleUpdateGasFeeDepositStatus)
	s.mux.HandleFunc("POST /api/v1/admin/signals", s.handleCreateAdminSignal)
	s.mux.HandleFunc("GET /api/v1/admin/signals/active", s.handleAdminListActiveSignals)
	s.mux.HandleFunc("GET /api/v1/admin/signals/orders", s.handleAdminListOpenOrders)
	s.mux.HandleFunc("POST /api/v1/admin/executions/{job_id}/retry", s.handleRetryAdminExecution)
	s.mux.HandleFunc("GET /api/v1/admin/signals/{signal_id}/executions", s.handleAdminSignalExecutions)
	s.mux.HandleFunc("GET /api/v1/admin/reconciliation-events", s.handleAdminReconciliationEvents)
	s.mux.HandleFunc("PATCH /api/v1/admin/reconciliation-events/{event_id}/resolve", s.handleResolveReconciliationEvent)
	s.mux.HandleFunc("GET /api/v1/settings", s.handleGetGlobalSettings)
	s.mux.HandleFunc("PUT /api/v1/admin/settings", s.handleAdminUpdateGlobalSettings)
	s.mux.HandleFunc("GET /api/v1/internal/exchange-bindings/{binding_id}/credentials", s.handleInternalExchangeBindingCredential)
	s.mux.HandleFunc("POST /api/v1/internal/execution-results", s.handleExecutionResult)
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.config.AllowedCORSOrigin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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
