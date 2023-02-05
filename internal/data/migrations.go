package data

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

const source = "file://db/migrations"

func doMigrations(dbURI string) error {
	m, err := migrate.New(source, dbURI)
	if err != nil {
		return fmt.Errorf("migrate instance creation error: %w", err)
	}

	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		log.Info().Msg("migrations no changed")
		return nil
	}

	if err != nil {
		return fmt.Errorf("applying migrations error: %w", err)
	}

	log.Info().Msg("migrations done")
	return nil
}
