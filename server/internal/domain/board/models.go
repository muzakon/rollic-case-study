package board

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BoardSchedule struct {
	Type            string `json:"type"`
	IntervalSeconds *int   `json:"intervalSeconds,omitempty"`
}

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

func (b *Board) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
