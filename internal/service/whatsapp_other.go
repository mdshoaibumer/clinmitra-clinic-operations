//go:build !windows

package service

// IsWhatsAppInstalled returns false on non-Windows platforms
// since WhatsApp Desktop registry detection is Windows-only.
func IsWhatsAppInstalled() bool {
	return false
}
