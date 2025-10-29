package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger(level, format string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set output format - Azure Log Analytics needs stdout for JSON logs
	if strings.ToLower(format) == "console" && os.Getenv("ENV") == "development" {
		// Only use console format in development
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	} else {
		// Production: always JSON to stdout for Azure Log Analytics
		log.Logger = log.Output(os.Stdout)
	}

	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().
		Str("level", level).
		Str("format", format).
		Str("env", os.Getenv("ENV")).
		Msg("Logger initialized")
}
