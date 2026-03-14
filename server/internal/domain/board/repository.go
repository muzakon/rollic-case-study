package board

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(board *Board) error {
	return r.db.Create(board).Error
}

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

// List returns boards with cursor-based pagination.
// Orders by created_at DESC, id DESC (newest first).
// Fetches limit+1 rows to determine hasNext without an extra query.
func (r *Repository) List(limit int, cursor *Cursor) ([]Board, bool, error) {
	query := r.db.Model(&Board{}).Order("created_at DESC, id DESC")

	if cursor != nil {
		query = query.Where(
			"(created_at, id) < (?, ?)",
			cursor.CreatedAt, cursor.ID,
		)
	}

	var boards []Board
	err := query.Limit(limit + 1).Find(&boards).Error
	if err != nil {
		return nil, false, err
	}

	hasNext := len(boards) > limit
	if hasNext {
		boards = boards[:limit]
	}

	return boards, hasNext, nil
}
