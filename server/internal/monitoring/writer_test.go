package monitoring

import "testing"

func TestWriterFramesLinesAndBuffersPartialChunks(t *testing.T) {
	hub := NewHub()
	writer := NewWriter(hub, "LOG", "stdlib")

	if n, err := writer.Write([]byte("first")); err != nil || n != len("first") {
		t.Fatalf("unexpected write result n=%d err=%v", n, err)
	}
	if got := hub.Recent(10, 0); len(got.Entries) != 0 {
		t.Fatalf("expected no entry for partial line, got %d", len(got.Entries))
	}
	_, _ = writer.Write([]byte(" line\nsecond line\n"))

	result := hub.Recent(10, 0)
	if len(result.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Entries))
	}
	if result.Entries[0].Message != "first line" || result.Entries[1].Message != "second line" {
		t.Fatalf("unexpected messages: %#v", result.Entries)
	}
}
