package store

import (
	"testing"
	"time"
)

func TestParseExecutionExecutedAtAcceptsRFC3339(t *testing.T) {
	t.Parallel()

	got, err := parseExecutionExecutedAt("2026-07-29T19:43:44.341817464Z")
	if err != nil {
		t.Fatalf("expected RFC3339 timestamp to parse: %v", err)
	}

	want := time.Date(2026, 7, 29, 19, 43, 44, 341817464, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want.Format(time.RFC3339Nano), got.Format(time.RFC3339Nano))
	}
}

func TestParseExecutionExecutedAtAcceptsLegacyUnixTimestamp(t *testing.T) {
	t.Parallel()

	got, err := parseExecutionExecutedAt("1785354224.341817464Z")
	if err != nil {
		t.Fatalf("expected legacy Unix timestamp to parse: %v", err)
	}

	want := time.Unix(1785354224, 341817464).UTC()
	if !got.Equal(want) {
		t.Fatalf("expected %s, got %s", want.Format(time.RFC3339Nano), got.Format(time.RFC3339Nano))
	}
}
