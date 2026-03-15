package score

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetTopScores returns the top n scores for a board, ordered by score descending.
func (r *Repository) GetTopScores(boardID uuid.UUID, n int) ([]Score, error) {
	var scores []Score
	err := r.db.
		Where("board_id = ?", boardID).
		Order("score DESC, achieved_at ASC").
		Limit(n).
		Find(&scores).Error
	return scores, err
}
