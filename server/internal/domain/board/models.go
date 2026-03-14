package board

import (
	"time"
)

type BoardSchedule struct {
	Type            string `json:"type"`
	IntervalSeconds *int   `json:"intervalSeconds,omitempty"`
}

type Board struct {
	ID          string `gorm:"type:varchar(50);primaryKey"`
	Name        string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`

	// Schedule settings (Nullable)
	Schedule *BoardSchedule `gorm:"type:jsonb"`

	// Indexed so the background worker can quickly find boards due for a reset.
	NextResetAt *time.Time `gorm:"index"`

	// Standard audit timestamps
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
