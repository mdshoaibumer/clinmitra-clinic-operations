package service

import (
	"errors"
	"testing"
	"time"

	"clinmitra/internal/auth"
	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Mocks for PatientService unit tests ---

type unitMockPatientRepo struct {
	patients   map[string]*models.Patient
	byPhone    map[string]*models.Patient
	createErr  error
	deleteErr  error
	listResult []models.Patient
	listTotal  int64
}

func newUnitMockPatientRepo() *unitMockPatientRepo {
	return &unitMockPatientRepo{
		patients: make(map[string]*models.Patient),
		byPhone:  make(map[string]*models.Patient),
	}
}

func (m *unitMockPatientRepo) Create(patient *models.Patient) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.patients[patient.ID] = patient
	m.byPhone[patient.Phone] = patient
	return nil
}

func (m *unitMockPatientRepo) FindByID(id string) (*models.Patient, error) {
	p, ok := m.patients[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return p, nil
}

func (m *unitMockPatientRepo) Update(patient *models.Patient) error {
	m.patients[patient.ID] = patient
	m.byPhone[patient.Phone] = patient
	return nil
}

func (m *unitMockPatientRepo) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.patients, id)
	return nil
}

func (m *unitMockPatientRepo) List(page, pageSize int, search string) ([]models.Patient, int64, error) {
	return m.listResult, m.listTotal, nil
}

func (m *unitMockPatientRepo) FindByPhone(phone string) (*models.Patient, error) {
	p, ok := m.byPhone[phone]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return p, nil
}

func (m *unitMockPatientRepo) Count() (int64, error)                      { return int64(len(m.patients)), nil }
func (m *unitMockPatientRepo) CountSince(sinceDate string) (int64, error) { return 0, nil }

type unitMockPatientTreatmentRepo struct{}

func (m *unitMockPatientTreatmentRepo) Create(pt *models.PatientTreatment) error        { return nil }
func (m *unitMockPatientTreatmentRepo) CreateBatch(pts []models.PatientTreatment) error { return nil }
func (m *unitMockPatientTreatmentRepo) ListByPatient(patientID string) ([]models.PatientTreatment, error) {
	return nil, nil
}

type unitMockInvoiceRepo struct {
	outstanding int64
}

func (m *unitMockInvoiceRepo) Create(invoice *models.Invoice) error        { return nil }
func (m *unitMockInvoiceRepo) FindByID(id string) (*models.Invoice, error) { return nil, nil }
func (m *unitMockInvoiceRepo) Update(invoice *models.Invoice) error        { return nil }
func (m *unitMockInvoiceRepo) List(page, pageSize int, filters repository.InvoiceFilters) ([]models.Invoice, int64, error) {
	return nil, 0, nil
}
func (m *unitMockInvoiceRepo) ListByPatient(patientID string) ([]models.Invoice, error) {
	return nil, nil
}
func (m *unitMockInvoiceRepo) GetLastInvoiceNumber(prefix, yearMonth string) (string, error) {
	return "", nil
}
func (m *unitMockInvoiceRepo) GetOutstandingByPatient(patientID string) (int64, error) {
	return m.outstanding, nil
}
func (m *unitMockInvoiceRepo) GetTotalOutstanding() (int64, error) { return 0, nil }
func (m *unitMockInvoiceRepo) GetRevenueByDateRange(startDate, endDate string) (int64, error) {
	return 0, nil
}
func (m *unitMockInvoiceRepo) GetTotalInvoicedByDateRange(startDate, endDate string) (int64, error) {
	return 0, nil
}
func (m *unitMockInvoiceRepo) GetOutstandingByDateRange(startDate, endDate string) (int64, error) {
	return 0, nil
}

func newUnitTestAuditService() *AuditService {
	return NewAuditService(&unitMockAuditRepo{})
}

type unitMockAuditRepo struct{}

func (m *unitMockAuditRepo) Create(log *models.AuditLog) error                { return nil }
func (m *unitMockAuditRepo) CreateTx(tx *gorm.DB, log *models.AuditLog) error { return nil }
func (m *unitMockAuditRepo) ListByEntity(entityType, entityID string) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *unitMockAuditRepo) ListByUser(userID string, limit int) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *unitMockAuditRepo) ListRecent(limit int) ([]models.AuditLog, error) { return nil, nil }

// newTestPatientService creates a PatientService with an in-memory DB and mocked repos.
func newTestPatientService(t *testing.T) (*PatientService, *unitMockPatientRepo, *unitMockInvoiceRepo) {
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
	db.AutoMigrate(&models.Patient{})

	patientRepo := newUnitMockPatientRepo()
	invoiceRepo := &unitMockInvoiceRepo{}
	auditService := newUnitTestAuditService()

	// Stub auth service with a "logged in" admin user
	authService := &AuthService{currentSession: testAdminSession()}

	svc := NewPatientService(patientRepo, &unitMockPatientTreatmentRepo{}, invoiceRepo, authService, auditService, db)
	return svc, patientRepo, invoiceRepo
}

// testAdminSession returns a stub auth.Session for unit tests.
func testAdminSession() *auth.Session {
	return &auth.Session{
		Token:     "test-token",
		UserID:    "test-user-id",
		Username:  "admin",
		FullName:  "Test Admin",
		Role:      models.RoleAdmin,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}
}

// --- Tests ---

func TestPatientServiceUnit_CreatePatient_Success(t *testing.T) {
	svc, _, _ := newTestPatientService(t)

	patient, err := svc.CreatePatient(CreatePatientInput{
		Name:   "Ramesh Kumar",
		Phone:  "9876543210",
		Gender: "male",
		Age:    35,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if patient == nil {
		t.Fatal("expected patient to be non-nil")
	}
	if patient.Name != "Ramesh Kumar" {
		t.Errorf("expected name 'Ramesh Kumar', got: %s", patient.Name)
	}
	if patient.Phone != "9876543210" {
		t.Errorf("expected phone '9876543210', got: %s", patient.Phone)
	}
	if patient.ID == "" {
		t.Error("expected patient ID to be generated")
	}
	if patient.CreatedBy != "test-user-id" {
		t.Errorf("expected createdBy 'test-user-id', got: %s", patient.CreatedBy)
	}
}

func TestPatientServiceUnit_CreatePatient_DuplicatePhone(t *testing.T) {
	svc, patientRepo, _ := newTestPatientService(t)

	// Pre-populate an existing patient with same phone
	patientRepo.byPhone["9876543210"] = &models.Patient{
		BaseModel: models.BaseModel{ID: "existing-id"},
		Phone:     "9876543210",
	}

	_, err := svc.CreatePatient(CreatePatientInput{
		Name:   "Another Patient",
		Phone:  "9876543210",
		Gender: "male",
		Age:    25,
	})

	if err == nil {
		t.Fatal("expected error for duplicate phone")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T (%v)", err, err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestPatientService_CreatePatient_InvalidEmail(t *testing.T) {
	svc, _, _ := newTestPatientService(t)

	_, err := svc.CreatePatient(CreatePatientInput{
		Name:   "Valid Name",
		Phone:  "9876543210",
		Gender: "male",
		Email:  "not-an-email",
	})

	if err == nil {
		t.Fatal("expected error for invalid email")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
}

func TestPatientService_CreatePatient_InvalidGender(t *testing.T) {
	svc, _, _ := newTestPatientService(t)

	_, err := svc.CreatePatient(CreatePatientInput{
		Name:   "Valid Name",
		Phone:  "9876543210",
		Gender: "invalid",
	})

	if err == nil {
		t.Fatal("expected error for invalid gender")
	}
}

func TestPatientService_DeletePatient_WithUnpaidInvoices(t *testing.T) {
	svc, patientRepo, invoiceRepo := newTestPatientService(t)

	// Add a patient
	patientRepo.patients["pat-1"] = &models.Patient{
		BaseModel: models.BaseModel{ID: "pat-1"},
		Name:      "Patient 1",
		Phone:     "9876543210",
	}

	// Set outstanding amount > 0
	invoiceRepo.outstanding = 50000

	err := svc.DeletePatient("pat-1")
	if err == nil {
		t.Fatal("expected error when deleting patient with unpaid invoices")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T (%v)", err, err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got: %s", appErr.Code)
	}
}

func TestPatientService_DeletePatient_NoOutstanding(t *testing.T) {
	svc, patientRepo, invoiceRepo := newTestPatientService(t)

	patientRepo.patients["pat-1"] = &models.Patient{
		BaseModel: models.BaseModel{ID: "pat-1"},
		Name:      "Patient 1",
		Phone:     "9876543210",
	}
	invoiceRepo.outstanding = 0

	err := svc.DeletePatient("pat-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestPatientService_ListPatients_PaginationDefaults(t *testing.T) {
	svc, _, _ := newTestPatientService(t)

	// Test default values
	resp, err := svc.ListPatients(0, 0, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Page != 1 {
		t.Errorf("expected page=1, got %d", resp.Page)
	}
	if resp.PageSize != 20 {
		t.Errorf("expected pageSize=20, got %d", resp.PageSize)
	}
}

func TestPatientService_ListPatients_MaxPageSize(t *testing.T) {
	svc, _, _ := newTestPatientService(t)

	resp, err := svc.ListPatients(1, 500, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.PageSize != 100 {
		t.Errorf("expected pageSize capped at 100, got %d", resp.PageSize)
	}
}

func TestPatientServiceUnit_GetPatient_NotFound(t *testing.T) {
	svc, _, _ := newTestPatientService(t)

	_, err := svc.GetPatient("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent patient")
	}
}
