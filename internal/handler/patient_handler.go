package handler

import (
	"clinmitra/internal/models"
	"clinmitra/internal/service"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
	maxSearchLength = 100
)

type PatientHandler struct {
	patientService *service.PatientService
}

// NewPatientHandler creates a PatientHandler backed by the given PatientService.
func NewPatientHandler(patientService *service.PatientService) *PatientHandler {
	return &PatientHandler{patientService: patientService}
}

// CreatePatient validates and creates a new patient record.
func (h *PatientHandler) CreatePatient(input service.CreatePatientInput) (*models.Patient, error) {
	result, err := h.patientService.CreatePatient(input)
	return result, safeError(err)
}

// UpdatePatient updates an existing patient record by ID.
func (h *PatientHandler) UpdatePatient(id string, input service.CreatePatientInput) (*models.Patient, error) {
	result, err := h.patientService.UpdatePatient(id, input)
	return result, safeError(err)
}

// GetPatient retrieves a single patient by ID.
func (h *PatientHandler) GetPatient(id string) (*models.Patient, error) {
	result, err := h.patientService.GetPatient(id)
	return result, safeError(err)
}

// ListPatients returns a paginated, optionally filtered list of patients.
func (h *PatientHandler) ListPatients(page, pageSize int, search string) (*service.PatientListResponse, error) {
	page, pageSize = sanitizePagination(page, pageSize)
	search = sanitizeSearch(search)
	result, err := h.patientService.ListPatients(page, pageSize, search)
	return result, safeError(err)
}

// DeletePatient soft-deletes a patient by ID (blocked if unpaid invoices exist).
func (h *PatientHandler) DeletePatient(id string) error {
	return safeError(h.patientService.DeletePatient(id))
}

// GetPatientHistory returns the treatment history for a patient.
func (h *PatientHandler) GetPatientHistory(patientID string) ([]models.PatientTreatment, error) {
	result, err := h.patientService.GetPatientHistory(patientID)
	return result, safeError(err)
}

// CheckDuplicatePhone checks if a phone number is already registered.
func (h *PatientHandler) CheckDuplicatePhone(phone string) (*models.Patient, error) {
	result, err := h.patientService.CheckDuplicatePhone(phone)
	return result, safeError(err)
}

// GetPatientCount returns the total number of patients in the system.
func (h *PatientHandler) GetPatientCount() (int64, error) {
	result, err := h.patientService.GetPatientCount()
	return result, safeError(err)
}

// sanitizePagination enforces safe defaults and limits on pagination parameters.
func sanitizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}

// sanitizeSearch truncates search strings to prevent excessively long LIKE queries.
func sanitizeSearch(search string) string {
	if len(search) > maxSearchLength {
		return search[:maxSearchLength]
	}
	return search
}
