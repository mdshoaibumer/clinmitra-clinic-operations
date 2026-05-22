package service

import (
	"errors"
	"testing"

	"clinmitra/internal/models"
	"clinmitra/internal/utils"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Mocks for AppointmentService unit tests ---

type unitMockAppointmentRepo struct {
	appointments map[string]*models.Appointment
	conflicting  *models.Appointment
	byDate       []models.Appointment
}

func newUnitMockAppointmentRepo() *unitMockAppointmentRepo {
	return &unitMockAppointmentRepo{
		appointments: make(map[string]*models.Appointment),
	}
}

func (m *unitMockAppointmentRepo) Create(appointment *models.Appointment) error {
	m.appointments[appointment.ID] = appointment
	return nil
}

func (m *unitMockAppointmentRepo) FindByID(id string) (*models.Appointment, error) {
	a, ok := m.appointments[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return a, nil
}

func (m *unitMockAppointmentRepo) Update(appointment *models.Appointment) error {
	m.appointments[appointment.ID] = appointment
	return nil
}

func (m *unitMockAppointmentRepo) Delete(id string) error {
	delete(m.appointments, id)
	return nil
}

func (m *unitMockAppointmentRepo) ListByDate(date string) ([]models.Appointment, error) {
	return m.byDate, nil
}

func (m *unitMockAppointmentRepo) ListByDateRange(startDate, endDate string) ([]models.Appointment, error) {
	return nil, nil
}

func (m *unitMockAppointmentRepo) ListByPatient(patientID string) ([]models.Appointment, error) {
	return nil, nil
}

func (m *unitMockAppointmentRepo) FindConflicting(date, startTime, endTime, excludeID string) (*models.Appointment, error) {
	if m.conflicting != nil && m.conflicting.ID != excludeID {
		return m.conflicting, nil
	}
	return nil, nil
}

func (m *unitMockAppointmentRepo) CountByDate(date string) (int64, error) {
	return int64(len(m.byDate)), nil
}

type apptMockPatientRepo struct {
	patients map[string]*models.Patient
}

func (m *apptMockPatientRepo) Create(patient *models.Patient) error { return nil }
func (m *apptMockPatientRepo) FindByID(id string) (*models.Patient, error) {
	p, ok := m.patients[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return p, nil
}
func (m *apptMockPatientRepo) Update(patient *models.Patient) error { return nil }
func (m *apptMockPatientRepo) Delete(id string) error               { return nil }
func (m *apptMockPatientRepo) List(page, pageSize int, search string) ([]models.Patient, int64, error) {
	return nil, 0, nil
}
func (m *apptMockPatientRepo) FindByPhone(phone string) (*models.Patient, error) {
	return nil, utils.ErrNotFound
}
func (m *apptMockPatientRepo) Count() (int64, error)                      { return 0, nil }
func (m *apptMockPatientRepo) CountSince(sinceDate string) (int64, error) { return 0, nil }

func newTestAppointmentService(t *testing.T) (*AppointmentService, *unitMockAppointmentRepo, *apptMockPatientRepo) {
	t.Helper()

	dsn := "file:" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	db.AutoMigrate(&models.Appointment{}, &models.Patient{})

	apptRepo := newUnitMockAppointmentRepo()
	patientRepo := &apptMockPatientRepo{
		patients: map[string]*models.Patient{
			"patient-1": {BaseModel: models.BaseModel{ID: "patient-1"}, Name: "Test Patient"},
		},
	}
	auditService := newUnitTestAuditService()
	authService := &AuthService{currentSession: testAdminSession()}

	svc := NewAppointmentService(apptRepo, patientRepo, authService, auditService, db)
	return svc, apptRepo, patientRepo
}

// --- Tests ---

func TestAppointmentService_Create_Success(t *testing.T) {
	svc, _, _ := newTestAppointmentService(t)

	appt, err := svc.CreateAppointment(CreateAppointmentInput{
		PatientID: "patient-1",
		Date:      "2026-06-15",
		StartTime: "10:00",
		EndTime:   "10:30",
		Duration:  30,
		Purpose:   "Checkup",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if appt == nil {
		t.Fatal("expected appointment to be non-nil")
	}
	if appt.Status != models.AppointmentScheduled {
		t.Errorf("expected status 'scheduled', got: %s", appt.Status)
	}
	if appt.Duration != 30 {
		t.Errorf("expected duration 30, got: %d", appt.Duration)
	}
}

func TestAppointmentService_Create_PatientNotFound(t *testing.T) {
	svc, _, _ := newTestAppointmentService(t)

	_, err := svc.CreateAppointment(CreateAppointmentInput{
		PatientID: "nonexistent-patient",
		Date:      "2026-06-15",
		StartTime: "10:00",
		EndTime:   "10:30",
		Duration:  30,
	})

	if err == nil {
		t.Fatal("expected error for nonexistent patient")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
}

func TestAppointmentService_Create_TimeConflict(t *testing.T) {
	svc, apptRepo, _ := newTestAppointmentService(t)

	// Set up a conflicting appointment
	apptRepo.conflicting = &models.Appointment{
		BaseModel: models.BaseModel{ID: "existing-appt"},
		PatientID: "patient-1",
		StartTime: "10:00",
		EndTime:   "10:30",
	}

	_, err := svc.CreateAppointment(CreateAppointmentInput{
		PatientID: "patient-1",
		Date:      "2026-06-15",
		StartTime: "10:15",
		EndTime:   "10:45",
		Duration:  30,
	})

	if err == nil {
		t.Fatal("expected error for time conflict")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestAppointmentService_Create_StartAfterEnd(t *testing.T) {
	svc, _, _ := newTestAppointmentService(t)

	_, err := svc.CreateAppointment(CreateAppointmentInput{
		PatientID: "patient-1",
		Date:      "2026-06-15",
		StartTime: "10:30",
		EndTime:   "10:00",
		Duration:  30,
	})

	if err == nil {
		t.Fatal("expected error when start time is after end time")
	}
}

func TestAppointmentService_Create_InvalidDate(t *testing.T) {
	svc, _, _ := newTestAppointmentService(t)

	_, err := svc.CreateAppointment(CreateAppointmentInput{
		PatientID: "patient-1",
		Date:      "not-a-date",
		StartTime: "10:00",
		EndTime:   "10:30",
	})

	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestAppointmentService_Create_InvalidTime(t *testing.T) {
	svc, _, _ := newTestAppointmentService(t)

	_, err := svc.CreateAppointment(CreateAppointmentInput{
		PatientID: "patient-1",
		Date:      "2026-06-15",
		StartTime: "25:00",
		EndTime:   "10:30",
	})

	if err == nil {
		t.Fatal("expected error for invalid time")
	}
}

func TestAppointmentService_Create_DefaultDuration(t *testing.T) {
	svc, _, _ := newTestAppointmentService(t)

	appt, err := svc.CreateAppointment(CreateAppointmentInput{
		PatientID: "patient-1",
		Date:      "2026-06-15",
		StartTime: "10:00",
		EndTime:   "10:30",
		Duration:  0, // Should default to 30
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if appt.Duration != 30 {
		t.Errorf("expected default duration 30, got: %d", appt.Duration)
	}
}

func TestAppointmentService_Cancel_Success(t *testing.T) {
	svc, apptRepo, _ := newTestAppointmentService(t)

	apptRepo.appointments["appt-1"] = &models.Appointment{
		BaseModel: models.BaseModel{ID: "appt-1"},
		PatientID: "patient-1",
		Status:    models.AppointmentScheduled,
	}

	err := svc.CancelAppointment("appt-1", "Patient requested")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	updated := apptRepo.appointments["appt-1"]
	if updated.Status != models.AppointmentCancelled {
		t.Errorf("expected status 'cancelled', got: %s", updated.Status)
	}
	if updated.CancelReason != "Patient requested" {
		t.Errorf("expected reason 'Patient requested', got: %s", updated.CancelReason)
	}
}

func TestAppointmentService_Cancel_CompletedAppointment(t *testing.T) {
	svc, apptRepo, _ := newTestAppointmentService(t)

	apptRepo.appointments["appt-1"] = &models.Appointment{
		BaseModel: models.BaseModel{ID: "appt-1"},
		Status:    models.AppointmentCompleted,
	}

	err := svc.CancelAppointment("appt-1", "Changed mind")
	if err == nil {
		t.Fatal("expected error when cancelling completed appointment")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
}

func TestAppointmentService_Cancel_AlreadyCancelled(t *testing.T) {
	svc, apptRepo, _ := newTestAppointmentService(t)

	apptRepo.appointments["appt-1"] = &models.Appointment{
		BaseModel: models.BaseModel{ID: "appt-1"},
		Status:    models.AppointmentCancelled,
	}

	err := svc.CancelAppointment("appt-1", "Again")
	if err == nil {
		t.Fatal("expected error when cancelling already cancelled appointment")
	}
}

func TestAppointmentService_Complete_Success(t *testing.T) {
	svc, apptRepo, _ := newTestAppointmentService(t)

	apptRepo.appointments["appt-1"] = &models.Appointment{
		BaseModel: models.BaseModel{ID: "appt-1"},
		Status:    models.AppointmentScheduled,
	}

	err := svc.CompleteAppointment("appt-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	updated := apptRepo.appointments["appt-1"]
	if updated.Status != models.AppointmentCompleted {
		t.Errorf("expected status 'completed', got: %s", updated.Status)
	}
}

func TestAppointmentService_Complete_NonScheduled(t *testing.T) {
	svc, apptRepo, _ := newTestAppointmentService(t)

	apptRepo.appointments["appt-1"] = &models.Appointment{
		BaseModel: models.BaseModel{ID: "appt-1"},
		Status:    models.AppointmentCancelled,
	}

	err := svc.CompleteAppointment("appt-1")
	if err == nil {
		t.Fatal("expected error when completing non-scheduled appointment")
	}
}

func TestAppointmentService_Cancel_NotFound(t *testing.T) {
	svc, _, _ := newTestAppointmentService(t)

	err := svc.CancelAppointment("nonexistent", "reason")
	if err == nil {
		t.Fatal("expected error for nonexistent appointment")
	}
}
