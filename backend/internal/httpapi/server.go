package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/config"
	"github.com/MeViksry/Mautrade/backend/internal/domain/gasfee"
	"github.com/MeViksry/Mautrade/backend/internal/platform/queue"
	"github.com/MeViksry/Mautrade/backend/internal/store"
	"github.com/MeViksry/qdecimal"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	config     config.Config
	db         *pgxpool.Pool
	queue      *queue.Client
	store      *store.DashboardStore
	gasFeeCalc gasfee.Calculator
	logger     *slog.Logger
	mux        *http.ServeMux
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
	server := &Server{
		config:     cfg,
		db:         db,
		queue:      queueClient,
		store:      store.NewDashboardStore(db),
		gasFeeCalc: calculator,
		logger:     logger,
		mux:        http.NewServeMux(),
	}
	server.routes()
	return server, nil
}

func (s *Server) Handler() http.Handler {
	return s.cors(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /readyz", s.handleReadyz)
	s.mux.HandleFunc("GET /api/v1/user/stats", s.handleUserStats)
	s.mux.HandleFunc("GET /api/v1/user/exchange-bindings", s.handleExchangeBindings)
	s.mux.HandleFunc("GET /api/v1/user/layers", s.handleLayers)
	s.mux.HandleFunc("GET /api/v1/user/history/trades", s.handleTradeHistory)
	s.mux.HandleFunc("GET /api/v1/admin/overview", s.handleAdminOverview)
	s.mux.HandleFunc("POST /api/v1/admin/gas-fee/preview", s.handleGasFeePreview)
	s.mux.HandleFunc("POST /api/v1/admin/signals", s.handleCreateAdminSignal)
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
