package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateBoards, downCreateBoards)
}

func upCreateBoards(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
        CREATE TABLE boards (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            description TEXT,
            schedule JSONB,
            next_reset_at TIMESTAMPTZ,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
    `)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_boards_next_reset_at ON boards(next_reset_at);`)
	return err
}

func downCreateBoards(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS boards;`)
	return err
}
