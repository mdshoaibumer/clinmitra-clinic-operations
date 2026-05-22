package utils

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

var phoneRegex = regexp.MustCompile(`^[6-9]\d{9}$`)

// stripIndianPrefix removes common Indian country-code prefixes (+91, 091, 0091, 91)
// and returns a clean digit string. Only strips "91" if the remainder is a valid
// 10-digit mobile number to avoid mangling numbers that naturally start with 91.
func stripIndianPrefix(digits string) string {
	// Try common prefixes in order of specificity
	for _, prefix := range []string{"+91", "0091", "091", "91"} {
		if strings.HasPrefix(digits, prefix) {
			candidate := digits[len(prefix):]
			if phoneRegex.MatchString(candidate) {
				return candidate
			}
		}
	}
	return digits
}

// ValidatePhone checks that a phone number is a valid 10-digit Indian
// mobile number (starting with 6-9) after stripping spaces and country-code prefixes.
func ValidatePhone(phone string) error {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = stripIndianPrefix(cleaned)
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

// CleanPhone strips spaces, dashes, and Indian country-code prefixes from a phone number.
func CleanPhone(phone string) string {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = stripIndianPrefix(cleaned)
	return cleaned
}

// ValidateDate validates a date string matches YYYY-MM-DD format and is a real date.
func ValidateDate(field, value string) error {
	if value == "" {
		return nil // empty dates are allowed (optional field)
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return ValidationError(fmt.Sprintf("%s must be a valid date (YYYY-MM-DD)", field))
	}
	return nil
}

// ValidateTime validates a time string matches HH:MM format.
func ValidateTime(field, value string) error {
	if value == "" {
		return nil
	}
	if _, err := time.Parse("15:04", value); err != nil {
		return ValidationError(fmt.Sprintf("%s must be a valid time (HH:MM)", field))
	}
	return nil
}

// ValidateEmail validates an email address format. Empty values are allowed.
func ValidateEmail(field, value string) error {
	if value == "" {
		return nil
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return ValidationError(fmt.Sprintf("%s must be a valid email address", field))
	}
	return nil
}

// ValidateEnum checks that a value is one of the allowed values. Empty values are allowed.
func ValidateEnum(field, value string, allowed []string) error {
	if value == "" {
		return nil
	}
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return ValidationError(fmt.Sprintf("%s must be one of: %s", field, strings.Join(allowed, ", ")))
}

// ValidateNonNegativeAmount returns a validation error if the amount is negative.
func ValidateNonNegativeAmount(field string, amount int64) error {
	if amount < 0 {
		return ValidationError(fmt.Sprintf("%s must not be negative", field))
	}
	return nil
}

// ValidateRange returns a validation error if the value is outside [min, max].
func ValidateRange(field string, value, min, max float64) error {
	if value < min || value > max {
		return ValidationError(fmt.Sprintf("%s must be between %.0f and %.0f", field, min, max))
	}
	return nil
}

// ValidatePassword enforces password complexity rules:
// - Minimum 8 characters
// - At least one uppercase letter
// - At least one lowercase letter
// - At least one digit
// - Maximum 128 characters
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ValidationError("Password must be at least 8 characters")
	}
	if len(password) > 128 {
		return ValidationError("Password must not exceed 128 characters")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		}
	}

	if !hasUpper {
		return ValidationError("Password must contain at least one uppercase letter")
	}
	if !hasLower {
		return ValidationError("Password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return ValidationError("Password must contain at least one digit")
	}
	return nil
}
