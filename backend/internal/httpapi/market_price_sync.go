package httpapi

import (
	"context"
	"time"

	"github.com/MeViksry/Mautrade/backend/internal/store"
)

const marketPriceSyncStaleAfter = 15 * time.Second
const marketPriceSyncTimeout = 5 * time.Second

func (s *Server) syncDueActiveLayerMarketPrices(ctx context.Context, userID string) {
	if s == nil || s.marketPrice == nil || !s.store.Ready() {
		return
	}

	staleBefore := time.Now().UTC().Add(-marketPriceSyncStaleAfter)
	symbols, err := s.store.ActiveLayerPriceRefreshTargets(ctx, userID, staleBefore)
	if err != nil {
		s.logger.Warn("load active layer market price targets", "user_id", userID, "error", err)
		return
	}
	if len(symbols) == 0 {
		return
	}

	syncCtx, cancel := context.WithTimeout(ctx, marketPriceSyncTimeout)
	prices, err := s.marketPrice.FetchSpotPrices(syncCtx, symbols)
	cancel()
	if err != nil {
		s.logger.Warn("sync active layer market prices", "user_id", userID, "symbols", symbols, "error", err)
		return
	}

	now := time.Now().UTC()
	snapshots := make([]store.MarketPriceSnapshot, 0, len(prices))
	for _, price := range prices {
		snapshots = append(snapshots, store.MarketPriceSnapshot{
			Symbol:     price.Symbol,
			PriceQuote: price.PriceQuote,
			Source:     "binance_public",
			CapturedAt: now,
		})
	}
	if err := s.store.RecordMarketPrices(ctx, snapshots); err != nil {
		s.logger.Warn("record active layer market prices", "user_id", userID, "error", err)
	}
}
