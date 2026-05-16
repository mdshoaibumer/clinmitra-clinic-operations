package handler

import (
	"clinmitra/internal/models"
	"clinmitra/internal/service"
)

type SettingsHandler struct {
	settingsService *service.SettingsService
}

// NewSettingsHandler creates a SettingsHandler backed by the given service.
func NewSettingsHandler(settingsService *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsService: settingsService}
}

// IsSetupComplete returns whether the initial clinic setup wizard has been completed.
func (h *SettingsHandler) IsSetupComplete() (bool, error) {
	result, err := h.settingsService.IsSetupComplete()
	return result, safeError(err)
}

// CompleteSetup runs the first-time setup wizard: creates the admin user
// and saves clinic settings. Can only be called once.
func (h *SettingsHandler) CompleteSetup(input service.SetupInput) error {
	return safeError(h.settingsService.CompleteSetup(input))
}

// GetClinicSettings returns the current clinic configuration.
func (h *SettingsHandler) GetClinicSettings() (*models.ClinicSettings, error) {
	result, err := h.settingsService.GetClinicSettings()
	return result, safeError(err)
}

// UpdateClinicSettings persists updated clinic settings.
func (h *SettingsHandler) UpdateClinicSettings(settings *models.ClinicSettings) error {
	return safeError(h.settingsService.UpdateClinicSettings(settings))
}

// ListTreatments returns all active treatments available for invoicing.
func (h *SettingsHandler) ListTreatments() ([]models.Treatment, error) {
	result, err := h.settingsService.ListTreatments()
	return result, safeError(err)
}

// ListAllTreatments returns all treatments including inactive ones.
func (h *SettingsHandler) ListAllTreatments() ([]models.Treatment, error) {
	result, err := h.settingsService.ListAllTreatments()
	return result, safeError(err)
}

// CreateTreatment adds a new dental treatment/procedure to the system.
func (h *SettingsHandler) CreateTreatment(name, code, category, description string, defaultPrice int64) (*models.Treatment, error) {
	result, err := h.settingsService.CreateTreatment(name, code, category, description, defaultPrice)
	return result, safeError(err)
}

// UpdateTreatment modifies an existing treatment's details.
func (h *SettingsHandler) UpdateTreatment(id, name, code, category, description string, defaultPrice int64) error {
	return safeError(h.settingsService.UpdateTreatment(id, name, code, category, description, defaultPrice))
}

// DeleteTreatment soft-deletes a treatment (marks as inactive).
func (h *SettingsHandler) DeleteTreatment(id string) error {
	return safeError(h.settingsService.DeleteTreatment(id))
}
