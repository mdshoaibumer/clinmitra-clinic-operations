package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error: %v", err)
	}

	if cfg.AppName != "Clinmitra Dental" {
		t.Errorf("AppName = %q, want 'Clinmitra Dental'", cfg.AppName)
	}
	if cfg.Version != "1.1.1" {
		t.Errorf("Version = %q, want '1.1.1'", cfg.Version)
	}
	if cfg.MaxLoginAttempts != 5 {
		t.Errorf("MaxLoginAttempts = %d, want 5", cfg.MaxLoginAttempts)
	}
	if cfg.LockoutMinutes != 15 {
		t.Errorf("LockoutMinutes = %d, want 15", cfg.LockoutMinutes)
	}
	if cfg.SessionHours != 8 {
		t.Errorf("SessionHours = %d, want 8", cfg.SessionHours)
	}
	if cfg.BcryptCost != 12 {
		t.Errorf("BcryptCost = %d, want 12", cfg.BcryptCost)
	}

	// Verify paths are set
	if cfg.DataDir == "" {
		t.Error("DataDir should not be empty")
	}
	if cfg.DBPath == "" {
		t.Error("DBPath should not be empty")
	}
	if cfg.BackupDir == "" {
		t.Error("BackupDir should not be empty")
	}
	if cfg.LogDir == "" {
		t.Error("LogDir should not be empty")
	}

	// DBPath should be inside DataDir
	if filepath.Dir(cfg.DBPath) != cfg.DataDir {
		t.Errorf("DBPath %q should be inside DataDir %q", cfg.DBPath, cfg.DataDir)
	}
}

func TestNewConfig_DirectoriesExist(t *testing.T) {
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig() error: %v", err)
	}

	for _, dir := range []string{cfg.DataDir, cfg.BackupDir, cfg.LogDir} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("directory %q should exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q should be a directory", dir)
		}
	}
}

func TestEnsureDirectories_CreatesNewDirs(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &Config{
		DataDir:   filepath.Join(tmpDir, "data"),
		BackupDir: filepath.Join(tmpDir, "data", "backups"),
		LogDir:    filepath.Join(tmpDir, "data", "logs"),
	}

	err := cfg.ensureDirectories()
	if err != nil {
		t.Fatalf("ensureDirectories() error: %v", err)
	}

	for _, dir := range []string{cfg.DataDir, cfg.BackupDir, cfg.LogDir} {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("directory %q was not created", dir)
		}
	}
}

func TestGetDataDir(t *testing.T) {
	dir, err := getDataDir()
	if err != nil {
		t.Fatalf("getDataDir() error: %v", err)
	}
	if dir == "" {
		t.Error("getDataDir() should not return empty string")
	}
	if filepath.Base(dir) != "ClinmitraDental" {
		t.Errorf("getDataDir() base should be 'ClinmitraDental', got: %q", filepath.Base(dir))
	}
}
