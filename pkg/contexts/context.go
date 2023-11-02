package contexts

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/HalvaPovidlo/halvabot-go/pkg/log"
)

type contextKey string

const (
	loggerKey  contextKey = "logger"
	traceIDKey contextKey = "trace_id"

	defaultStringValue = "other"
)

func WithValues(parent context.Context, logger *zap.Logger, traceID string) context.Context {
	if traceID == "" {
		traceID = uuid.New().String()
	}
	ctx := WithTraceID(parent, traceID)
	ctx = WithLogger(ctx, logger.With(zap.String("traceID", traceID)))
	return ctx
}

func WithCommandValues(parent context.Context, name string, logger *zap.Logger, traceID string) context.Context {
	if traceID == "" {
		traceID = uuid.New().String()
	}
	ctx := WithTraceID(parent, traceID)
	ctx = WithLogger(ctx, logger.With(zap.String("traceID", traceID), zap.String("command", name)))
	return ctx
}

func WithLogger(parent context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(parent, loggerKey, logger)
}

func GetLogger(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(loggerKey).(*zap.Logger)
	if ok && logger != nil {
		return logger
	}
	return log.NewLogger(false).With(zap.String("logger", "default"))
}

func WithTraceID(parent context.Context, id string) context.Context {
	return context.WithValue(parent, traceIDKey, id)
}

func GetTraceID(ctx context.Context) string {
	traceID, ok := ctx.Value(traceIDKey).(string)
	if ok && traceID != "" {
		return traceID
	}

	return defaultStringValue
}
