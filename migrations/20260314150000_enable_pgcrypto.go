package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upEnablePgcrypto, downEnablePgcrypto)
}

func upEnablePgcrypto(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`)
	return err
}

func downEnablePgcrypto(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP EXTENSION IF EXISTS "pgcrypto";`)
	return err
}
