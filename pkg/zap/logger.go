package zap

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger(debug bool) Logger {
	config := zap.NewDevelopmentConfig()
	if debug {
		config.Level.SetLevel(zapcore.DebugLevel)
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config.Level.SetLevel(zapcore.InfoLevel)
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	zapLogger, _ := config.Build()
	return Logger{
		SugaredLogger: zapLogger.Sugar(),
	}
}
