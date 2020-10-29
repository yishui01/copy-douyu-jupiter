package xlog

import (
	"copy/pkg/defers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

const (
	DebugLevel = zap.DebugLevel
	InfoLevel  = zap.InfoLevel
	WarnLevel  = zap.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-Level logs.
	ErrorLevel = zap.ErrorLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zap.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zap.FatalLevel
)

type (
	Func   func(string, ...zap.Field)
	Field  = zap.Field
	Level  = zapcore.Level
	Logger struct {
		desugar *zap.Logger
		lv      *zap.AtomicLevel
		config  Config
		sugar   *zap.SugaredLogger
	}
)

var (
	String = zap.String
	Any    = zap.Any
	Int64  = zap.Int64
	// Int ...
	Int = zap.Int
	// Int32 ...
	Int32 = zap.Int32
	// Uint ...
	Uint = zap.Uint
	// Duration ...
	Duration = zap.Duration
	// Durationp ...
	Durationp = zap.Durationp
	// Object ...
	Object = zap.Object
	// Namespace ...
	Namespace = zap.Namespace
	// Reflect ...
	Reflect = zap.Reflect
	// Skip ...
	Skip = zap.Skip()
	// ByteString ...
	ByteString = zap.ByteString
)

func newLogger(config *Config) *Logger {
	zapOptions := make([]zap.Option, 0)
	zapOptions = append(zapOptions, zap.AddStacktrace(zap.DPanicLevel))
	if config.AddCaller {
		zapOptions = append(zapOptions, zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	}
	if len(config.Fields) > 0 {
		zapOptions = append(zapOptions, zap.Fields(config.Fields...))
	}

	var ws zapcore.WriteSyncer
	if config.Debug {
		ws = os.Stdout
	} else {
		ws = zapcore.AddSync(newRotate(config))
	}

	if config.Async {
		var close CloseFunc
		ws, close = Buffer(ws, defaultBufferSize, defaultFlushInterval)

		defers.Register(close)
	}

	lv := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	if err := lv.UnmarshalText([]byte(config.Level)); err != nil {
		panic(err)
	}
	encoderConfig := *config.EncoderConfig
	core := config.Core
	if core == nil {
		core = zapcore.NewCore(
			func() zapcore.Encoder {
				if config.Debug {
					return zapcore.NewConsoleEncoder(encoderConfig)
				}
				return zapcore.NewJSONEncoder(encoderConfig)
			}(),
			ws,
			lv,
		)
	}

	zapLogger := zap.New(
		core,
		zapOptions, ...
	)
	return &Logger{
		desugar: zapLogger,
		lv:      &lv,
		config:  *config,
		sugar:   zapLogger.Sugar(),
	}
}
