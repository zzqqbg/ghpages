package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const defaultRetain = 30 * 24 * time.Hour

// WeeklyWriter appends to mobagent-YYYY-Www.log and deletes files older than retain.
type WeeklyWriter struct {
	dir    string
	prefix string
	retain time.Duration

	mu   sync.Mutex
	week string
	file *os.File
}

func NewWeeklyWriter(dir, prefix string, retain time.Duration) *WeeklyWriter {
	if prefix == "" {
		prefix = "mobagent"
	}
	if retain <= 0 {
		retain = defaultRetain
	}
	return &WeeklyWriter{dir: dir, prefix: prefix, retain: retain}
}

func weekKey(t time.Time) string {
	y, w := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}

func (w *WeeklyWriter) path(t time.Time) string {
	y, ww := t.ISOWeek()
	name := fmt.Sprintf("%s-%d-W%02d.log", w.prefix, y, ww)
	return filepath.Join(w.dir, name)
}

func (w *WeeklyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	key := weekKey(time.Now())
	if w.file == nil || w.week != key {
		if err := w.rotate(key); err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

func (w *WeeklyWriter) rotate(key string) error {
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}
	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(w.path(time.Now()), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.file = f
	w.week = key
	go w.cleanup()
	return nil
}

func (w *WeeklyWriter) cleanup() {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-w.retain)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(w.dir, e.Name()))
		}
	}
}

func (w *WeeklyWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}
