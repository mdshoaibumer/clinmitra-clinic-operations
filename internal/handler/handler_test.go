package handler

import (
	"strings"
	"testing"
)

func TestSanitizePagination(t *testing.T) {
	tests := []struct {
		name             string
		page             int
		pageSize         int
		expectedPage     int
		expectedPageSize int
	}{
		{"valid normal values", 1, 20, 1, 20},
		{"page 5 size 50", 5, 50, 5, 50},
		{"zero page", 0, 20, 1, 20},
		{"negative page", -3, 20, 1, 20},
		{"zero pageSize", 1, 0, 1, defaultPageSize},
		{"negative pageSize", 1, -10, 1, defaultPageSize},
		{"pageSize over max", 1, 500, 1, maxPageSize},
		{"pageSize exactly max", 1, 100, 1, 100},
		{"both bad", -1, -1, 1, defaultPageSize},
		{"pageSize 1 is valid", 1, 1, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, pageSize := sanitizePagination(tt.page, tt.pageSize)
			if page != tt.expectedPage {
				t.Errorf("page: expected %d, got %d", tt.expectedPage, page)
			}
			if pageSize != tt.expectedPageSize {
				t.Errorf("pageSize: expected %d, got %d", tt.expectedPageSize, pageSize)
			}
		})
	}
}

func TestSanitizeSearch(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"short string", "John", "John"},
		{"exact max length", strings.Repeat("a", maxSearchLength), strings.Repeat("a", maxSearchLength)},
		{"over max length", strings.Repeat("b", maxSearchLength+50), strings.Repeat("b", maxSearchLength)},
		{"normal search", "Dr. Smith", "Dr. Smith"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeSearch(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
