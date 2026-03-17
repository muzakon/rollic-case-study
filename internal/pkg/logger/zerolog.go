package logger

import (
	"os"
	"time"

	"server/internal/config"

	"github.com/rs/zerolog"
)

func New(env config.Env) *zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339

	var logger zerolog.Logger

	if env == config.EnvProduction {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		logger = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Logger()
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		logger = zerolog.New(consoleWriter).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	return &logger
}
