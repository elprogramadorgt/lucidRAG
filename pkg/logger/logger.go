package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// Custom log levels.
const (
	LevelTrace    = slog.Level(-8) // More verbose than DEBUG
	LevelCritical = slog.Level(12) // More severe than ERROR
)

// ContextKey is a type-safe key for context values.
type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	UserIDKey    ContextKey = "user_id"
)

// Logger wraps slog.Logger with additional features like custom levels and log persistence.
type Logger struct {
	log     *slog.Logger
	level   *slog.LevelVar
	handler *MultiHandler // nil if no store configured
}

type Options struct {
	Level     string
	JSON      bool
	AddSource bool
	Store     LogStore
}

// New creates a new Logger with the given options.
func New(opts ...Options) *Logger {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	levelVar := &slog.LevelVar{}
	levelVar.Set(parseLevel(opt.Level))

	handlerOpts := &slog.HandlerOptions{
		Level:     levelVar,
		AddSource: opt.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Replace custom level names
			if a.Key == slog.LevelKey {
				level := a.Value.Any().(slog.Level)
				switch level {
				case LevelTrace:
					a.Value = slog.StringValue("TRACE")
				case LevelCritical:
					a.Value = slog.StringValue("CRITICAL")
				}
			}
			return a
		},
	}

	var stdoutHandler slog.Handler
	if opt.JSON {
		stdoutHandler = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		stdoutHandler = slog.NewTextHandler(os.Stdout, handlerOpts)
	}

	var handler slog.Handler
	var multiHandler *MultiHandler
	if opt.Store != nil {
		multiHandler = NewMultiHandler([]slog.Handler{stdoutHandler}, opt.Store)
		handler = multiHandler
	} else {
		handler = stdoutHandler
	}

	return &Logger{
		log:     slog.New(handler),
		level:   levelVar,
		handler: multiHandler,
	}
}

// Stop flushes pending log entries and releases resources.
// Call this during application shutdown. Safe to call multiple times or on nil.
func (l *Logger) Stop() {
	if l != nil && l.handler != nil {
		l.handler.Stop()
	}
}

// parseLevel converts a string level to slog.Level (case-insensitive).
// Levels (lowest to highest): trace, debug, info, warn, error, critical
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return LevelTrace
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "critical", "fatal":
		return LevelCritical
	default:
		return slog.LevelInfo
	}
}

// SetLevel changes the log level at runtime.
func (l *Logger) SetLevel(level string) {
	l.level.Set(parseLevel(level))
}

// GetLevel returns the current log level as a string.
func (l *Logger) GetLevel() string {
	switch l.level.Level() {
	case LevelTrace:
		return "trace"
	case slog.LevelDebug:
		return "debug"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	case LevelCritical:
		return "critical"
	default:
		return "info"
	}
}

// Trace logs at TRACE level (most verbose).
func (l *Logger) Trace(msg string, args ...any) {
	l.log.Log(context.Background(), LevelTrace, msg, args...)
}

// Debug logs at DEBUG level.
func (l *Logger) Debug(msg string, args ...any) {
	l.log.Debug(msg, args...)
}

// Info logs at INFO level.
func (l *Logger) Info(msg string, args ...any) {
	l.log.Info(msg, args...)
}

// Warn logs at WARN level.
func (l *Logger) Warn(msg string, args ...any) {
	l.log.Warn(msg, args...)
}

// Error logs at ERROR level.
func (l *Logger) Error(msg string, args ...any) {
	l.log.Error(msg, args...)
}

// Critical logs at CRITICAL level (most severe).
func (l *Logger) Critical(msg string, args ...any) {
	l.log.Log(context.Background(), LevelCritical, msg, args...)
}

// TraceContext logs at TRACE level with context.
func (l *Logger) TraceContext(ctx context.Context, msg string, args ...any) {
	l.log.Log(ctx, LevelTrace, msg, args...)
}

// DebugContext logs at DEBUG level with context.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log.DebugContext(ctx, msg, args...)
}

// InfoContext logs at INFO level with context.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log.InfoContext(ctx, msg, args...)
}

// WarnContext logs at WARN level with context.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log.WarnContext(ctx, msg, args...)
}

// ErrorContext logs at ERROR level with context.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.log.ErrorContext(ctx, msg, args...)
}

// CriticalContext logs at CRITICAL level with context.
func (l *Logger) CriticalContext(ctx context.Context, msg string, args ...any) {
	l.log.Log(ctx, LevelCritical, msg, args...)
}

// With returns a new Logger with the given attributes.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		log:     l.log.With(args...),
		level:   l.level,
		handler: l.handler,
	}
}

// WithGroup returns a new Logger with the given group name.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		log:     l.log.WithGroup(name),
		level:   l.level,
		handler: l.handler,
	}
}

// WithError returns a new Logger with the error as an attribute.
func (l *Logger) WithError(err error) *Logger {
	return l.With("error", err)
}

// WithContext extracts known context values and returns a new Logger with them.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		logger = logger.With("request_id", requestID)
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		logger = logger.With("user_id", userID)
	}
	return logger
}
