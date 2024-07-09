package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)


func CreateLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMilli}
	return zerolog.New(output).With().Timestamp().Logger()
}