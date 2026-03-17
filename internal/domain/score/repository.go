package score

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository handles all database operations for the Score entity.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new score repository backed by the given GORM connection.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// GetTopScores returns the top n scores for a board, ordered by score descending
// with ties broken by earliest achieved_at (first-to-score ranks higher).
func (r *Repository) GetTopScores(boardID uuid.UUID, n int) ([]Score, error) {
	var scores []Score
	err := r.db.
		Where("board_id = ?", boardID).
		Order("score DESC, achieved_at ASC").
		Limit(n).
		Find(&scores).Error
	return scores, err
}

// CreateMany inserts multiple scores in a single batch.
func (r *Repository) CreateMany(scores []Score) error {
	return r.db.Create(&scores).Error
}

// Upsert atomically creates or updates a score for a user on a board using
// PostgreSQL ON CONFLICT DO UPDATE. This avoids race conditions that
// FirstOrCreate would introduce under concurrent writes.
func (r *Repository) Upsert(s *Score) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "board_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "achieved_at", "updated_at"}),
	}).Create(s).Error
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

// DeleteByBoardIDs permanently removes all scores for the given boards in a single query.
func (r *Repository) DeleteByBoardIDs(boardIDs []uuid.UUID) error {
	return r.db.Where("board_id IN ?", boardIDs).Delete(&Score{}).Error
}

// GetScoresAbove returns n scores ranked immediately above the given pivot
// (higher score, or same score with earlier achieved_at).
func (r *Repository) GetScoresAbove(boardID uuid.UUID, pivotScore int, pivotAchievedAt time.Time, n int) ([]Score, error) {
	var scores []Score
	err := r.db.
		Where("board_id = ? AND (score > ? OR (score = ? AND achieved_at < ?))",
			boardID, pivotScore, pivotScore, pivotAchievedAt).
		Order("score ASC, achieved_at DESC").
		Limit(n).
		Find(&scores).Error
	return scores, err
}

// GetScoresBelow returns n scores ranked immediately below the given pivot
// (lower score, or same score with later achieved_at).
func (r *Repository) GetScoresBelow(boardID uuid.UUID, pivotScore int, pivotAchievedAt time.Time, n int) ([]Score, error) {
	var scores []Score
	err := r.db.
		Where("board_id = ? AND (score < ? OR (score = ? AND achieved_at > ?))",
			boardID, pivotScore, pivotScore, pivotAchievedAt).
		Order("score DESC, achieved_at ASC").
		Limit(n).
		Find(&scores).Error
	return scores, err
}
