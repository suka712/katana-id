package db

import (
	"context"
	"log"

	_ "github.com/lib/pq"

	"github.com/trnahnh/katana-id/internal/db/ent"
)

// Connect opens the Ent client against Postgres and runs Ent's auto-migration,
// which keeps the live schema in sync with the typed schema in ent/schema.
// Because the schema is the single source of truth, there is no separate SQL
// migration step and no drift between code and database.
func Connect(ctx context.Context, connString string) (*ent.Client, error) {
	client, err := ent.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	if err := client.Schema.Create(ctx); err != nil {
		client.Close()
		return nil, err
	}

	log.Print("☁️  DB connected & schema synced")
	return client, nil
}
