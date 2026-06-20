package monitoring

import (
	"bytes"
	"sync"
)

const defaultMaxLineBytes = 8192

type Writer struct {
	hub          *Hub
	level        string
	source       string
	mu           sync.Mutex
	partial      []byte
	maxLineBytes int
}

func NewWriter(hub *Hub, level, source string) *Writer {
	return &Writer{hub: hub, level: level, source: source, maxLineBytes: defaultMaxLineBytes}
}

func (w *Writer) Write(p []byte) (int, error) {
	if w == nil || w.hub == nil {
		return len(p), nil
	}
	w.mu.Lock()
	defer w.mu.Unlock()

	data := append(w.partial, p...)
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines[:len(lines)-1] {
		w.appendLine(line)
	}
	w.partial = append(w.partial[:0], lines[len(lines)-1]...)
	if len(w.partial) > w.maxLineBytes {
		w.appendLine(w.partial[:w.maxLineBytes])
		w.partial = w.partial[:0]
	}
	return len(p), nil
}

func (w *Writer) appendLine(line []byte) {
	line = bytes.TrimRight(line, "\r")
	if len(line) == 0 {
		return
	}
	w.hub.Append(w.level, w.source, string(line))
}
