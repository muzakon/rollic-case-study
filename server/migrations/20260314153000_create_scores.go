package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateScores, downCreateScores)
}

func upCreateScores(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
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

	_, err = tx.Exec(`
        CREATE INDEX idx_board_score_time ON scores(board_id, score DESC, achieved_at ASC);
    `)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
        ALTER TABLE scores
        ADD CONSTRAINT fk_scores_board
        FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE;
    `)
	return err
}

func downCreateScores(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS scores;`)
	return err
}
