package migratedb

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed gen/migrations/*
var migrations embed.FS

func ApplyMigrations(db *sql.DB, dsn string) error {
	sourceInstance, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("error creating source driver: %w", err)
	}

	postgresDriver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		return fmt.Errorf("cannot create database driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceInstance, "postgres", postgresDriver)
	if err != nil {
		return fmt.Errorf("cannot create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("error applying migrations: %w", err)
	}

	return nil
}
