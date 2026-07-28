package store

import (
	"context"

	"github.com/MeViksry/qdecimal"
)

type AdminAnalyticsData struct {
	TotalRevenue         qdecimal.Decimal     `json:"totalRevenue"`
	TotalUsers           int64                `json:"totalUsers"`
	ActiveUsers          int64                `json:"activeUsers"`
	Transactions         int64                `json:"transactions"`
	DepositGasFeeTracker qdecimal.Decimal     `json:"depositGasFeeTracker"`
	RecentSignups        int64                `json:"recentSignups"`
	SignupsChartData     []DailySignupData    `json:"signupsChartData"`
	CountryDemographics  []CountryDemographic `json:"countryDemographics"`
}

type DailySignupData struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Count int64  `json:"count"`
}

type CountryDemographic struct {
	CountryCode string `json:"countryCode"`
	Count       int64  `json:"count"`
}

// GetAdminAnalytics retrieves aggregated statistics for the admin dashboard.
func (s *DashboardStore) AdminGetAnalytics(ctx context.Context) (AdminAnalyticsData, error) {
	var data AdminAnalyticsData

	// Query TotalRevenue from confirmed user gas fee deposits.
	err := s.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM gas_fee_deposits WHERE status = 'confirmed';`).Scan(&data.TotalRevenue)
	if err != nil {
		return data, err
	}

	// Query TotalUsers
	err = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users;`).Scan(&data.TotalUsers)
	if err != nil {
		return data, err
	}

	// Query ActiveUsers
	err = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status = 'active';`).Scan(&data.ActiveUsers)
	if err != nil {
		return data, err
	}

	// Query Transactions (Successful Layer Executions)
	err = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM layer_executions WHERE status = 'success';`).Scan(&data.Transactions)
	if err != nil {
		return data, err
	}

	// Query DepositGasFeeTracker
	err = s.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM gas_fee_deposits WHERE status = 'confirmed';`).Scan(&data.DepositGasFeeTracker)
	if err != nil {
		return data, err
	}

	// Query RecentSignups (Last 7 days)
	err = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '7 days';`).Scan(&data.RecentSignups)
	if err != nil {
		return data, err
	}

	// Query SignupsChartData (Grouped by Day for the last 7 days)
	rows, err := s.db.Query(ctx, `
		SELECT TO_CHAR(created_at, 'YYYY-MM-DD') AS date, COUNT(*) 
		FROM users 
		WHERE created_at >= NOW() - INTERVAL '7 days' 
		GROUP BY date 
		ORDER BY date ASC;
	`)
	if err != nil {
		return data, err
	}
	defer rows.Close()

	for rows.Next() {
		var d DailySignupData
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return data, err
		}
		data.SignupsChartData = append(data.SignupsChartData, d)
	}

	// Query CountryDemographics
	cRows, err := s.db.Query(ctx, `
		SELECT COALESCE(country_code, 'Unknown') AS country_code, COUNT(*) 
		FROM users 
		GROUP BY COALESCE(country_code, 'Unknown')
		ORDER BY count DESC
		LIMIT 10;
	`)
	if err != nil {
		return data, err
	}
	defer cRows.Close()

	for cRows.Next() {
		var c CountryDemographic
		if err := cRows.Scan(&c.CountryCode, &c.Count); err != nil {
			return data, err
		}
		data.CountryDemographics = append(data.CountryDemographics, c)
	}

	// Ensure arrays are not nil
	if data.SignupsChartData == nil {
		data.SignupsChartData = []DailySignupData{}
	}
	if data.CountryDemographics == nil {
		data.CountryDemographics = []CountryDemographic{}
	}

	return data, nil
}
