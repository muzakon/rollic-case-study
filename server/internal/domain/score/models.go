package score

import (
	"time"
)

type Score struct {
	// I use composite primary key to ensure one score per user, per board.
	BoardID string `gorm:"type:varchar(50);primaryKey;index:idx_board_score_time,priority:1"`
	UserID  string `gorm:"type:varchar(50);primaryKey"`

	// The score and when it was achieved
	ScoreValue int       `gorm:"column:score;not null;index:idx_board_score_time,priority:2,sort:desc"`
	AchievedAt time.Time `gorm:"not null;index:idx_board_score_time,priority:3,sort:asc"`

	// Standard audit timestamps
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
