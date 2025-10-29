package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init menginisialisasi logger terstruktur (zerolog).
// Uses production-ready JSON format with optional console mode via LOGGER_CONSOLE env var.
func Init(mode string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Optional: Use console writer for easier reading during debugging
	// Set LOGGER_CONSOLE=true in environment to enable
	if mode == "console" || os.Getenv("LOGGER_CONSOLE") == "true" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		log.Info().Msg("Logger initialized (console mode)")
	} else {
		// Production-ready JSON format
		log.Info().Msg("Logger initialized (JSON format)")
	}
}
