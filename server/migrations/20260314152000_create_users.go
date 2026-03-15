package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateUsers, downCreateUsers)
}

func upCreateUsers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE users (
			id VARCHAR(50) PRIMARY KEY,
			username VARCHAR(100) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func downCreateUsers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS users;`)
	return err
}
