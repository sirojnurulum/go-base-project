package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init menginisialisasi logger terstruktur (zerolog).
func Init(env string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Di lingkungan development, gunakan console writer yang lebih mudah dibaca.
	// Di produksi, gunakan format JSON default.
	if env == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	log.Info().Msg("Logger initialized")
}