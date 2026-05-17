package service

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CloudDriveInfo represents a detected cloud sync folder on the user's system.
type CloudDriveInfo struct {
	Provider  string `json:"provider"`  // "google_drive", "onedrive", "dropbox"
	Path      string `json:"path"`      // Full path to the sync folder
	Available bool   `json:"available"` // Whether the folder exists and is writable
}

// DetectCloudDrives scans for known cloud storage sync folders on the system.
// Works by checking standard installation paths for Google Drive, OneDrive,
// and Dropbox desktop clients.
func DetectCloudDrives() []CloudDriveInfo {
	var drives []CloudDriveInfo

	if runtime.GOOS == "windows" {
		drives = append(drives, detectWindowsCloudDrives()...)
	} else {
		drives = append(drives, detectUnixCloudDrives()...)
	}

	return drives
}

func detectWindowsCloudDrives() []CloudDriveInfo {
	var drives []CloudDriveInfo
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return drives
	}

	// Google Drive for Desktop — checks common mount points
	googleDrivePaths := []string{
		filepath.Join(homeDir, "Google Drive"),
		filepath.Join(homeDir, "My Drive"),
	}
	// Also check drive letters G: through Z: for Google Drive virtual drive
	for letter := 'G'; letter <= 'Z'; letter++ {
		driveRoot := string(letter) + ":\\My Drive"
		googleDrivePaths = append(googleDrivePaths, driveRoot)
	}
	// Check via environment variable if set
	if gdPath := os.Getenv("GOOGLE_DRIVE_PATH"); gdPath != "" {
		googleDrivePaths = append([]string{gdPath}, googleDrivePaths...)
	}

	for _, p := range googleDrivePaths {
		if isWritableDir(p) {
			drives = append(drives, CloudDriveInfo{
				Provider:  "google_drive",
				Path:      p,
				Available: true,
			})
			break
		}
	}

	// OneDrive
	oneDrivePaths := []string{
		filepath.Join(homeDir, "OneDrive"),
		os.Getenv("OneDrive"),
		os.Getenv("OneDriveConsumer"),
		os.Getenv("OneDriveCommercial"),
	}
	for _, p := range oneDrivePaths {
		if p != "" && isWritableDir(p) {
			drives = append(drives, CloudDriveInfo{
				Provider:  "onedrive",
				Path:      p,
				Available: true,
			})
			break
		}
	}

	// Dropbox
	dropboxPath := filepath.Join(homeDir, "Dropbox")
	if isWritableDir(dropboxPath) {
		drives = append(drives, CloudDriveInfo{
			Provider:  "dropbox",
			Path:      dropboxPath,
			Available: true,
		})
	}

	return drives
}

func detectUnixCloudDrives() []CloudDriveInfo {
	var drives []CloudDriveInfo
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return drives
	}

	// Google Drive (via google-drive-ocamlfuse or similar)
	gdPaths := []string{
		filepath.Join(homeDir, "Google Drive"),
		filepath.Join(homeDir, "google-drive"),
	}
	for _, p := range gdPaths {
		if isWritableDir(p) {
			drives = append(drives, CloudDriveInfo{Provider: "google_drive", Path: p, Available: true})
			break
		}
	}

	// Dropbox
	dropboxPath := filepath.Join(homeDir, "Dropbox")
	if isWritableDir(dropboxPath) {
		drives = append(drives, CloudDriveInfo{Provider: "dropbox", Path: dropboxPath, Available: true})
	}

	return drives
}

// isWritableDir checks if a path exists, is a directory, and is writable.
func isWritableDir(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}

	// Test write access by creating a temp file
	testFile := filepath.Join(path, ".clinmitra_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(testFile)
	return true
}
