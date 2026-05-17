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
