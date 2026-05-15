package handler

import (
	"practivo/internal/models"
	"practivo/internal/service"
)

type AppointmentHandler struct {
	appointmentService *service.AppointmentService
}

// NewAppointmentHandler creates an AppointmentHandler backed by the given service.
func NewAppointmentHandler(appointmentService *service.AppointmentService) *AppointmentHandler {
	return &AppointmentHandler{appointmentService: appointmentService}
}

// CreateAppointment schedules a new appointment after conflict checking.
func (h *AppointmentHandler) CreateAppointment(input service.CreateAppointmentInput) (*models.Appointment, error) {
	result, err := h.appointmentService.CreateAppointment(input)
	return result, safeError(err)
}

// UpdateAppointment modifies an existing scheduled appointment.
func (h *AppointmentHandler) UpdateAppointment(id string, input service.CreateAppointmentInput) (*models.Appointment, error) {
	result, err := h.appointmentService.UpdateAppointment(id, input)
	return result, safeError(err)
}

// CancelAppointment marks an appointment as cancelled with a reason.
func (h *AppointmentHandler) CancelAppointment(id, reason string) error {
	return safeError(h.appointmentService.CancelAppointment(id, reason))
}

// CompleteAppointment transitions a scheduled appointment to completed.
func (h *AppointmentHandler) CompleteAppointment(id string) error {
	return safeError(h.appointmentService.CompleteAppointment(id))
}

// GetTodayAppointments returns all appointments for today.
func (h *AppointmentHandler) GetTodayAppointments() ([]models.Appointment, error) {
	result, err := h.appointmentService.GetTodayAppointments()
	return result, safeError(err)
}

// GetAppointmentsByDate returns all appointments for a specific date.
func (h *AppointmentHandler) GetAppointmentsByDate(date string) ([]models.Appointment, error) {
	result, err := h.appointmentService.GetAppointmentsByDate(date)
	return result, safeError(err)
}

// GetWeekAppointments returns appointments within a date range.
func (h *AppointmentHandler) GetWeekAppointments(startDate, endDate string) ([]models.Appointment, error) {
	result, err := h.appointmentService.GetWeekAppointments(startDate, endDate)
	return result, safeError(err)
}

// GetPatientAppointments returns all appointments for a specific patient.
func (h *AppointmentHandler) GetPatientAppointments(patientID string) ([]models.Appointment, error) {
	result, err := h.appointmentService.GetPatientAppointments(patientID)
	return result, safeError(err)
}

// GetAppointment retrieves a single appointment by ID.
func (h *AppointmentHandler) GetAppointment(id string) (*models.Appointment, error) {
	result, err := h.appointmentService.GetAppointment(id)
	return result, safeError(err)
}
