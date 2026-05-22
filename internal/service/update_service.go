package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"clinmitra/internal/config"
	"clinmitra/internal/utils"
)

// UpdateInfo holds information about an available update.
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	DownloadURL    string `json:"downloadURL"`
	ReleaseNotes   string `json:"releaseNotes"`
	PublishedAt    string `json:"publishedAt"`
}

// GitHubRelease represents the GitHub API response for a release.
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	PublishedAt string        `json:"published_at"`
	Assets      []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a downloadable asset in a release.
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateService handles checking for and applying updates.
type UpdateService struct {
	cfg        *config.Config
	owner      string
	repo       string
	httpClient *http.Client
}

// NewUpdateService creates a new UpdateService.
func NewUpdateService(cfg *config.Config) *UpdateService {
	return &UpdateService{
		cfg:   cfg,
		owner: "mdshoaibumer",
		repo:  "clinmitra-clinic-operations",
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// CheckForUpdate checks GitHub Releases for a newer version.
func (s *UpdateService) CheckForUpdate() (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", s.owner, s.repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("failed to create update check request", "error", err)
		return nil, utils.InternalError("Failed to check for updates")
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "ClinmitraDental-Updater")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		slog.Warn("update check failed - no internet?", "error", err)
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: s.cfg.Version,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No releases yet
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: s.cfg.Version,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: s.cfg.Version,
		}, nil
	}

	var release GitHubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1*1024*1024)).Decode(&release); err != nil {
		slog.Error("failed to parse release info", "error", err)
		return nil, utils.InternalError("Failed to parse update information")
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(s.cfg.Version, "v")

	info := &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		ReleaseNotes:   release.Body,
		PublishedAt:    release.PublishedAt,
	}

	if isNewerVersion(currentVersion, latestVersion) {
		info.Available = true
		// Find the Windows installer asset
		for _, asset := range release.Assets {
			if strings.HasSuffix(asset.Name, ".exe") && strings.Contains(strings.ToLower(asset.Name), "setup") {
				info.DownloadURL = asset.BrowserDownloadURL
				break
			}
			// Fallback: any .exe asset
			if strings.HasSuffix(asset.Name, ".exe") {
				info.DownloadURL = asset.BrowserDownloadURL
			}
		}
	}

	return info, nil
}

// DownloadAndInstallUpdate downloads the installer and launches it.
// Only allows downloads from the trusted GitHub releases domain.
func (s *UpdateService) DownloadAndInstallUpdate(downloadURL string) error {
	if downloadURL == "" {
		return utils.ValidationError("No download URL provided")
	}

	// Validate that the URL is from the trusted GitHub releases domain
	if err := validateUpdateURL(downloadURL, s.owner, s.repo); err != nil {
		return err
	}

	slog.Info("downloading update", "url", downloadURL)

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		slog.Error("failed to create download request", "error", err)
		return utils.InternalError("Failed to start download")
	}
	req.Header.Set("User-Agent", "ClinmitraDental-Updater")

	// Use a longer timeout for large binary downloads
	downloadClient := &http.Client{Timeout: 5 * time.Minute}
	resp, err := downloadClient.Do(req)
	if err != nil {
		slog.Error("failed to download update", "error", err)
		return utils.InternalError("Failed to download update")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return utils.InternalError("Download failed with unexpected status")
	}

	// Save installer to temp file
	tmpDir := os.TempDir()
	installerPath := filepath.Join(tmpDir, "ClinmitraDental-Setup.exe")

	out, err := os.Create(installerPath)
	if err != nil {
		slog.Error("failed to create temp file", "path", installerPath, "error", err)
		return utils.InternalError("Failed to save installer")
	}
	defer out.Close()

	if _, err := io.Copy(out, io.LimitReader(resp.Body, 200*1024*1024)); err != nil {
		slog.Error("failed to save installer", "error", err)
		return utils.InternalError("Failed to save installer")
	}
	out.Close()

	slog.Info("update downloaded, launching installer", "path", installerPath)

	// Launch the installer with /SILENT flag (NSIS silent install = shows progress, no prompts)
	cmd := exec.Command(installerPath, "/SILENT")
	if err := cmd.Start(); err != nil {
		slog.Error("failed to launch installer", "path", installerPath, "error", err)
		return utils.InternalError("Failed to launch installer")
	}

	// The app will be closed by the installer (NSIS CloseFirst plugin or manual close)
	return nil
}

// isNewerVersion compares semver strings (e.g., "1.0.0" vs "1.1.0").
func isNewerVersion(current, latest string) bool {
	currentParts := parseVersion(current)
	latestParts := parseVersion(latest)

	for i := 0; i < 3; i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return false
}

// parseVersion splits "1.2.3" into [1, 2, 3].
func parseVersion(v string) [3]int {
	var parts [3]int
	fmt.Sscanf(v, "%d.%d.%d", &parts[0], &parts[1], &parts[2])
	return parts
}

// validateUpdateURL ensures the download URL points to the trusted GitHub
// releases domain for the expected owner/repo. Prevents downloading
// arbitrary executables from untrusted sources.
func validateUpdateURL(rawURL, owner, repo string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return utils.ValidationError("Invalid download URL")
	}
	if parsed.Scheme != "https" {
		return utils.ValidationError("Download URL must use HTTPS")
	}
	host := strings.ToLower(parsed.Host)
	if host != "github.com" && host != "objects.githubusercontent.com" {
		return utils.ValidationError("Download URL must be from github.com")
	}
	// For github.com URLs, ensure the path matches the expected owner/repo
	if host == "github.com" {
		expectedPrefix := fmt.Sprintf("/%s/%s/", owner, repo)
		if !strings.HasPrefix(parsed.Path, expectedPrefix) {
			return utils.ValidationError("Download URL must be from the official repository")
		}
	}
	return nil
}
