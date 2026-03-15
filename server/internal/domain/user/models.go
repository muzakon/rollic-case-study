package user

import "time"

type User struct {
	ID        string    `gorm:"type:varchar(50);primaryKey"`
	Username  string    `gorm:"type:varchar(100);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
