package xlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DebugLevel = zap.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zap.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zap.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-Level logs.
	ErrorLevel = zap.ErrorLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zap.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zap.FatalLevel
)

type (
	Func func(string,...zap.Field)
	Field = zap.Field
	Level = zapcore.Level
	Logger struct {
		desugar *zap.Logger
		lv *zap.AtomicLevel
		config Config
	}
)
