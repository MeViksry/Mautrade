package store

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type layerNumberSeed struct {
	Symbol      string
	LayerNumber int
}

type fakeSignalTx struct {
	masterSignals []layerNumberSeed
	layers        []layerNumberSeed
	locks         []string
	queries       []fakeSignalQuery
}

type fakeSignalQuery struct {
	SQL  string
	Args []any
}

func (tx *fakeSignalTx) Begin(context.Context) (pgx.Tx, error) {
	return tx, nil
}

func (tx *fakeSignalTx) Commit(context.Context) error {
	return nil
}

func (tx *fakeSignalTx) Rollback(context.Context) error {
	return nil
}

func (tx *fakeSignalTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, errors.New("unexpected CopyFrom")
}

func (tx *fakeSignalTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	return nil
}

func (tx *fakeSignalTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (tx *fakeSignalTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, errors.New("unexpected Prepare")
}

func (tx *fakeSignalTx) Exec(_ context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "pg_advisory_xact_lock") && len(arguments) == 1 {
		if scope, ok := arguments[0].(string); ok {
			tx.locks = append(tx.locks, scope)
		}
	}
	return pgconn.CommandTag{}, nil
}

func (tx *fakeSignalTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (tx *fakeSignalTx) QueryRow(_ context.Context, sql string, args ...any) pgx.Row {
	tx.queries = append(tx.queries, fakeSignalQuery{SQL: sql, Args: append([]any(nil), args...)})
	symbol, _ := args[0].(string)
	maxLayerNumber := 0
	for _, signal := range tx.masterSignals {
		if signal.Symbol == symbol && signal.LayerNumber > maxLayerNumber {
			maxLayerNumber = signal.LayerNumber
		}
	}
	for _, layer := range tx.layers {
		if layer.Symbol == symbol && layer.LayerNumber > maxLayerNumber {
			maxLayerNumber = layer.LayerNumber
		}
	}
	return fakeSignalRow{value: maxLayerNumber + 1}
}

func (tx *fakeSignalTx) Conn() *pgx.Conn {
	return nil
}

type fakeSignalRow struct {
	value int
	err   error
}

func (row fakeSignalRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(dest) != 1 {
		return errors.New("unexpected scan destination count")
	}
	target, ok := dest[0].(*int)
	if !ok {
		return errors.New("unexpected scan destination type")
	}
	*target = row.value
	return nil
}

func TestReserveNextBuyLayerNumberIsScopedPerSymbol(t *testing.T) {
	t.Parallel()

	tx := &fakeSignalTx{
		masterSignals: []layerNumberSeed{
			{Symbol: "BTC/USDT", LayerNumber: 1},
			{Symbol: "ETH/USDT", LayerNumber: 1},
			{Symbol: "BTC/USDT", LayerNumber: 2},
		},
		layers: []layerNumberSeed{
			{Symbol: "SOL/USDT", LayerNumber: 9},
			{Symbol: "ETH/USDT", LayerNumber: 3},
			{Symbol: "BTC/USDT", LayerNumber: 8},
		},
	}

	tests := []struct {
		symbol string
		want   int
	}{
		{symbol: "BTC/USDT", want: 9},
		{symbol: "ETH/USDT", want: 4},
		{symbol: "AVAX/USDT", want: 1},
	}

	for _, tt := range tests {
		got, err := reserveNextBuyLayerNumber(context.Background(), tx, tt.symbol)
		if err != nil {
			t.Fatalf("reserve layer for %s: %v", tt.symbol, err)
		}
		if got != tt.want {
			t.Fatalf("expected next %s layer %d, got %d", tt.symbol, tt.want, got)
		}
	}
}

func TestReserveNextBuyLayerNumberUsesPairScopedLockAndQuery(t *testing.T) {
	t.Parallel()

	tx := &fakeSignalTx{}
	layerNumber, err := reserveNextBuyLayerNumber(context.Background(), tx, " eth/usdt ")
	if err != nil {
		t.Fatalf("reserve layer: %v", err)
	}
	if layerNumber != 1 {
		t.Fatalf("expected first layer, got %d", layerNumber)
	}
	if len(tx.locks) != 1 || tx.locks[0] != "buy-layer-number:ETH/USDT" {
		t.Fatalf("expected pair-scoped lock, got %#v", tx.locks)
	}
	if len(tx.queries) != 1 {
		t.Fatalf("expected one layer query, got %d", len(tx.queries))
	}
	if len(tx.queries[0].Args) != 1 || tx.queries[0].Args[0] != "ETH/USDT" {
		t.Fatalf("expected normalized symbol query arg, got %#v", tx.queries[0].Args)
	}
	sql := tx.queries[0].SQL
	if !strings.Contains(sql, "FROM master_signals") || !strings.Contains(sql, "FROM layers") {
		t.Fatalf("expected query to inspect master_signals and layers: %s", sql)
	}
	if strings.Count(sql, "symbol = $1") < 2 {
		t.Fatalf("expected symbol-scoped query filters, got: %s", sql)
	}
}
