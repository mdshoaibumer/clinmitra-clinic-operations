package repository

import (
	"time"

	"clinmitra/internal/models"

	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository creates a GORM-backed UserRepository implementation.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepo{db: db}
}

// Create persists a new User record to the database.
func (r *userRepo) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByID retrieves a user by primary key.
func (r *userRepo) FindByID(id string) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername retrieves a user by unique username.
func (r *userRepo) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update saves changed fields on an existing user record.
func (r *userRepo) Update(user *models.User) error {
	return r.db.Model(user).Updates(user).Error
}

// UpdateLastLogin sets the last_login_at timestamp for a user to the current time.
func (r *userRepo) UpdateLastLogin(id string) error {
	now := time.Now()
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("last_login_at", &now).Error
}

// Count returns the total number of registered users.
func (r *userRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.User{}).Count(&count).Error
	return count, err
}
