package main

import (
	"os"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"ritual/cmd"
	"ritual/internal/db"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	database, err := db.Init()
	if err != nil {
		zlog.Error().
			Err(err).
			Msg("Database initialization error. Exiting ...")
		os.Exit(1)
	}
	defer db.Close(database)

	if err := cmd.Execute(); err != nil {
		zlog.Error().Err(err)
		os.Exit(1)
	}
}
