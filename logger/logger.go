package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/talatmursalin/ekshunno-executor/commonutils"
	"github.com/talatmursalin/ekshunno-executor/config"
	"os"
)

func InitLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func ConfigureLogger(cfg *config.Config) error {
	logConfig := cfg.LogConfig
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	level, err := zerolog.ParseLevel(logConfig.Level)
	if err != nil {
		commonutils.OnError(err, "logger:: log level is not recognized. defaulting to DEBUG")
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)
	switch logConfig.Agent {
	case "file":
		logFile, err := os.OpenFile(logConfig.FilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		log.Logger = log.Output(logFile)
		break
	case "std":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
		break
	default:
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	return nil
}
