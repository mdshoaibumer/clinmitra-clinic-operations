package service

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateJSON_ShortString(t *testing.T) {
	s := `{"name":"test"}`
	result := truncateJSON(s, 100)
	if result != s {
		t.Errorf("expected unchanged string, got: %s", result)
	}
}

func TestTruncateJSON_ExactMaxLen(t *testing.T) {
	s := strings.Repeat("a", 50)
	result := truncateJSON(s, 50)
	if result != s {
		t.Errorf("expected unchanged string at exact maxLen, got length: %d", len(result))
	}
}

func TestTruncateJSON_Truncated(t *testing.T) {
	s := strings.Repeat("a", 100)
	result := truncateJSON(s, 50)
	if !strings.HasSuffix(result, "...[TRUNCATED]") {
		t.Errorf("expected TRUNCATED marker, got: %s", result)
	}
	if len(result) > 50 {
		t.Errorf("result should not exceed maxLen, got length: %d", len(result))
	}
}

func TestTruncateJSON_UTF8Safety(t *testing.T) {
	// Create a string with multi-byte UTF-8 characters (emoji: 4 bytes each)
	// "Hello 🦷🦷🦷🦷🦷" where each 🦷 is 4 bytes
	s := "Hello " + strings.Repeat("🦷", 20) // 6 + 80 = 86 bytes

	// Truncate at a point that would split a multi-byte char
	result := truncateJSON(s, 30)

	if !utf8.ValidString(result) {
		t.Errorf("truncateJSON produced invalid UTF-8: %q", result)
	}
	if !strings.HasSuffix(result, "...[TRUNCATED]") {
		t.Errorf("expected TRUNCATED marker, got: %s", result)
	}
}

func TestTruncateJSON_AllMultiByte(t *testing.T) {
	// String of only multi-byte characters
	s := strings.Repeat("日本語", 100) // 3 bytes each = 900 bytes

	result := truncateJSON(s, 50)

	if !utf8.ValidString(result) {
		t.Errorf("truncateJSON produced invalid UTF-8: %q", result)
	}
	if !strings.HasSuffix(result, "...[TRUNCATED]") {
		t.Error("expected TRUNCATED marker")
	}
}
