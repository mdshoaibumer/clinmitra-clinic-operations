package handler

import (
	"practivo/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates an AuthHandler backed by the given AuthService.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login authenticates a user with username/password and returns session info.
// Exposed to the Wails frontend via binding.
func (h *AuthHandler) Login(username, password string) (*service.AuthResponse, error) {
	result, err := h.authService.Login(username, password)
	return result, safeError(err)
}

// Logout destroys the current user session.
func (h *AuthHandler) Logout() error {
	return safeError(h.authService.Logout())
}

// GetCurrentUser returns the currently authenticated user info, or
// a response with LoggedIn=false if no session is active.
func (h *AuthHandler) GetCurrentUser() (*service.AuthResponse, error) {
	result, err := h.authService.GetCurrentUser()
	return result, safeError(err)
}

// ChangePassword updates the password for the currently logged-in user
// after verifying the old password.
func (h *AuthHandler) ChangePassword(oldPassword, newPassword string) error {
	return safeError(h.authService.ChangePassword(oldPassword, newPassword))
}
