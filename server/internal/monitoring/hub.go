// Package monitoring 提供后台实时监控使用的进程内日志缓冲区。它只保存最近日志，
// 不承担持久化或多实例聚合职责。
package monitoring

import (
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	DefaultLimit         = 200
	MaxLimit             = 500
	DefaultMaxEntries    = 1000
	DefaultMaxEntryBytes = 4096
)

type LogEntry struct {
	ID      uint64    `json:"id"`
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Source  string    `json:"source"`
	Message string    `json:"message"`
}

type RecentResult struct {
	StreamID string     `json:"stream_id"`
	Entries  []LogEntry `json:"entries"`
	Gap      bool       `json:"gap"`
	Skipped  uint64     `json:"skipped,omitempty"`
	OldestID uint64     `json:"oldest_id"`
	NewestID uint64     `json:"newest_id"`
}

type Hub struct {
	mu            sync.RWMutex
	streamID      string
	nextID        uint64
	maxEntries    int
	maxEntryBytes int
	buffer        []LogEntry
}

type Option func(*Hub)

func WithMaxEntries(n int) Option {
	return func(h *Hub) {
		if n > 0 {
			h.maxEntries = n
		}
	}
}

func WithMaxEntryBytes(n int) Option {
	return func(h *Hub) {
		if n > 0 {
			h.maxEntryBytes = n
		}
	}
}

func NewHub(opts ...Option) *Hub {
	h := &Hub{
		streamID:      uuid.NewString(),
		nextID:        1,
		maxEntries:    DefaultMaxEntries,
		maxEntryBytes: DefaultMaxEntryBytes,
		buffer:        make([]LogEntry, 0, DefaultMaxEntries),
	}
	for _, opt := range opts {
		opt(h)
	}
	if h.maxEntries <= 0 {
		h.maxEntries = DefaultMaxEntries
	}
	if h.maxEntryBytes <= 0 {
		h.maxEntryBytes = DefaultMaxEntryBytes
	}
	if cap(h.buffer) != h.maxEntries {
		h.buffer = make([]LogEntry, 0, h.maxEntries)
	}
	return h
}

func (h *Hub) StreamID() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.streamID
}

func (h *Hub) Append(level, source, message string) LogEntry {
	level = strings.TrimSpace(level)
	if level == "" {
		level = "LOG"
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = "app"
	}
	message = truncateUTF8(Sanitize(message), h.maxEntryBytes)

	h.mu.Lock()
	defer h.mu.Unlock()
	entry := LogEntry{
		ID:      h.nextID,
		Time:    time.Now().UTC(),
		Level:   level,
		Source:  source,
		Message: message,
	}
	h.nextID++
	if len(h.buffer) == h.maxEntries {
		copy(h.buffer, h.buffer[1:])
		h.buffer[len(h.buffer)-1] = entry
		return entry
	}
	h.buffer = append(h.buffer, entry)
	return entry
}

func (h *Hub) Recent(limit int, afterID uint64) RecentResult {
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	result := RecentResult{StreamID: h.streamID, Entries: []LogEntry{}}
	if len(h.buffer) == 0 {
		return result
	}

	oldestID := h.buffer[0].ID
	newestID := h.buffer[len(h.buffer)-1].ID
	result.OldestID = oldestID
	result.NewestID = newestID
	if afterID > 0 && afterID < oldestID-1 {
		result.Gap = true
		result.Skipped = oldestID - afterID - 1
	}

	start := 0
	if afterID > 0 {
		for i, entry := range h.buffer {
			if entry.ID > afterID {
				start = i
				break
			}
			if i == len(h.buffer)-1 {
				return result
			}
		}
	}
	if len(h.buffer)-start > limit {
		start = len(h.buffer) - limit
	}
	result.Entries = append(result.Entries, h.buffer[start:]...)
	return result
}

func truncateUTF8(value string, maxBytes int) string {
	if maxBytes <= 0 || len(value) <= maxBytes {
		return value
	}
	truncated := value[:maxBytes]
	for !utf8.ValidString(truncated) && len(truncated) > 0 {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated + "…"
}
