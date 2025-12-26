package logger

import (
	"context"
	"log/slog"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/system"
)

// LogStore defines the interface for persisting logs.
type LogStore interface {
	Insert(ctx context.Context, entry *system.LogEntry) error
}

// MultiHandler writes logs to multiple handlers.
type MultiHandler struct {
	handlers []slog.Handler
	store    LogStore
	attrs    []slog.Attr
	groups   []string
}

func NewMultiHandler(handlers []slog.Handler, store LogStore) *MultiHandler {
	return &MultiHandler{handlers: handlers, store: store}
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
		go h.persistLog(r)
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
	}
}

func (h *MultiHandler) persistLog(r slog.Record) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = h.store.Insert(ctx, entry)
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
