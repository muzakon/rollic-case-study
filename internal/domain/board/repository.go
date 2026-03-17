package board

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository handles all database operations for the Board entity.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new board repository backed by the given GORM connection.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new board into the database.
func (r *Repository) Create(board *Board) error {
	return r.db.Create(board).Error
}

// GetByID retrieves a board by its UUID primary key.
func (r *Repository) GetByID(id uuid.UUID) (*Board, error) {
	var board Board
	err := r.db.Where("id = ?", id).First(&board).Error
	if err != nil {
		return nil, err
	}
	return &board, nil
}

// Exists checks whether a board with the given ID exists.
func (r *Repository) Exists(id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&Board{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Count returns the total number of boards.
func (r *Repository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&Board{}).Count(&count).Error
	return count, err
}

// Cursor represents the decoded pagination cursor.
type Cursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// EncodeCursor encodes a cursor to a base64 string.
func EncodeCursor(createdAt time.Time, id uuid.UUID) string {
	raw := fmt.Sprintf("%s,%s", createdAt.Format(time.RFC3339Nano), id.String())
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

// DecodeCursor decodes a base64 cursor string.
func DecodeCursor(encoded string) (*Cursor, error) {
	raw, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor")
	}

	parts := strings.SplitN(string(raw), ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cursor format")
	}

	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid cursor timestamp")
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid cursor id")
	}

	return &Cursor{CreatedAt: t, ID: id}, nil
}

// FindDueBoards returns all boards whose next_reset_at is due (i.e. <= now) and have a schedule set.
func (r *Repository) FindDueBoards(now time.Time) ([]Board, error) {
	var boards []Board
	err := r.db.
		Where("next_reset_at IS NOT NULL AND next_reset_at <= ?", now).
		Find(&boards).Error
	return boards, err
}

// UpdateNextResetAt sets a new next_reset_at value for the given board.
// Passing nil clears the field (for schedules that only fire once).
func (r *Repository) UpdateNextResetAt(id uuid.UUID, nextResetAt *time.Time) error {
	return r.db.Model(&Board{}).
		Where("id = ?", id).
		Update("next_reset_at", nextResetAt).Error
}

// ResetEntry holds the new next_reset_at value for a single board.
type ResetEntry struct {
	ID          uuid.UUID
	NextResetAt *time.Time
}

// BatchUpdateNextResetAt updates next_reset_at for multiple boards in a single transaction.
func (r *Repository) BatchUpdateNextResetAt(entries []ResetEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, e := range entries {
			if err := tx.Model(&Board{}).
				Where("id = ?", e.ID).
				Update("next_reset_at", e.NextResetAt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ResetBoards deletes all scores for the given board IDs and advances their
// next_reset_at in a single transaction, ensuring atomicity.
func (r *Repository) ResetBoards(boardIDs []uuid.UUID, entries []ResetEntry) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM scores WHERE board_id IN ?", boardIDs).Error; err != nil {
			return err
		}
		for _, e := range entries {
			if err := tx.Model(&Board{}).
				Where("id = ?", e.ID).
				Update("next_reset_at", e.NextResetAt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// List returns boards with cursor-based pagination.
// Orders by created_at DESC, id DESC (newest first).
// If limit is nil, fetches all boards without pagination.
// If limit is provided, fetches limit+1 rows to determine hasNext without an extra query.
func (r *Repository) List(limit *int, cursor *Cursor) ([]Board, bool, error) {
	query := r.db.Model(&Board{}).Order("created_at DESC, id DESC")

	if cursor != nil {
		query = query.Where(
			"(created_at, id) < (?, ?)",
			cursor.CreatedAt, cursor.ID,
		)
	}

	var boards []Board
	var err error

	if limit != nil {
		err = query.Limit(*limit + 1).Find(&boards).Error
	} else {
		err = query.Find(&boards).Error
	}

	if err != nil {
		return nil, false, err
	}

	var hasNext bool
	if limit != nil {
		hasNext = len(boards) > *limit
		if hasNext {
			boards = boards[:*limit]
		}
	} else {
		hasNext = false // No pagination when limit is nil
	}

	return boards, hasNext, nil
}
