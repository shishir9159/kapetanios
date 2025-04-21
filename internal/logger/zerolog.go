package logger

import (
	"github.com/rs/zerolog"
	"io"
	"os"
	"time"
)

type ZerologBuilder struct {
	level      zerolog.Level
	output     io.Writer
	timeFormat string
	pretty     bool
}

func NewZerologBuilder() *ZerologBuilder {
	return &ZerologBuilder{
		level:      zerolog.InfoLevel,
		output:     os.Stdout,
		timeFormat: time.RFC3339,
		pretty:     false,
	}
}

func (b *ZerologBuilder) SetLevel(level zerolog.Level) *ZerologBuilder {
	b.level = level
	return b
}

func (b *ZerologBuilder) SetOutput(output io.Writer) *ZerologBuilder {
	b.output = output
	return b
}

func (b *ZerologBuilder) SetTimeFormat(format string) *ZerologBuilder {
	b.timeFormat = format
	return b
}

func (b *ZerologBuilder) SetPretty(pretty bool) *ZerologBuilder {
	b.pretty = pretty
	return b
}

func (b *ZerologBuilder) Build() zerolog.Logger {
	zerolog.SetGlobalLevel(b.level)
	zerolog.TimeFieldFormat = b.timeFormat

	var output io.Writer = b.output
	if b.pretty {
		output = zerolog.ConsoleWriter{
			Out:        b.output,
			TimeFormat: b.timeFormat,
		}
	}

	return zerolog.New(output).With().Timestamp().Logger()
}
