package handler

import (
	"errors"
	"testing"

	"clinmitra/internal/utils"
)

func TestValidateWhatsAppURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid whatsapp protocol URL",
			url:     "whatsapp://send?phone=919876543210&text=Hello",
			wantErr: false,
		},
		{
			name:    "valid wa.me URL",
			url:     "https://wa.me/919876543210?text=Hello%20World",
			wantErr: false,
		},
		{
			name:    "whatsapp URL without query",
			url:     "whatsapp://send?phone=919876543210",
			wantErr: false,
		},
		{
			name:    "http URL rejected",
			url:     "http://evil.com/steal-data",
			wantErr: true,
		},
		{
			name:    "https non-whatsapp domain",
			url:     "https://evil.com/phishing",
			wantErr: true,
		},
		{
			name:    "javascript injection",
			url:     "javascript:alert(document.cookie)",
			wantErr: true,
		},
		{
			name:    "file scheme",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "whatsapp without send",
			url:     "whatsapp://other",
			wantErr: true,
		},
		{
			name:    "data URL",
			url:     "data:text/html,<script>alert(1)</script>",
			wantErr: true,
		},
		{
			name:    "https wa.me without path",
			url:     "https://wa.me/",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWhatsAppURL(tt.url)
			if tt.wantErr && err == nil {
				t.Errorf("validateWhatsAppURL(%q) expected error, got nil", tt.url)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateWhatsAppURL(%q) unexpected error: %v", tt.url, err)
			}
			if err != nil {
				var appErr *utils.AppError
				if !errors.As(err, &appErr) {
					t.Errorf("expected AppError, got: %T", err)
				}
			}
		})
	}
}

func TestSendViaWhatsApp_RejectsInvalidURL(t *testing.T) {
	handler := NewWhatsAppHandler(nil, nil)

	// Should reject non-whatsapp URLs
	err := handler.SendViaWhatsApp("https://evil.com/phishing")
	if err == nil {
		t.Fatal("expected error for non-whatsapp URL")
	}

	// Should accept valid whatsapp URL (no ctx, so no browser action)
	err = handler.SendViaWhatsApp("https://wa.me/919876543210?text=hello")
	if err != nil {
		t.Fatalf("unexpected error for valid wa.me URL: %v", err)
	}

	// Should accept valid whatsapp:// URL
	err = handler.SendViaWhatsApp("whatsapp://send?phone=919876543210&text=hello")
	if err != nil {
		t.Fatalf("unexpected error for valid whatsapp:// URL: %v", err)
	}
}
