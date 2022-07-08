package contexts

import (
	"context"

	"github.com/HalvaPovidlo/discordBotGo/pkg/zap"
)

type key string

const (
	loggerKey key = "loggerKey"
)

// Context TODO write normal context logger
type Context struct {
	context.Context
}

var Background = context.Background

func WithLogger(ctx context.Context, logger zap.Logger) (Context, context.CancelFunc) {
	ctx, f := context.WithCancel(context.WithValue(ctx, loggerKey, logger))
	return Context{ctx}, f
}
func (ctx Context) LoggerFromContext() zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(zap.Logger); ok {
		return logger
	}
	return zap.NewLogger(true)
}
