package utils

import "fmt"

// AppError represents a structured error with a machine-readable code and
// a human-readable message, suitable for returning to the frontend.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface, formatting the code and message.
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewError creates a new AppError with the given code and message.
func NewError(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

var (
	ErrNotFound            = NewError("NOT_FOUND", "Resource not found")
	ErrUnauthorized        = NewError("UNAUTHORIZED", "Authentication required")
	ErrForbidden           = NewError("FORBIDDEN", "Access denied")
	ErrValidation          = NewError("VALIDATION_ERROR", "Validation failed")
	ErrDuplicate           = NewError("DUPLICATE", "Resource already exists")
	ErrAccountLocked       = NewError("ACCOUNT_LOCKED", "Account is temporarily locked")
	ErrInvalidCredentials  = NewError("INVALID_CREDENTIALS", "Invalid username or password")
	ErrSetupRequired       = NewError("SETUP_REQUIRED", "Initial setup has not been completed")
	ErrSetupAlreadyDone    = NewError("SETUP_COMPLETE", "Setup has already been completed")
	ErrInvoiceImmutable    = NewError("INVOICE_IMMUTABLE", "Cannot modify invoice after payment")
	ErrInsufficientBalance = NewError("INSUFFICIENT_BALANCE", "Payment exceeds remaining balance")
)

// ValidationError creates an AppError with code "VALIDATION_ERROR".
func ValidationError(message string) *AppError {
	return &AppError{Code: "VALIDATION_ERROR", Message: message}
}
