package store

import (
	"context"
	"fmt"
	"time"

	"github.com/MeViksry/qdecimal"
)

type GlobalSettingsView struct {
	MaintenanceMode    bool             `json:"maintenanceMode"`
	AllowRegistrations bool             `json:"allowRegistrations"`
	GasFeePercentage   qdecimal.Decimal `json:"gasFeePercentage"`
	MinDepositUsdt     qdecimal.Decimal `json:"minDepositUsdt"`

	SupportEmail string `json:"supportEmail"`
}

func (s *DashboardStore) GlobalSettings(ctx context.Context) (GlobalSettingsView, error) {
	var v GlobalSettingsView
	if !s.Ready() {
		return v, fmt.Errorf("store: settings requires postgres")
	}

	err := s.db.QueryRow(ctx, `
SELECT
  maintenance_mode,
  allow_registrations,
  gas_fee_percentage,
  min_deposit_usdt,
  support_email
FROM global_settings
WHERE id = 1
`).Scan(
		&v.MaintenanceMode,
		&v.AllowRegistrations,
		&v.GasFeePercentage,
		&v.MinDepositUsdt,
		&v.SupportEmail,
	)
	if err != nil {
		return v, fmt.Errorf("store: scan global_settings: %w", err)
	}

	return v, nil
}

type UpdateGlobalSettingsParams struct {
	MaintenanceMode    bool
	AllowRegistrations bool
	GasFeePercentage   qdecimal.Decimal
	MinDepositUsdt     qdecimal.Decimal
	SupportEmail       string
	Now                time.Time
}

func (s *DashboardStore) UpdateGlobalSettings(ctx context.Context, params UpdateGlobalSettingsParams) (GlobalSettingsView, error) {
	if !s.Ready() {
		return GlobalSettingsView{}, fmt.Errorf("store: update settings requires postgres")
	}

	now := normalizedNow(params.Now)
	if _, err := s.db.Exec(ctx, `
UPDATE global_settings
SET
  maintenance_mode = $1,
  allow_registrations = $2,
  gas_fee_percentage = $3,
  min_deposit_usdt = $4,
  support_email = $5,
  updated_at = $6
WHERE id = 1
`,
		params.MaintenanceMode,
		params.AllowRegistrations,
		params.GasFeePercentage,
		params.MinDepositUsdt,

		params.SupportEmail,
		now,
	); err != nil {
		return GlobalSettingsView{}, fmt.Errorf("store: update global_settings: %w", err)
	}

	return s.GlobalSettings(ctx)
}
