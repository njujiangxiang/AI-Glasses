package monitoring

import (
	"strings"
	"testing"
)

func TestHubRecentAfterIDAndLimitCap(t *testing.T) {
	hub := NewHub(WithMaxEntries(600))
	for i := 0; i < 600; i++ {
		hub.Append("LOG", "test", "line")
	}

	result := hub.Recent(999999, 100)
	if len(result.Entries) != MaxLimit {
		t.Fatalf("expected capped result length %d, got %d", MaxLimit, len(result.Entries))
	}
	if result.Entries[0].ID <= 100 {
		t.Fatalf("expected entries after id 100, got first id %d", result.Entries[0].ID)
	}
}

func TestHubDetectsGapAfterEviction(t *testing.T) {
	hub := NewHub(WithMaxEntries(3))
	for i := 0; i < 5; i++ {
		hub.Append("LOG", "test", "line")
	}

	result := hub.Recent(200, 1)
	if !result.Gap {
		t.Fatal("expected gap for stale after_id")
	}
	if result.Skipped != 1 {
		t.Fatalf("expected skipped=1, got %d", result.Skipped)
	}
	if result.OldestID != 3 || result.NewestID != 5 {
		t.Fatalf("expected oldest/newest 3/5, got %d/%d", result.OldestID, result.NewestID)
	}
}

func TestHubSanitizesAndTruncates(t *testing.T) {
	hub := NewHub(WithMaxEntryBytes(12))
	entry := hub.Append("", "", "\x1b[31mpassword=secret-value\x1b[0m and more text")
	if strings.Contains(entry.Message, "secret-value") {
		t.Fatalf("expected secret redacted, got %q", entry.Message)
	}
	if strings.Contains(entry.Message, "\x1b") {
		t.Fatalf("expected ANSI stripped, got %q", entry.Message)
	}
	if len(entry.Message) > 15 {
		t.Fatalf("expected truncated message, got %q", entry.Message)
	}
}

func TestHubStreamIDChangesPerHub(t *testing.T) {
	first := NewHub()
	second := NewHub()
	if first.StreamID() == "" {
		t.Fatal("expected stream id")
	}
	if first.StreamID() == second.StreamID() {
		t.Fatal("expected separate hubs to have different stream ids")
	}
}
