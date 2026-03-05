package db

import (
	"context"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trnahnh/katana-id/internal/db/generated"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func Connect(ctx context.Context, connString string) (*gendb.Queries, *pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, nil, err
	}

	queries := gendb.New(pool)
	
	return queries, pool, nil
}

func runMigrations(dbURL string) error {
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dbURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}	
	return nil
}