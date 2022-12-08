package simple_router

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func CreateRootLogger() zerolog.Logger {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	var level = zerolog.DebugLevel
	return logger.Level(level).With().Timestamp().Logger()
}
