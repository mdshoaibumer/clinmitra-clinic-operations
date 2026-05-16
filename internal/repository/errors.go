package repository

import (
	"errors"

	"clinmitra/internal/utils"

	"gorm.io/gorm"
)

// WrapError translates GORM-specific errors into application-level errors.
// This ensures no GORM implementation details leak to the service layer.
// All repositories should call this before returning errors to callers.
func WrapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return utils.ErrNotFound
	}
	return err
}
