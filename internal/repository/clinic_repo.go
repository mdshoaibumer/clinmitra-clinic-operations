package repository

import (
	"clinmitra/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type clinicRepo struct {
	db *gorm.DB
}

// NewClinicRepository creates a GORM-backed ClinicRepository implementation.
func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepo{db: db}
}

// Get retrieves the single ClinicSettings row. Returns gorm.ErrRecordNotFound
// if setup has not been completed.
func (r *clinicRepo) Get() (*models.ClinicSettings, error) {
	var settings models.ClinicSettings
	err := r.db.First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// Upsert creates the clinic settings row if none exists, or updates the
// existing row. Ensures only one ClinicSettings record exists.
func (r *clinicRepo) Upsert(settings *models.ClinicSettings) error {
	if settings.ID == "" {
		settings.ID = uuid.New().String()
	}

	var existing models.ClinicSettings
	err := r.db.First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(settings).Error
	}
	if err != nil {
		return err
	}

	settings.ID = existing.ID
	return r.db.Save(settings).Error
}

// IsSetupComplete returns true if a ClinicSettings record exists with
// SetupComplete=true. Returns (false, nil) if no record found.
func (r *clinicRepo) IsSetupComplete() (bool, error) {
	var settings models.ClinicSettings
	err := r.db.First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return settings.SetupComplete, nil
}
