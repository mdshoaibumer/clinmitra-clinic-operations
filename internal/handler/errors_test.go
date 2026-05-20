package handler

import (
	"errors"
	"testing"

	"clinmitra/internal/utils"
)

func TestSafeError_NilInput(t *testing.T) {
	result := safeError(nil)
	if result != nil {
		t.Errorf("safeError(nil) should return nil, got: %v", result)
	}
}

func TestSafeError_AppError(t *testing.T) {
	appErr := utils.NewError("VALIDATION_ERROR", "Name is required")
	result := safeError(appErr)
	if result != appErr {
		t.Errorf("safeError should return AppError unchanged, got: %v", result)
	}

	// Also test ValidationError helper
	valErr := utils.ValidationError("Invalid phone")
	result = safeError(valErr)
	var gotAppErr *utils.AppError
	if !errors.As(result, &gotAppErr) {
		t.Errorf("safeError(ValidationError) should return AppError, got: %T", result)
	}
	if gotAppErr.Message != "Invalid phone" {
		t.Errorf("expected message 'Invalid phone', got: %s", gotAppErr.Message)
	}
}

func TestSafeError_InternalError(t *testing.T) {
	internalErr := errors.New("sql: connection refused")
	result := safeError(internalErr)

	if result == nil {
		t.Fatal("safeError should not return nil for internal errors")
	}

	var appErr *utils.AppError
	if !errors.As(result, &appErr) {
		t.Fatalf("safeError should wrap internal errors as AppError, got: %T", result)
	}

	if appErr.Code != "INTERNAL_ERROR" {
		t.Errorf("expected code INTERNAL_ERROR, got: %s", appErr.Code)
	}

	// Should NOT leak the original error message
	if appErr.Message == "sql: connection refused" {
		t.Error("safeError should not expose internal error details")
	}
}
