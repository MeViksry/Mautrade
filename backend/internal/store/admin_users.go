package store

import (
	"context"
	"fmt"
	"time"

	"github.com/MeViksry/qdecimal"
)

type AdminEndUserView struct {
	ID                  string           `json:"id"`
	Email               string           `json:"email"`
	DisplayName         string           `json:"displayName"`
	Status              string           `json:"status"`
	EmailVerified       bool             `json:"emailVerified"`
	OnboardingCompleted bool             `json:"onboardingCompleted"`
	CountryCode         string           `json:"countryCode"`
	Age                 int              `json:"age"`
	CreatedAt           time.Time        `json:"createdAt"`
	LastLoginAt         *time.Time       `json:"lastLoginAt,omitempty"`
	GasFeeBalance       qdecimal.Decimal `json:"gasFeeBalance"`
}

func (s *DashboardStore) AdminListUsers(ctx context.Context, search string, limit, offset int) ([]AdminEndUserView, error) {
	const query = `
		SELECT
			id::text, email, display_name, status, (email_verified_at IS NOT NULL) AS email_verified, (onboarding_completed_at IS NOT NULL) AS onboarding_completed, country_code, age, created_at, updated_at,
			(
				COALESCE((SELECT SUM(amount) FROM gas_fee_deposits WHERE user_id = users.id AND status = 'confirmed'), 0) -
				COALESCE((SELECT SUM(gas_fee_amount) FROM gas_fee_ledger WHERE user_id = users.id), 0)
			)::numeric AS gas_fee_balance
		FROM users
		WHERE $1 = '' OR (email ILIKE '%' || $1 || '%' OR display_name ILIKE '%' || $1 || '%' OR id::text ILIKE '%' || $1 || '%')
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.Query(ctx, query, search, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("store: list users: %w", err)
	}
	defer rows.Close()

	var users []AdminEndUserView
	for rows.Next() {
		var u AdminEndUserView
		var (
			email       *string
			displayName *string
			status      *string
			country     *string
			age         *int
			createdAt   *time.Time
			updatedAt   *time.Time
			gasFeeBal   qdecimal.Decimal
		)
		if err := rows.Scan(
			&u.ID, &email, &displayName, &status, &u.EmailVerified, &u.OnboardingCompleted, &country, &age, &createdAt, &updatedAt, &gasFeeBal,
		); err != nil {
			return nil, fmt.Errorf("store: scan user: %w", err)
		}

		if email != nil {
			u.Email = *email
		}
		if displayName != nil {
			u.DisplayName = *displayName
		}
		if status != nil {
			u.Status = *status
		}
		if age != nil {
			u.Age = *age
		}
		if country != nil {
			u.CountryCode = *country
		}
		if createdAt != nil {
			u.CreatedAt = *createdAt
		}
		u.GasFeeBalance = gasFeeBal
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if users == nil {
		users = []AdminEndUserView{}
	}
	return users, nil
}
