package service

import (
	"errors"
	"testing"

	"clinmitra/internal/utils"
)

func TestValidateUpdateURL(t *testing.T) {
	owner := "mdshoaibumer"
	repo := "clinmitra-clinic-operations"

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid github release URL",
			url:     "https://github.com/mdshoaibumer/clinmitra-clinic-operations/releases/download/v1.0.0/installer.exe",
			wantErr: false,
		},
		{
			name:    "valid githubusercontent URL",
			url:     "https://objects.githubusercontent.com/some-path/installer.exe",
			wantErr: false,
		},
		{
			name:    "http scheme rejected",
			url:     "http://github.com/mdshoaibumer/clinmitra-clinic-operations/releases/download/v1.0.0/installer.exe",
			wantErr: true,
		},
		{
			name:    "wrong domain",
			url:     "https://evil.com/malware.exe",
			wantErr: true,
		},
		{
			name:    "wrong owner on github",
			url:     "https://github.com/attacker/evil-repo/releases/download/v1.0.0/malware.exe",
			wantErr: true,
		},
		{
			name:    "wrong repo on github",
			url:     "https://github.com/mdshoaibumer/wrong-repo/releases/download/v1.0.0/installer.exe",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "ftp scheme",
			url:     "ftp://github.com/file.exe",
			wantErr: true,
		},
		{
			name:    "file scheme",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "javascript scheme",
			url:     "javascript:alert(1)",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUpdateURL(tt.url, owner, repo)
			if tt.wantErr && err == nil {
				t.Errorf("validateUpdateURL(%q) expected error, got nil", tt.url)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("validateUpdateURL(%q) unexpected error: %v", tt.url, err)
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

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  [3]int
	}{
		{"1.2.3", [3]int{1, 2, 3}},
		{"0.0.1", [3]int{0, 0, 1}},
		{"10.20.30", [3]int{10, 20, 30}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseVersion(tt.input)
			if got != tt.want {
				t.Errorf("parseVersion(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"major bump", "1.0.0", "2.0.0", true},
		{"minor bump", "1.0.0", "1.1.0", true},
		{"patch bump", "1.0.0", "1.0.1", true},
		{"same version", "1.0.0", "1.0.0", false},
		{"older version", "1.0.1", "1.0.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNewerVersion(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
			}
		})
	}
}
