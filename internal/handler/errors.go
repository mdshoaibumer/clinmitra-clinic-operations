package handler

import (
	"errors"
	"log/slog"

	"clinmitra/internal/utils"
)

// safeError ensures only AppError types are returned to the frontend.
// Internal errors are logged and replaced with a generic message to prevent
// leaking implementation details through the Wails binding layer.
//
// Time complexity: O(1) — single type assertion
// Space complexity: O(1) — no allocations beyond the sentinel error
func safeError(err error) error {
	if err == nil {
		return nil
	}

	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Log the internal error for debugging, return generic to frontend
	slog.Error("internal error", "error", err.Error())
	return utils.NewError("INTERNAL_ERROR", "An unexpected error occurred. Please try again.")
}
