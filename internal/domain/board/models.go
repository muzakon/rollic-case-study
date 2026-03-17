package board

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BoardSchedule holds the reset schedule configuration stored as JSONB in Postgres.
type BoardSchedule struct {
	Type            string `json:"type"`
	IntervalSeconds *int   `json:"intervalSeconds,omitempty"`
}

// Scan implements sql.Scanner for reading JSONB from the database.
func (s *BoardSchedule) Scan(value any) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan BoardSchedule: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer for writing JSONB to the database.
func (s BoardSchedule) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Board represents a leaderboard entity with optional reset scheduling.
type Board struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`

	// Schedule settings (Nullable)
	Schedule *BoardSchedule `gorm:"type:jsonb"`

	// Indexed so the background worker can quickly find boards due for a reset.
	NextResetAt *time.Time `gorm:"index"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
}

// BeforeCreate is a GORM hook that generates a UUID if one is not already set.
func (b *Board) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
