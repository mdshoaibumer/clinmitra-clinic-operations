package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var phoneRegex = regexp.MustCompile(`^[6-9]\d{9}$`)

// ValidatePhone checks that a phone number is a valid 10-digit Indian
// mobile number (starting with 6-9) after stripping spaces and +91 prefix.
func ValidatePhone(phone string) error {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.TrimPrefix(cleaned, "+91")
	cleaned = strings.TrimPrefix(cleaned, "91")
	if !phoneRegex.MatchString(cleaned) {
		return ValidationError("Invalid Indian phone number")
	}
	return nil
}

// ValidateRequired returns a validation error if the value is empty or whitespace.
func ValidateRequired(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return ValidationError(fmt.Sprintf("%s is required", field))
	}
	return nil
}

// ValidateMinLength returns a validation error if the trimmed value is
// shorter than min characters.
func ValidateMinLength(field, value string, min int) error {
	if len(strings.TrimSpace(value)) < min {
		return ValidationError(fmt.Sprintf("%s must be at least %d characters", field, min))
	}
	return nil
}

// ValidateMaxLength returns a validation error if the value exceeds max characters.
func ValidateMaxLength(field, value string, max int) error {
	if len(value) > max {
		return ValidationError(fmt.Sprintf("%s must not exceed %d characters", field, max))
	}
	return nil
}

// ValidateAge returns a validation error if age is outside 0-120.
func ValidateAge(age int) error {
	if age < 0 || age > 120 {
		return ValidationError("Age must be between 0 and 120")
	}
	return nil
}

// ValidatePositiveAmount returns a validation error if the amount is zero or negative.
func ValidatePositiveAmount(field string, amount int64) error {
	if amount <= 0 {
		return ValidationError(fmt.Sprintf("%s must be a positive amount", field))
	}
	return nil
}

// CleanPhone strips spaces, dashes, and the +91/91 prefix from a phone number.
func CleanPhone(phone string) string {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.TrimPrefix(cleaned, "+91")
	cleaned = strings.TrimPrefix(cleaned, "91")
	return cleaned
}
