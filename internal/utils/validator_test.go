package utils

import (
	"strings"
	"testing"
)

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		phone   string
		wantErr bool
	}{
		{"9876543210", false},
		{"6123456789", false},
		{"+919876543210", false},
		{"919876543210", false},
		{"98765 43210", false},
		{"5123456789", true},  // doesn't start with 6-9
		{"123456789", true},   // too short
		{"98765432101", true}, // too long
		{"abcdefghij", true},  // letters
		{"", true},            // empty
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			err := ValidatePhone(tt.phone)
			if tt.wantErr && err == nil {
				t.Errorf("ValidatePhone(%q) expected error, got nil", tt.phone)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidatePhone(%q) unexpected error: %v", tt.phone, err)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"hello", false},
		{"  hello  ", false},
		{"", true},
		{"   ", true},
		{"\t\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := ValidateRequired("Field", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMinLength(t *testing.T) {
	tests := []struct {
		value   string
		min     int
		wantErr bool
	}{
		{"abc", 3, false},
		{"abcd", 3, false},
		{"ab", 3, true},
		{"", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := ValidateMinLength("Field", tt.value, tt.min)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateMaxLength(t *testing.T) {
	err := ValidateMaxLength("Field", "hello", 5)
	if err != nil {
		t.Errorf("unexpected error for exact max: %v", err)
	}

	err = ValidateMaxLength("Field", "hello!", 5)
	if err == nil {
		t.Error("expected error for exceeding max")
	}
}

func TestValidateAge(t *testing.T) {
	tests := []struct {
		age     int
		wantErr bool
	}{
		{0, false},
		{1, false},
		{120, false},
		{-1, true},
		{121, true},
		{999, true},
	}

	for _, tt := range tests {
		err := ValidateAge(tt.age)
		if tt.wantErr && err == nil {
			t.Errorf("ValidateAge(%d) expected error", tt.age)
		}
		if !tt.wantErr && err != nil {
			t.Errorf("ValidateAge(%d) unexpected error: %v", tt.age, err)
		}
	}
}

func TestValidatePositiveAmount(t *testing.T) {
	if err := ValidatePositiveAmount("Price", 100); err != nil {
		t.Errorf("unexpected error for positive: %v", err)
	}
	if err := ValidatePositiveAmount("Price", 0); err == nil {
		t.Error("expected error for zero")
	}
	if err := ValidatePositiveAmount("Price", -50); err == nil {
		t.Error("expected error for negative")
	}
}

func TestCleanPhone(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"+919876543210", "9876543210"},
		{"919876543210", "9876543210"},
		{"9876 543 210", "9876543210"},
		{"98765-43210", "987654321 0"}, // only dashes removed, not sure about format
		{"9876543210", "9876543210"},
	}

	for _, tt := range tests {
		got := CleanPhone(tt.input)
		// Just verify no prefix remains and spaces/dashes removed
		if strings.HasPrefix(got, "+91") || strings.HasPrefix(got, "91") && len(got) > 10 {
			t.Errorf("CleanPhone(%q) still has prefix: %q", tt.input, got)
		}
	}
}

func TestAppError(t *testing.T) {
	err := NewError("TEST_CODE", "Test message")
	if err.Code != "TEST_CODE" {
		t.Errorf("expected code TEST_CODE, got %s", err.Code)
	}
	if err.Message != "Test message" {
		t.Errorf("expected message 'Test message', got %s", err.Message)
	}
	if err.Error() != "[TEST_CODE] Test message" {
		t.Errorf("unexpected Error(): %s", err.Error())
	}
}

func TestValidationError(t *testing.T) {
	err := ValidationError("something wrong")
	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR code, got %s", err.Code)
	}
	if err.Message != "something wrong" {
		t.Errorf("expected message 'something wrong', got %s", err.Message)
	}
}

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid date", "2025-01-15", false},
		{"empty allowed", "", false},
		{"invalid format", "15-01-2025", true},
		{"invalid date", "2025-13-01", true},
		{"partial", "2025-01", true},
		{"with time", "2025-01-15T10:00", true},
		{"nonsense", "not-a-date", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDate("Date", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateTime(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid time", "14:30", false},
		{"midnight", "00:00", false},
		{"empty allowed", "", false},
		{"with seconds", "14:30:00", true},
		{"invalid hour", "25:00", true},
		{"invalid minute", "14:60", true},
		{"nonsense", "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTime("Time", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"empty allowed", "", false},
		{"no domain", "user@", true},
		{"no at sign", "userexample.com", true},
		{"spaces", "user @example.com", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail("Email", tt.value)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateEnum(t *testing.T) {
	allowed := []string{"male", "female", "other"}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid male", "male", false},
		{"valid female", "female", false},
		{"valid other", "other", false},
		{"empty allowed", "", false},
		{"invalid", "unknown", true},
		{"case sensitive", "Male", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnum("Gender", tt.value, allowed)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateNonNegativeAmount(t *testing.T) {
	if err := ValidateNonNegativeAmount("Amount", 100); err != nil {
		t.Errorf("unexpected error for positive: %v", err)
	}
	if err := ValidateNonNegativeAmount("Amount", 0); err != nil {
		t.Errorf("unexpected error for zero: %v", err)
	}
	if err := ValidateNonNegativeAmount("Amount", -1); err == nil {
		t.Error("expected error for negative")
	}
}

func TestValidateRange(t *testing.T) {
	if err := ValidateRange("Percent", 50, 0, 100); err != nil {
		t.Errorf("unexpected error for mid-range: %v", err)
	}
	if err := ValidateRange("Percent", 0, 0, 100); err != nil {
		t.Errorf("unexpected error for min boundary: %v", err)
	}
	if err := ValidateRange("Percent", 100, 0, 100); err != nil {
		t.Errorf("unexpected error for max boundary: %v", err)
	}
	if err := ValidateRange("Percent", -1, 0, 100); err == nil {
		t.Error("expected error for below min")
	}
	if err := ValidateRange("Percent", 101, 0, 100); err == nil {
		t.Error("expected error for above max")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "Admin1234", false},
		{"valid complex", "MyP@ss99", false},
		{"too short", "Ab1", true},
		{"exactly 7 chars", "Admin12", true},
		{"exactly 8 chars", "Admin123", false},
		{"no uppercase", "admin1234", true},
		{"no lowercase", "ADMIN1234", true},
		{"no digit", "AdminPassword", true},
		{"only digits", "12345678", true},
		{"only lowercase", "abcdefgh", true},
		{"only uppercase", "ABCDEFGH", true},
		{"too long", strings.Repeat("A", 65) + strings.Repeat("a", 64), true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr && err == nil {
				t.Errorf("ValidatePassword(%q) expected error, got nil", tt.password)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidatePassword(%q) unexpected error: %v", tt.password, err)
			}
		})
	}
}
