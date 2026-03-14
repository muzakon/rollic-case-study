package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInit, downInit)
}

func upInit(ctx context.Context, tx *sql.Tx) error {
	// 1. Enable uuid extension
	_, err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`)
	if err != nil {
		return err
	}

	// 2. Create boards table
	_, err = tx.Exec(`
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

	// 3. Create index on boards.next_reset_at
	_, err = tx.Exec(`CREATE INDEX idx_boards_next_reset_at ON boards(next_reset_at);`)
	if err != nil {
		return err
	}

	// 4. Create scores table
	// Composite Primary Key on (board_id, user_id)
	_, err = tx.Exec(`
        CREATE TABLE scores (
            board_id UUID NOT NULL,
            user_id VARCHAR(50) NOT NULL,
            score INTEGER NOT NULL,
            achieved_at TIMESTAMPTZ NOT NULL,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
            PRIMARY KEY (board_id, user_id)
        );
    `)
	if err != nil {
		return err
	}

	// 5. Create composite index for high score lookups
	_, err = tx.Exec(`
        CREATE INDEX idx_board_score_time ON scores(board_id, score DESC, achieved_at ASC);
    `)
	if err != nil {
		return err
	}

	// 6. Add Foreign Key Constraint
	_, err = tx.Exec(`
        ALTER TABLE scores
        ADD CONSTRAINT fk_scores_board
        FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE;
    `)

	return err
}

func downInit(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS scores;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DROP TABLE IF EXISTS boards;`)
	if err != nil {
		return err
	}

	return nil
}
