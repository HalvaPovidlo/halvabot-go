package zap

import "go.uber.org/zap"

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger() *Logger {
	zapLogger, _ := zap.NewProduction()
	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}
}
