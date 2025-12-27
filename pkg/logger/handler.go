package logger

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/system"
)

// LogStore defines the interface for persisting logs.
type LogStore interface {
	Insert(ctx context.Context, entry *system.LogEntry) error
}

// MultiHandler writes logs to multiple handlers and optionally persists to a store.
// Call Stop() during shutdown to flush pending log entries.
type MultiHandler struct {
	handlers []slog.Handler
	store    LogStore
	attrs    []slog.Attr
	groups   []string

	// Worker pool for async log persistence (pointers to allow sharing with WithAttrs/WithGroup)
	logChan  chan *system.LogEntry
	wg       *sync.WaitGroup
	stopOnce *sync.Once
	stopped  chan struct{}
}

const logBufferSize = 1000

// NewMultiHandler creates a new MultiHandler. If store is non-nil, logs are
// persisted asynchronously. Call Stop() during shutdown to flush pending logs.
func NewMultiHandler(handlers []slog.Handler, store LogStore) *MultiHandler {
	h := &MultiHandler{
		handlers: handlers,
		store:    store,
		logChan:  make(chan *system.LogEntry, logBufferSize),
		wg:       &sync.WaitGroup{},
		stopOnce: &sync.Once{},
		stopped:  make(chan struct{}),
	}

	if store != nil {
		h.wg.Add(1)
		go h.worker()
	}

	return h
}

// Stop flushes pending log entries and stops the worker. Safe to call multiple times.
func (h *MultiHandler) Stop() {
	h.stopOnce.Do(func() {
		close(h.stopped)
		h.wg.Wait()
	})
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}

	if h.store != nil {
		entry := h.buildEntry(r)
		select {
		case h.logChan <- entry:
			// Queued for persistence
		default:
			// Buffer full, drop log entry to avoid blocking
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{
		handlers: newHandlers,
		store:    h.store,
		attrs:    append(h.attrs, attrs...),
		groups:   h.groups,
		logChan:  h.logChan,
		wg:       h.wg,
		stopped:  h.stopped,
	}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{
		handlers: newHandlers,
		store:    h.store,
		attrs:    h.attrs,
		groups:   append(h.groups, name),
		logChan:  h.logChan,
		wg:       h.wg,
		stopped:  h.stopped,
	}
}

// worker processes log entries from the channel until stopped.
func (h *MultiHandler) worker() {
	defer h.wg.Done()

	for {
		select {
		case entry := <-h.logChan:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = h.store.Insert(ctx, entry)
			cancel()
		case <-h.stopped:
			// Drain remaining entries before exiting
			for {
				select {
				case entry := <-h.logChan:
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					_ = h.store.Insert(ctx, entry)
					cancel()
				default:
					return
				}
			}
		}
	}
}

// buildEntry creates a LogEntry from an slog.Record.
func (h *MultiHandler) buildEntry(r slog.Record) *system.LogEntry {
	entry := &system.LogEntry{
		Level:     levelToString(r.Level),
		Message:   r.Message,
		Timestamp: r.Time,
		Attrs:     make(map[string]any),
	}

	for _, attr := range h.attrs {
		h.addAttr(entry, attr)
	}

	r.Attrs(func(a slog.Attr) bool {
		h.addAttr(entry, a)
		return true
	})

	return entry
}

func (h *MultiHandler) addAttr(entry *system.LogEntry, attr slog.Attr) {
	switch attr.Key {
	case "request_id":
		entry.RequestID = attr.Value.String()
	case "user_id":
		entry.UserID = attr.Value.String()
	case "source", "handler":
		entry.Source = attr.Value.String()
	default:
		entry.Attrs[attr.Key] = attr.Value.Any()
	}
}

func levelToString(level slog.Level) string {
	switch level {
	case LevelTrace:
		return "TRACE"
	case slog.LevelDebug:
		return "DEBUG"
	case slog.LevelInfo:
		return "INFO"
	case slog.LevelWarn:
		return "WARN"
	case slog.LevelError:
		return "ERROR"
	case LevelCritical:
		return "CRITICAL"
	default:
		return "INFO"
	}
}
