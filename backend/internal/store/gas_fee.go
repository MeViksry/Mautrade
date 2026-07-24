package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MeViksry/qdecimal"
	"github.com/jackc/pgx/v5"
)

var (
	ErrGasFeeDepositNotFound     = errors.New("store: gas fee deposit not found")
	ErrGasFeeDepositStatus       = errors.New("store: invalid gas fee deposit status")
	ErrGasFeeDepositAmount       = errors.New("store: invalid gas fee deposit amount")
	ErrGasFeeDepositTransition   = errors.New("store: invalid gas fee deposit transition")
	ErrGasFeeDepositTxIDRequired = errors.New("store: gas fee deposit tx id required")
)

type GasFeeAccountView struct {
	Asset             string              `json:"asset"`
	Balance           string              `json:"balance"`
	ConfirmedDeposits string              `json:"confirmedDeposits"`
	PendingDeposits   string              `json:"pendingDeposits"`
	NetGasFeeMovement string              `json:"netGasFeeMovement"`
	FeesUsed          string              `json:"feesUsed"`
	Rebates           string              `json:"rebates"`
	MinimumDeposit    string              `json:"minimumDeposit"`
	History           []GasFeeHistoryItem `json:"history"`
	ServerTime        time.Time           `json:"serverTime"`
}

type GasFeeHistoryItem struct {
	ID            string     `json:"id"`
	Kind          string     `json:"kind"`
	Type          string     `json:"type"`
	Status        string     `json:"status"`
	Amount        string     `json:"amount"`
	BalanceImpact string     `json:"balanceImpact"`
	Asset         string     `json:"asset"`
	Reference     string     `json:"reference"`
	TxID          *string    `json:"txId,omitempty"`
	LayerID       *string    `json:"layerId,omitempty"`
	ExecutionID   *string    `json:"executionId,omitempty"`
	GrossPnL      *string    `json:"grossPnl,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	ConfirmedAt   *time.Time `json:"confirmedAt,omitempty"`
}

type GasFeeDepositView struct {
	ID             string     `json:"id"`
	UserID         string     `json:"userId"`
	UserEmail      string     `json:"userEmail,omitempty"`
	UserDisplay    string     `json:"userDisplayName,omitempty"`
	Amount         string     `json:"amount"`
	Asset          string     `json:"asset"`
	DepositAddress string     `json:"depositAddress"`
	TxID           *string    `json:"txId,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"createdAt"`
	ConfirmedAt    *time.Time `json:"confirmedAt,omitempty"`
}

type CreateGasFeeDepositParams struct {
	UserID         string
	Amount         string
	Asset          string
	DepositAddress string
	TxID           string
	Now            time.Time
}

type AdminGasFeeDepositsParams struct {
	Status string
	Limit  int
	Offset int
}

type UpdateGasFeeDepositStatusParams struct {
	DepositID      string
	AdminID        string
	Status         string
	ResolutionNote string
	Now            time.Time
}

func (s *DashboardStore) UserGasFeeAccount(ctx context.Context, userID, asset string, limit int) (GasFeeAccountView, error) {
	if !s.Ready() {
		return GasFeeAccountView{}, fmt.Errorf("store: gas fee account requires postgres")
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	asset = normalizeGasFeeAsset(asset)
	settings, err := s.GlobalSettings(ctx)
	if err != nil {
		return GasFeeAccountView{}, fmt.Errorf("store: fetch settings for gas fee account: %w", err)
	}
	account := GasFeeAccountView{
		Asset:          asset,
		MinimumDeposit: settings.MinDepositUsdt.String(),
		ServerTime:     time.Now().UTC(),
	}
	const summaryQuery = `
WITH deposit_sum AS (
  SELECT
    COALESCE(SUM(amount) FILTER (WHERE status = 'confirmed'), 0) AS confirmed_deposits,
    COALESCE(SUM(amount) FILTER (WHERE status = 'pending'), 0) AS pending_deposits
  FROM gas_fee_deposits
  WHERE user_id = $1::uuid
    AND asset = $2
),
ledger_sum AS (
  SELECT
    COALESCE(SUM(gas_fee_amount), 0) AS net_gas_fee_movement,
    COALESCE(SUM(CASE WHEN gas_fee_amount > 0 THEN gas_fee_amount ELSE 0 END), 0) AS fees_used,
    COALESCE(SUM(CASE WHEN gas_fee_amount < 0 THEN -gas_fee_amount ELSE 0 END), 0) AS rebates
  FROM gas_fee_ledger
  WHERE user_id = $1::uuid
)
SELECT
  (deposit_sum.confirmed_deposits - ledger_sum.net_gas_fee_movement)::text AS balance,
  deposit_sum.confirmed_deposits::text,
  deposit_sum.pending_deposits::text,
  ledger_sum.net_gas_fee_movement::text,
  ledger_sum.fees_used::text,
  ledger_sum.rebates::text
FROM deposit_sum, ledger_sum`
	if err := s.db.QueryRow(ctx, summaryQuery, userID, asset).Scan(
		&account.Balance,
		&account.ConfirmedDeposits,
		&account.PendingDeposits,
		&account.NetGasFeeMovement,
		&account.FeesUsed,
		&account.Rebates,
	); err != nil {
		return GasFeeAccountView{}, fmt.Errorf("store: gas fee account summary: %w", err)
	}
	history, err := s.UserGasFeeHistory(ctx, userID, asset, limit)
	if err != nil {
		return GasFeeAccountView{}, err
	}
	account.History = history
	return account, nil
}

func (s *DashboardStore) UserGasFeeHistory(ctx context.Context, userID, asset string, limit int) ([]GasFeeHistoryItem, error) {
	if !s.Ready() {
		return nil, fmt.Errorf("store: gas fee history requires postgres")
	}
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	asset = normalizeGasFeeAsset(asset)
	const query = `
SELECT
  d.id::text,
  'deposit' AS kind,
  'deposit' AS type,
  d.status,
  d.amount::text,
  CASE WHEN d.status = 'confirmed' THEN d.amount ELSE 0 END::text AS balance_impact,
  d.asset,
  d.deposit_address,
  d.tx_id,
  NULL::text AS layer_id,
  NULL::text AS execution_id,
  NULL::text AS gross_pnl,
  d.created_at,
  d.confirmed_at,
  d.created_at AS ordered_at
FROM gas_fee_deposits d
WHERE d.user_id = $1::uuid
  AND d.asset = $2
UNION ALL
SELECT
  g.id::text,
  'ledger' AS kind,
  g.type,
  'confirmed' AS status,
  g.gas_fee_amount::text,
  (-g.gas_fee_amount)::text AS balance_impact,
  $2 AS asset,
  l.symbol AS reference,
  NULL::text AS tx_id,
  g.layer_id::text,
  g.execution_id::text,
  g.gross_pnl::text,
  g.calculated_at AS created_at,
  NULL::timestamptz AS confirmed_at,
  g.calculated_at AS ordered_at
FROM gas_fee_ledger g
JOIN layers l ON l.id = g.layer_id
WHERE g.user_id = $1::uuid
ORDER BY ordered_at DESC
LIMIT $3`

	rows, err := s.db.Query(ctx, query, userID, asset, limit)
	if err != nil {
		return nil, fmt.Errorf("store: gas fee history: %w", err)
	}
	defer rows.Close()

	var history []GasFeeHistoryItem
	for rows.Next() {
		var item GasFeeHistoryItem
		if err := rows.Scan(
			&item.ID,
			&item.Kind,
			&item.Type,
			&item.Status,
			&item.Amount,
			&item.BalanceImpact,
			&item.Asset,
			&item.Reference,
			&item.TxID,
			&item.LayerID,
			&item.ExecutionID,
			&item.GrossPnL,
			&item.CreatedAt,
			&item.ConfirmedAt,
			new(time.Time),
		); err != nil {
			return nil, fmt.Errorf("store: scan gas fee history: %w", err)
		}
		history = append(history, item)
	}
	if history == nil {
		history = []GasFeeHistoryItem{}
	}
	return history, rows.Err()
}

func (s *DashboardStore) CreateGasFeeDeposit(ctx context.Context, params CreateGasFeeDepositParams) (GasFeeDepositView, error) {
	if !s.Ready() {
		return GasFeeDepositView{}, fmt.Errorf("store: create gas fee deposit requires postgres")
	}
	now := normalizedNow(params.Now)
	asset := normalizeGasFeeAsset(params.Asset)
	amount := decimalOrZero(params.Amount)
	settings, err := s.GlobalSettings(ctx)
	if err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: fetch settings for gas fee deposit: %w", err)
	}
	parsed, err := qdecimal.Parse(amount)
	if err != nil || parsed.Cmp(settings.MinDepositUsdt) < 0 {
		return GasFeeDepositView{}, fmt.Errorf("store: gas fee deposit amount must be at least %s %s", settings.MinDepositUsdt.String(), asset)
	}
	txID := strings.TrimSpace(params.TxID)
	if txID == "" {
		return GasFeeDepositView{}, ErrGasFeeDepositTxIDRequired
	}
	address := strings.TrimSpace(params.DepositAddress)
	if address == "" {
		address = "MAUTRADE-USDT-DEPOSIT-PENDING"
	}
	depositID, err := newUUIDText()
	if err != nil {
		return GasFeeDepositView{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: begin create gas fee deposit: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
INSERT INTO gas_fee_deposits (
  id, user_id, amount, asset, deposit_address, tx_id, status, created_at
) VALUES (
  $1::uuid, $2::uuid, $3::numeric, $4, $5, $6, 'pending', $7
)`, depositID, params.UserID, amount, asset, address, txID, now); err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: insert gas fee deposit: %w", err)
	}
	if err := insertGasFeeDepositAudit(ctx, tx, "user", params.UserID, "gas_fee_deposit_created", depositID, map[string]any{
		"amount": amount,
		"asset":  asset,
		"status": "pending",
		"tx_id":  txID,
	}, now); err != nil {
		return GasFeeDepositView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: commit create gas fee deposit: %w", err)
	}
	return s.gasFeeDepositByID(ctx, depositID)
}

func (s *DashboardStore) AdminGasFeeDeposits(ctx context.Context, params AdminGasFeeDepositsParams) ([]GasFeeDepositView, error) {
	if !s.Ready() {
		return nil, fmt.Errorf("store: admin gas fee deposits requires postgres")
	}
	if params.Limit <= 0 || params.Limit > 500 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	status := strings.ToLower(strings.TrimSpace(params.Status))
	const query = `
SELECT
  d.id::text,
  d.user_id::text,
  u.email,
  u.display_name,
  d.amount::text,
  d.asset,
  d.deposit_address,
  d.tx_id,
  d.status,
  d.created_at,
  d.confirmed_at
FROM gas_fee_deposits d
JOIN users u ON u.id = d.user_id
WHERE ($1 = '' OR d.status = $1)
ORDER BY d.created_at DESC
LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(ctx, query, status, params.Limit, params.Offset)
	if err != nil {
		return nil, fmt.Errorf("store: admin gas fee deposits: %w", err)
	}
	defer rows.Close()

	var deposits []GasFeeDepositView
	for rows.Next() {
		var deposit GasFeeDepositView
		if err := rows.Scan(
			&deposit.ID,
			&deposit.UserID,
			&deposit.UserEmail,
			&deposit.UserDisplay,
			&deposit.Amount,
			&deposit.Asset,
			&deposit.DepositAddress,
			&deposit.TxID,
			&deposit.Status,
			&deposit.CreatedAt,
			&deposit.ConfirmedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan admin gas fee deposit: %w", err)
		}
		deposits = append(deposits, deposit)
	}
	if deposits == nil {
		deposits = []GasFeeDepositView{}
	}
	return deposits, rows.Err()
}

func (s *DashboardStore) UpdateGasFeeDepositStatus(ctx context.Context, params UpdateGasFeeDepositStatusParams) (GasFeeDepositView, error) {
	if !s.Ready() {
		return GasFeeDepositView{}, fmt.Errorf("store: update gas fee deposit requires postgres")
	}
	now := normalizedNow(params.Now)
	status := strings.ToLower(strings.TrimSpace(params.Status))
	if status != "confirmed" && status != "rejected" {
		return GasFeeDepositView{}, ErrGasFeeDepositStatus
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: begin update gas fee deposit: %w", err)
	}
	defer tx.Rollback(ctx)

	var previous GasFeeDepositView
	if err := tx.QueryRow(ctx, `
SELECT
  d.id::text,
  d.user_id::text,
  u.email,
  u.display_name,
  d.amount::text,
  d.asset,
  d.deposit_address,
  d.tx_id,
  d.status,
  d.created_at,
  d.confirmed_at
FROM gas_fee_deposits d
JOIN users u ON u.id = d.user_id
WHERE d.id = $1::uuid
FOR UPDATE`, params.DepositID).Scan(
		&previous.ID,
		&previous.UserID,
		&previous.UserEmail,
		&previous.UserDisplay,
		&previous.Amount,
		&previous.Asset,
		&previous.DepositAddress,
		&previous.TxID,
		&previous.Status,
		&previous.CreatedAt,
		&previous.ConfirmedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GasFeeDepositView{}, ErrGasFeeDepositNotFound
		}
		return GasFeeDepositView{}, fmt.Errorf("store: lock gas fee deposit: %w", err)
	}
	if previous.Status != "pending" {
		return GasFeeDepositView{}, ErrGasFeeDepositTransition
	}

	if _, err := tx.Exec(ctx, `
UPDATE gas_fee_deposits
SET status = $2,
    confirmed_at = CASE WHEN $2 = 'confirmed' THEN $3 ELSE NULL END
WHERE id = $1::uuid`, params.DepositID, status, now); err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: update gas fee deposit: %w", err)
	}
	if err := insertGasFeeDepositAudit(ctx, tx, "admin", params.AdminID, "gas_fee_deposit_"+status, params.DepositID, map[string]any{
		"previous_status": previous.Status,
		"status":          status,
		"amount":          previous.Amount,
		"asset":           previous.Asset,
		"user_id":         previous.UserID,
		"note":            strings.TrimSpace(params.ResolutionNote),
	}, now); err != nil {
		return GasFeeDepositView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: commit update gas fee deposit: %w", err)
	}
	return s.gasFeeDepositByID(ctx, params.DepositID)
}

func (s *DashboardStore) PendingGasFeeDeposits(ctx context.Context, limit int) ([]GasFeeDepositView, error) {
	if !s.Ready() {
		return nil, fmt.Errorf("store: pending gas fee deposits requires postgres")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	const query = `
SELECT
  d.id::text,
  d.user_id::text,
  u.email,
  u.display_name,
  d.amount::text,
  d.asset,
  d.deposit_address,
  d.tx_id,
  d.status,
  d.created_at,
  d.confirmed_at
FROM gas_fee_deposits d
JOIN users u ON u.id = d.user_id
WHERE d.status = 'pending' AND d.tx_id IS NOT NULL AND d.tx_id != ''
ORDER BY d.created_at ASC
LIMIT $1`
	rows, err := s.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("store: pending gas fee deposits: %w", err)
	}
	defer rows.Close()

	var deposits []GasFeeDepositView
	for rows.Next() {
		var deposit GasFeeDepositView
		if err := rows.Scan(
			&deposit.ID,
			&deposit.UserID,
			&deposit.UserEmail,
			&deposit.UserDisplay,
			&deposit.Amount,
			&deposit.Asset,
			&deposit.DepositAddress,
			&deposit.TxID,
			&deposit.Status,
			&deposit.CreatedAt,
			&deposit.ConfirmedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan pending gas fee deposit: %w", err)
		}
		deposits = append(deposits, deposit)
	}
	if deposits == nil {
		deposits = []GasFeeDepositView{}
	}
	return deposits, rows.Err()
}

type SystemUpdateGasFeeDepositStatusParams struct {
	DepositID      string
	Status         string
	ResolutionNote string
	Now            time.Time
}

func (s *DashboardStore) SystemUpdateGasFeeDepositStatus(ctx context.Context, params SystemUpdateGasFeeDepositStatusParams) (GasFeeDepositView, error) {
	if !s.Ready() {
		return GasFeeDepositView{}, fmt.Errorf("store: system update gas fee deposit requires postgres")
	}
	now := normalizedNow(params.Now)
	status := strings.ToLower(strings.TrimSpace(params.Status))
	if status != "confirmed" && status != "rejected" && status != "failed" {
		return GasFeeDepositView{}, ErrGasFeeDepositStatus
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: begin system update gas fee deposit: %w", err)
	}
	defer tx.Rollback(ctx)

	var previous GasFeeDepositView
	if err := tx.QueryRow(ctx, `
SELECT
  d.id::text,
  d.user_id::text,
  u.email,
  u.display_name,
  d.amount::text,
  d.asset,
  d.deposit_address,
  d.tx_id,
  d.status,
  d.created_at,
  d.confirmed_at
FROM gas_fee_deposits d
JOIN users u ON u.id = d.user_id
WHERE d.id = $1::uuid
FOR UPDATE`, params.DepositID).Scan(
		&previous.ID,
		&previous.UserID,
		&previous.UserEmail,
		&previous.UserDisplay,
		&previous.Amount,
		&previous.Asset,
		&previous.DepositAddress,
		&previous.TxID,
		&previous.Status,
		&previous.CreatedAt,
		&previous.ConfirmedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GasFeeDepositView{}, ErrGasFeeDepositNotFound
		}
		return GasFeeDepositView{}, fmt.Errorf("store: lock gas fee deposit: %w", err)
	}
	if previous.Status != "pending" {
		return GasFeeDepositView{}, ErrGasFeeDepositTransition
	}

	// Prevent duplicate successful TXID
	if status == "confirmed" && previous.TxID != nil {
		var exists bool
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM gas_fee_deposits WHERE tx_id = $1 AND status = 'confirmed' AND id != $2::uuid)`, *previous.TxID, params.DepositID).Scan(&exists)
		if err != nil {
			return GasFeeDepositView{}, fmt.Errorf("store: check duplicate tx_id: %w", err)
		}
		if exists {
			// TXID already confirmed by someone else, we must fail this one
			status = "failed"
			params.ResolutionNote = "TXID already claimed by another user."
		}
	}

	if _, err := tx.Exec(ctx, `
UPDATE gas_fee_deposits
SET status = $2,
    confirmed_at = CASE WHEN $2 = 'confirmed' THEN $3 ELSE NULL END
WHERE id = $1::uuid`, params.DepositID, status, now); err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: update gas fee deposit: %w", err)
	}

	if err := insertGasFeeDepositAudit(ctx, tx, "system", "", "gas_fee_deposit_"+status, params.DepositID, map[string]any{
		"previous_status": previous.Status,
		"status":          status,
		"amount":          previous.Amount,
		"asset":           previous.Asset,
		"user_id":         previous.UserID,
		"note":            strings.TrimSpace(params.ResolutionNote),
	}, now); err != nil {
		return GasFeeDepositView{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return GasFeeDepositView{}, fmt.Errorf("store: commit system update gas fee deposit: %w", err)
	}
	return s.gasFeeDepositByID(ctx, params.DepositID)
}

func (s *DashboardStore) gasFeeDepositByID(ctx context.Context, depositID string) (GasFeeDepositView, error) {
	var deposit GasFeeDepositView
	err := s.db.QueryRow(ctx, `
SELECT
  d.id::text,
  d.user_id::text,
  u.email,
  u.display_name,
  d.amount::text,
  d.asset,
  d.deposit_address,
  d.tx_id,
  d.status,
  d.created_at,
  d.confirmed_at
FROM gas_fee_deposits d
JOIN users u ON u.id = d.user_id
WHERE d.id = $1::uuid`, depositID).Scan(
		&deposit.ID,
		&deposit.UserID,
		&deposit.UserEmail,
		&deposit.UserDisplay,
		&deposit.Amount,
		&deposit.Asset,
		&deposit.DepositAddress,
		&deposit.TxID,
		&deposit.Status,
		&deposit.CreatedAt,
		&deposit.ConfirmedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GasFeeDepositView{}, ErrGasFeeDepositNotFound
		}
		return GasFeeDepositView{}, fmt.Errorf("store: read gas fee deposit: %w", err)
	}
	return deposit, nil
}

func normalizeGasFeeAsset(value string) string {
	asset := strings.ToUpper(strings.TrimSpace(value))
	if asset == "" {
		return "USDT"
	}
	return asset
}

func insertGasFeeDepositAudit(ctx context.Context, tx pgx.Tx, actorType, actorID, action, depositID string, after map[string]any, now time.Time) error {
	auditID, err := newUUIDText()
	if err != nil {
		return err
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return fmt.Errorf("store: marshal gas fee deposit audit: %w", err)
	}
	var actor any
	if strings.TrimSpace(actorID) != "" {
		actor = strings.TrimSpace(actorID)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO audit_logs (
  id, actor_type, actor_id, action, entity, entity_id, after_state, created_at
) VALUES (
  $1::uuid, $2, $3::uuid, $4, 'gas_fee_deposit', $5::uuid, $6::jsonb, $7
)`, auditID, actorType, actor, action, depositID, string(afterJSON), now); err != nil {
		return fmt.Errorf("store: insert gas fee deposit audit: %w", err)
	}
	return nil
}
