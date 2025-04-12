package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type logger struct {
}

func zeroLogger(l *logger) {
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("[%s]", i))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s |", i)
		},
		FormatCaller: func(i interface{}) string {
			return filepath.Base(fmt.Sprintf("%s", i))
		},
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}).With().Timestamp().Caller().Stack().Logger()

}
