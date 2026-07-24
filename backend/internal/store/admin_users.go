package store

import (
	"context"
	"fmt"
	"time"
)

type AdminEndUserView struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	DisplayName         string     `json:"displayName"`
	Status              string     `json:"status"`
	EmailVerified       bool       `json:"emailVerified"`
	OnboardingCompleted bool       `json:"onboardingCompleted"`
	CountryCode         string     `json:"countryCode"`
	Age                 int        `json:"age"`
	CreatedAt           time.Time  `json:"createdAt"`
	LastLoginAt         *time.Time `json:"lastLoginAt,omitempty"`
}

func (s *DashboardStore) AdminListUsers(ctx context.Context, limit, offset int) ([]AdminEndUserView, error) {
	const query = `
		SELECT
			id::text, email, display_name, status, (email_verified_at IS NOT NULL) AS email_verified, (onboarding_completed_at IS NOT NULL) AS onboarding_completed, country_code, age, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := s.db.Query(ctx, query, limit, offset)
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
		)
		if err := rows.Scan(
			&u.ID, &email, &displayName, &status, &u.EmailVerified, &u.OnboardingCompleted, &country, &age, &createdAt, &updatedAt,
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
