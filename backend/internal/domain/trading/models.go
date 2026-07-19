package trading

import (
	"time"

	"github.com/MeViksry/qdecimal"
	"github.com/MeViksry/quuid"
)

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleOps        Role = "ops"
	RoleAuditor    Role = "auditor"
	RoleUser       Role = "user"
)

type ExchangeName string

const (
	ExchangeBinance    ExchangeName = "binance"
	ExchangeOKX        ExchangeName = "okx"
	ExchangeBybit      ExchangeName = "bybit"
	ExchangeTokocrypto ExchangeName = "tokocrypto"
)

type ExchangeBindingStatus string

const (
	BindingActive  ExchangeBindingStatus = "active"
	BindingInvalid ExchangeBindingStatus = "invalid"
	BindingRevoked ExchangeBindingStatus = "revoked"
)

type LayerStatus string

const (
	LayerOpen     LayerStatus = "open"
	LayerPartial  LayerStatus = "partial"
	LayerClosed   LayerStatus = "closed"
	LayerOrphaned LayerStatus = "orphaned"
)

type SignalType string

const (
	SignalBuy  SignalType = "buy"
	SignalSell SignalType = "sell"
)

type SignalStatus string

const (
	SignalDraft       SignalStatus = "draft"
	SignalDispatching SignalStatus = "dispatching"
	SignalCompleted   SignalStatus = "completed"
	SignalFailed      SignalStatus = "failed"
	SignalCancelled   SignalStatus = "cancelled"
)

type User struct {
	ID              quuid.UUID `json:"id"`
	Email           string     `json:"email"`
	Username        string     `json:"username"`
	Name            string     `json:"name"`
	CountryCode     string     `json:"country_code"`
	Timezone        string     `json:"timezone"`
	Status          string     `json:"status"`
	KYCStatus       string     `json:"kyc_status"`
	ThemePreference string     `json:"theme_pref"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ExchangeBinding struct {
	ID              quuid.UUID            `json:"id"`
	UserID          quuid.UUID            `json:"user_id"`
	Exchange        ExchangeName          `json:"exchange"`
	Status          ExchangeBindingStatus `json:"status"`
	PermissionScope string                `json:"permission_scope"`
	LastVerifiedAt  *time.Time            `json:"last_verified_at"`
	CreatedAt       time.Time             `json:"created_at"`
}

type MasterSignal struct {
	ID             quuid.UUID       `json:"id"`
	AdminID        quuid.UUID       `json:"admin_id"`
	Type           SignalType       `json:"type"`
	Symbol         string           `json:"symbol"`
	LayerNumber    *int             `json:"layer_number,omitempty"`
	AllocationPct  qdecimal.Decimal `json:"allocation_pct"`
	SellPct        qdecimal.Decimal `json:"sell_pct"`
	Status         SignalStatus     `json:"status"`
	IdempotencyKey string           `json:"idempotency_key"`
	CreatedAt      time.Time        `json:"created_at"`
}

type Layer struct {
	ID                quuid.UUID       `json:"id"`
	UserID            quuid.UUID       `json:"user_id"`
	ExchangeBindingID quuid.UUID       `json:"exchange_binding_id"`
	MasterSignalID    quuid.UUID       `json:"master_signal_id"`
	LayerNumber       int              `json:"layer_number"`
	Symbol            string           `json:"symbol"`
	EntryPrice        qdecimal.Decimal `json:"entry_price"`
	EntryQuantity     qdecimal.Decimal `json:"entry_quantity"`
	EntryValueQuote   qdecimal.Decimal `json:"entry_value_quote"`
	RemainingQuantity qdecimal.Decimal `json:"remaining_quantity"`
	AllocationPct     qdecimal.Decimal `json:"allocation_pct"`
	Status            LayerStatus      `json:"status"`
	OpenedAt          time.Time        `json:"opened_at"`
	ClosedAt          *time.Time       `json:"closed_at"`
}
