// Package ctxlog provides context-aware structured logging helpers for Go's log/slog.
package ctxlog

import (
	"context"
	"log/slog"
)

// contextKey is an unexported type used for context keys to avoid collisions.
type contextKey int

const (
	// fieldsKey stores []any key-value pairs in the context.
	fieldsKey contextKey = iota
	// attrsKey stores []slog.Attr in the context.
	attrsKey
	// requestIDKey stores the request ID string in the context.
	requestIDKey
	// correlationIDKey stores the correlation ID string in the context.
	correlationIDKey
	// traceIDKey stores the trace ID string in the context.
	traceIDKey
)

// With attaches slog key-value pairs to the context. The args should be
// alternating key-value pairs as accepted by slog.Logger.With. Multiple calls
// to With accumulate fields.
func With(ctx context.Context, args ...any) context.Context {
	existing, _ := ctx.Value(fieldsKey).([]any)
	merged := make([]any, len(existing)+len(args))
	copy(merged, existing)
	copy(merged[len(existing):], args)
	return context.WithValue(ctx, fieldsKey, merged)
}

// WithAttrs attaches typed slog.Attr values to the context. Multiple calls
// to WithAttrs accumulate attributes.
func WithAttrs(ctx context.Context, attrs ...slog.Attr) context.Context {
	existing, _ := ctx.Value(attrsKey).([]slog.Attr)
	merged := make([]slog.Attr, len(existing)+len(attrs))
	copy(merged, existing)
	copy(merged[len(existing):], attrs)
	return context.WithValue(ctx, attrsKey, merged)
}

// WithRequestID attaches a request ID to the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID extracts the request ID from the context. Returns an empty string
// if no request ID has been set.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// WithCorrelationID attaches a correlation ID to the context. Correlation IDs
// are used to group related requests across services.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

// CorrelationID extracts the correlation ID from the context. Returns an empty
// string if no correlation ID has been set.
func CorrelationID(ctx context.Context) string {
	id, _ := ctx.Value(correlationIDKey).(string)
	return id
}

// WithTraceID attaches a trace ID to the context. Trace IDs are used for
// distributed tracing across service boundaries.
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

// TraceID extracts the trace ID from the context. Returns an empty string if
// no trace ID has been set.
func TraceID(ctx context.Context) string {
	id, _ := ctx.Value(traceIDKey).(string)
	return id
}

// Fields extracts all typed slog.Attr values stored in the context via
// WithAttrs. Returns nil if no attributes have been set.
func Fields(ctx context.Context) []slog.Attr {
	attrs, _ := ctx.Value(attrsKey).([]slog.Attr)
	if len(attrs) == 0 {
		return nil
	}
	result := make([]slog.Attr, len(attrs))
	copy(result, attrs)
	return result
}

// Logger returns slog.Default() enriched with all fields, attributes, and
// request ID stored in the context.
func Logger(ctx context.Context) *slog.Logger {
	return From(ctx, slog.Default())
}

// From enriches the provided logger with all fields, attributes, and request
// ID stored in the context.
func From(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if fields, ok := ctx.Value(fieldsKey).([]any); ok && len(fields) > 0 {
		logger = logger.With(fields...)
	}

	if attrs, ok := ctx.Value(attrsKey).([]slog.Attr); ok && len(attrs) > 0 {
		args := make([]any, len(attrs))
		for i, a := range attrs {
			args[i] = a
		}
		logger = logger.With(args...)
	}

	if id := RequestID(ctx); id != "" {
		logger = logger.With("request_id", id)
	}

	if id := CorrelationID(ctx); id != "" {
		logger = logger.With("correlation_id", id)
	}

	if id := TraceID(ctx); id != "" {
		logger = logger.With("trace_id", id)
	}

	return logger
}
