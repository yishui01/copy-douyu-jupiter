package xlog

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

type Config struct {
	Dir           string
	Name          string
	Level         string
	Fields        []zap.Field
	AddCaller     bool
	Prefix        string
	MaxSize       int
	MaxAge        int
	MaxBackup     int
	Interval      time.Duration
	CallerSkip    int
	Async         bool
	Queue         bool
	QueueSleep    time.Duration
	Core          zapcore.Core
	Debug         bool
	EncoderConfig *zapcore.EncoderConfig
	configKey     string
}

func (config *Config) Filename() string {
	return fmt.Sprintf("%s/%s", config.Dir, config.Name)
}

func RawConfig(key string) *Config {
	var config = DefaultConfig()
	if err := conf.Unmarshalkey(key, &config); err != nil {
		panic(err)
	}
	config.configKey = key
	return config
}

func StdConfig(name string) *Config {
	return RawConfig("jupiter.logger." + name)
}

func DefaultConfig() *Config {
	return &Config{
		Name:          "default.log",
		Dir:           ".",
		Level:         "info",
		MaxSize:       500, // 500M
		MaxAge:        1,   // 1 day
		MaxBackup:     10,  // 10 backup
		Interval:      24 * time.Hour,
		CallerSkip:    1,
		AddCaller:     false,
		Async:         true,
		Queue:         false,
		QueueSleep:    100 * time.Millisecond,
		EncoderConfig: DefaultZapConfig(),
	}
}
