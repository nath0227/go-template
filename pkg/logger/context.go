package logger

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TIDKey is the canonical key string for trace IDs — used as the Zap field name,
// the context key value, and the Kafka message header key.
const TIDKey = "tid"

type contextKey string

const tidCtxKey contextKey = TIDKey

// WithTID stores a trace ID in the context.
func WithTID(ctx context.Context, tid string) context.Context {
	return context.WithValue(ctx, tidCtxKey, tid)
}

// TIDFromContext returns the trace ID stored in ctx, or empty string.
func TIDFromContext(ctx context.Context) string {
	tid, _ := ctx.Value(tidCtxKey).(string)
	return tid
}

// FromContext returns a logger with the tid field injected from ctx.
// Falls back to zap.L() when no tid is present.
func FromContext(ctx context.Context) *zap.Logger {
	tid := TIDFromContext(ctx)
	if tid == "" {
		return zap.L()
	}
	return zap.L().With(zap.String(TIDKey, tid))
}

// NewTID generates a trace ID for ctx.
// If an active OpenTelemetry span exists with a valid trace ID, that is used.
// Otherwise, a new UUID v4 is generated.
func NewTID(ctx context.Context) string {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return uuid.New().String()
}
