package commonutils

import (
	"github.com/rs/zerolog/log"
)

func ExitOnError(err error, msg string) {
	if err != nil {
		log.Error().Msgf("%s: %s", msg, err)
	}
}

func PanicOnError(err error, msg string) {
	if err != nil {
		log.Panic().Msgf("%s: %s", msg, err)
	}
}

func ReportOnError(err error, msg string) {
	if err != nil {
		log.Info().Msgf("%s: %s", msg, err)
	}
}

func OnError(err error, msg string) {
	if err != nil {
		log.Error().Msgf("%s: %s", msg, err)
	}
}
