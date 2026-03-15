package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddScoresUserFK, downAddScoresUserFK)
}

func upAddScoresUserFK(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
		ALTER TABLE scores
		ADD CONSTRAINT fk_scores_user
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
	`)
	return err
}

func downAddScoresUserFK(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE scores DROP CONSTRAINT IF EXISTS fk_scores_user;`)
	return err
}
