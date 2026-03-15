package user

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Exists checks whether a user with the given ID exists.
func (r *Repository) Exists(id string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
