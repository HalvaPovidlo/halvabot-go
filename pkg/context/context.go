package context

import (
	"context"

	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

type key string

const (
	loggerKey key = "loggerKey"
)

type Context struct {
	context.Context
}

var Background = context.Background

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
func (ctx Context) LoggerFromContext() *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return zap.NewLogger()
}
