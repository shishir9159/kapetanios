package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLoggerBuilder struct {
	development bool
	level       zapcore.Level
	outputs     []string
	encoding    string
}

func NewZapLoggerBuilder() *ZapLoggerBuilder {
	return &ZapLoggerBuilder{
		development: false,
		level:       zapcore.InfoLevel,
		outputs:     []string{"stdout"},
		encoding:    "json",
	}
}

func (b *ZapLoggerBuilder) Development(isDev bool) *ZapLoggerBuilder {
	b.development = isDev
	return b
}

func (b *ZapLoggerBuilder) SetLevel(level zapcore.Level) *ZapLoggerBuilder {
	b.level = level
	return b
}

func (b *ZapLoggerBuilder) SetOutputs(outputs []string) *ZapLoggerBuilder {
	b.outputs = outputs
	return b
}

// SetEncoding sets the encoding (json or console)
func (b *ZapLoggerBuilder) SetEncoding(encoding string) *ZapLoggerBuilder {
	b.encoding = encoding
	return b
}

func (b *ZapLoggerBuilder) Build() (*zap.Logger, error) {
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(b.level),
		Development: b.development,
		Encoding:    b.encoding,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      b.outputs,
		ErrorOutputPaths: []string{"stderr"},
	}

	return config.Build()
}
