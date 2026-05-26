//go:build windows

package service

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// IsWhatsAppInstalled checks if WhatsApp Desktop is installed on Windows.
func IsWhatsAppInstalled() bool {
	// Check common WhatsApp Desktop installation paths
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		paths := []string{
			filepath.Join(localAppData, "WhatsApp", "WhatsApp.exe"),
			filepath.Join(localAppData, "Programs", "whatsapp-desktop", "WhatsApp.exe"),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return true
			}
		}
	}

	// Check if whatsapp:// protocol is registered in Windows Registry
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\whatsapp`, registry.READ)
	if err == nil {
		key.Close()
		return true
	}

	// Check HKLM as well (system-wide install)
	key, err = registry.OpenKey(registry.LOCAL_MACHINE, `Software\Classes\whatsapp`, registry.READ)
	if err == nil {
		key.Close()
		return true
	}

	return false
}
