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

// GetUserScore returns a single score for a user on a board.
func (r *Repository) GetUserScore(boardID uuid.UUID, userID string) (*Score, error) {
	var s Score
	err := r.db.
		Where("board_id = ? AND user_id = ?", boardID, userID).
		First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetScoresAbove returns n scores ranked immediately above the given score (higher score or same score with earlier time).
func (r *Repository) GetScoresAbove(boardID uuid.UUID, pivotScore int, pivotAchievedAt any, n int) ([]Score, error) {
	var scores []Score
	err := r.db.
		Where("board_id = ? AND (score > ? OR (score = ? AND achieved_at < ?))",
			boardID, pivotScore, pivotScore, pivotAchievedAt).
		Order("score ASC, achieved_at DESC").
		Limit(n).
		Find(&scores).Error
	return scores, err
}

// GetScoresBelow returns n scores ranked immediately below the given score (lower score or same score with later time).
func (r *Repository) GetScoresBelow(boardID uuid.UUID, pivotScore int, pivotAchievedAt any, n int) ([]Score, error) {
	var scores []Score
	err := r.db.
		Where("board_id = ? AND (score < ? OR (score = ? AND achieved_at > ?))",
			boardID, pivotScore, pivotScore, pivotAchievedAt).
		Order("score DESC, achieved_at ASC").
		Limit(n).
		Find(&scores).Error
	return scores, err
}
