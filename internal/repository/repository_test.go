package repository

import (
	"fmt"
	"testing"

	"clinmitra/internal/models"
	"clinmitra/internal/utils"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database with all tables migrated.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Patient{},
		&models.Appointment{},
		&models.Invoice{},
		&models.InvoiceItem{},
		&models.Payment{},
		&models.ClinicSettings{},
		&models.AuditLog{},
		&models.Treatment{},
		&models.PatientTreatment{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

// === USER REPOSITORY TESTS ===

func TestUserRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		BaseModel:    models.BaseModel{ID: uuid.New().String()},
		Username:     "admin",
		PasswordHash: "hashed",
		FullName:     "Admin User",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify retrieval
	found, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("expected no error on FindByID, got: %v", err)
	}
	if found.Username != "admin" {
		t.Errorf("expected username 'admin', got: %s", found.Username)
	}
}

func TestUserRepo_FindByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		BaseModel:    models.BaseModel{ID: uuid.New().String()},
		Username:     "doctor1",
		PasswordHash: "hashed",
		FullName:     "Dr. Smith",
		Role:         models.RoleDoctor,
		IsActive:     true,
	}
	repo.Create(user)

	found, err := repo.FindByUsername("doctor1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if found.FullName != "Dr. Smith" {
		t.Errorf("expected 'Dr. Smith', got: %s", found.FullName)
	}
}

func TestUserRepo_FindByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindByUsername("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	if err != utils.ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestUserRepo_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.FindByID("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent ID")
	}
	if err != utils.ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestUserRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		BaseModel:    models.BaseModel{ID: uuid.New().String()},
		Username:     "admin",
		PasswordHash: "hash1",
		FullName:     "Old Name",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	repo.Create(user)

	user.FullName = "New Name"
	err := repo.Update(user)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	found, _ := repo.FindByID(user.ID)
	if found.FullName != "New Name" {
		t.Errorf("expected 'New Name', got: %s", found.FullName)
	}
}

func TestUserRepo_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	count, err := repo.Count()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got: %d", count)
	}

	repo.Create(&models.User{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Username:  "user1", PasswordHash: "h", FullName: "U1", Role: models.RoleAdmin, IsActive: true,
	})
	repo.Create(&models.User{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Username:  "user2", PasswordHash: "h", FullName: "U2", Role: models.RoleDoctor, IsActive: true,
	})

	count, _ = repo.Count()
	if count != 2 {
		t.Errorf("expected 2, got: %d", count)
	}
}

func TestUserRepo_UpdateLastLogin(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &models.User{
		BaseModel:    models.BaseModel{ID: uuid.New().String()},
		Username:     "admin",
		PasswordHash: "h",
		FullName:     "Admin",
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	repo.Create(user)

	err := repo.UpdateLastLogin(user.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	found, _ := repo.FindByID(user.ID)
	if found.LastLoginAt == nil {
		t.Error("expected LastLoginAt to be set")
	}
}

// === PATIENT REPOSITORY TESTS ===

func TestPatientRepo_CreateAndFind(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "John Doe",
		Phone:     "9876543210",
		Gender:    "male",
		Age:       30,
	}

	err := repo.Create(patient)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	found, err := repo.FindByID(patient.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if found.Name != "John Doe" {
		t.Errorf("expected 'John Doe', got: %s", found.Name)
	}
	if found.Phone != "9876543210" {
		t.Errorf("expected '9876543210', got: %s", found.Phone)
	}
}

func TestPatientRepo_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientRepository(db)

	_, err := repo.FindByID("nonexistent")
	if err != utils.ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestPatientRepo_FindByPhone(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Jane Doe",
		Phone:     "9123456789",
		Gender:    "female",
	}
	repo.Create(patient)

	found, err := repo.FindByPhone("9123456789")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if found.Name != "Jane Doe" {
		t.Errorf("expected 'Jane Doe', got: %s", found.Name)
	}
}

func TestPatientRepo_FindByPhone_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientRepository(db)

	_, err := repo.FindByPhone("0000000000")
	if err != utils.ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestPatientRepo_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientRepository(db)

	// Create 15 patients
	for i := 0; i < 15; i++ {
		repo.Create(&models.Patient{
			BaseModel: models.BaseModel{ID: uuid.New().String()},
			Name:      "Patient " + uuid.New().String()[:4],
			Phone:     "98765" + fmt.Sprintf("%05d", i),
			Gender:    "male",
		})
	}

	// Page 1 with 10 per page
	patients, total, err := repo.List(1, 10, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if total != 15 {
		t.Errorf("expected total 15, got: %d", total)
	}
	if len(patients) != 10 {
		t.Errorf("expected 10 patients on page 1, got: %d", len(patients))
	}

	// Page 2
	patients, _, err = repo.List(2, 10, "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(patients) != 5 {
		t.Errorf("expected 5 patients on page 2, got: %d", len(patients))
	}
}

func TestPatientRepo_List_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientRepository(db)

	repo.Create(&models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Rahul Sharma",
		Phone:     "9876543210",
		Gender:    "male",
	})
	repo.Create(&models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Priya Patel",
		Phone:     "9123456789",
		Gender:    "female",
	})

	// Search by name
	patients, total, _ := repo.List(1, 10, "Rahul")
	if total != 1 {
		t.Errorf("expected 1 result for 'Rahul', got: %d", total)
	}
	if len(patients) != 1 || patients[0].Name != "Rahul Sharma" {
		t.Error("expected Rahul Sharma in results")
	}

	// Search by phone
	patients, total, _ = repo.List(1, 10, "9123")
	if total != 1 {
		t.Errorf("expected 1 result for phone '9123', got: %d", total)
	}
	if len(patients) != 1 || patients[0].Name != "Priya Patel" {
		t.Error("expected Priya Patel in results")
	}

	// No results
	_, total, _ = repo.List(1, 10, "ZZZZZ")
	if total != 0 {
		t.Errorf("expected 0 results, got: %d", total)
	}
}

func TestPatientRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Original",
		Phone:     "9876543210",
		Gender:    "male",
		Age:       25,
	}
	repo.Create(patient)

	patient.Name = "Updated"
	patient.Age = 26
	err := repo.Update(patient)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	found, _ := repo.FindByID(patient.ID)
	if found.Name != "Updated" {
		t.Errorf("expected 'Updated', got: %s", found.Name)
	}
	if found.Age != 26 {
		t.Errorf("expected age 26, got: %d", found.Age)
	}
}

func TestPatientRepo_Delete_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "ToDelete",
		Phone:     "9876543210",
		Gender:    "male",
	}
	repo.Create(patient)

	err := repo.Delete(patient.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should not be findable anymore
	_, err = repo.FindByID(patient.ID)
	if err != utils.ErrNotFound {
		t.Errorf("expected ErrNotFound after soft delete, got: %v", err)
	}

	// Should not appear in list
	_, total, _ := repo.List(1, 10, "")
	if total != 0 {
		t.Errorf("expected 0 patients after delete, got: %d", total)
	}
}

func TestPatientRepo_Count(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientRepository(db)

	count, _ := repo.Count()
	if count != 0 {
		t.Errorf("expected 0, got: %d", count)
	}

	repo.Create(&models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "P1", Phone: "9876543210", Gender: "male",
	})

	count, _ = repo.Count()
	if count != 1 {
		t.Errorf("expected 1, got: %d", count)
	}
}

// === APPOINTMENT REPOSITORY TESTS ===

func createTestPatient(t *testing.T, db *gorm.DB) *models.Patient {
	t.Helper()
	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Test Patient",
		Phone:     "9876543210",
		Gender:    "male",
	}
	db.Create(patient)
	return patient
}

func TestAppointmentRepo_CreateAndFind(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)
	patient := createTestPatient(t, db)

	appt := &models.Appointment{
		BaseModel:       models.BaseModel{ID: uuid.New().String()},
		PatientID:       patient.ID,
		AppointmentDate: "2025-01-15",
		StartTime:       "10:00",
		EndTime:         "10:30",
		Duration:        30,
		Status:          models.AppointmentScheduled,
		Purpose:         "Checkup",
	}

	err := repo.Create(appt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	found, err := repo.FindByID(appt.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if found.Purpose != "Checkup" {
		t.Errorf("expected 'Checkup', got: %s", found.Purpose)
	}
	// Patient should be preloaded
	if found.Patient.Name != "Test Patient" {
		t.Errorf("expected patient preloaded, got: %s", found.Patient.Name)
	}
}

func TestAppointmentRepo_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)

	_, err := repo.FindByID("nonexistent")
	if err != utils.ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestAppointmentRepo_ListByDate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)
	patient := createTestPatient(t, db)

	// Create appointments on different dates
	repo.Create(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-15",
		StartTime: "09:00", EndTime: "09:30", Duration: 30, Status: models.AppointmentScheduled,
	})
	repo.Create(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-15",
		StartTime: "11:00", EndTime: "11:30", Duration: 30, Status: models.AppointmentScheduled,
	})
	repo.Create(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-16",
		StartTime: "10:00", EndTime: "10:30", Duration: 30, Status: models.AppointmentScheduled,
	})

	results, err := repo.ListByDate("2025-01-15")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 appointments on 2025-01-15, got: %d", len(results))
	}
	// Should be ordered by start_time
	if len(results) == 2 && results[0].StartTime > results[1].StartTime {
		t.Error("expected appointments ordered by start time")
	}
}

func TestAppointmentRepo_ListByDateRange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)
	patient := createTestPatient(t, db)

	repo.Create(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-14",
		StartTime: "09:00", EndTime: "09:30", Duration: 30, Status: models.AppointmentScheduled,
	})
	repo.Create(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-15",
		StartTime: "09:00", EndTime: "09:30", Duration: 30, Status: models.AppointmentScheduled,
	})
	repo.Create(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-20",
		StartTime: "09:00", EndTime: "09:30", Duration: 30, Status: models.AppointmentScheduled,
	})

	results, err := repo.ListByDateRange("2025-01-14", "2025-01-15")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 appointments in range, got: %d", len(results))
	}
}

func TestAppointmentRepo_FindConflicting(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)
	patient := createTestPatient(t, db)

	existing := &models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-15",
		StartTime: "10:00", EndTime: "10:30", Duration: 30, Status: models.AppointmentScheduled,
	}
	repo.Create(existing)

	// Overlapping time
	conflict, err := repo.FindConflicting("2025-01-15", "10:15", "10:45", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if conflict == nil {
		t.Fatal("expected conflict to be found")
	}

	// Non-overlapping time
	conflict, err = repo.FindConflicting("2025-01-15", "11:00", "11:30", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if conflict != nil {
		t.Error("expected no conflict for non-overlapping time")
	}

	// Exclude self
	conflict, err = repo.FindConflicting("2025-01-15", "10:00", "10:30", existing.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if conflict != nil {
		t.Error("expected no conflict when excluding self")
	}
}

func TestAppointmentRepo_FindConflicting_IgnoresCancelled(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)
	patient := createTestPatient(t, db)

	// Create a cancelled appointment
	repo.Create(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-15",
		StartTime: "10:00", EndTime: "10:30", Duration: 30, Status: models.AppointmentCancelled,
	})

	// Same time should not conflict (cancelled doesn't block)
	conflict, err := repo.FindConflicting("2025-01-15", "10:00", "10:30", "")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if conflict != nil {
		t.Error("expected no conflict with cancelled appointment")
	}
}

func TestAppointmentRepo_CountByDate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)
	patient := createTestPatient(t, db)

	repo.Create(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-15",
		StartTime: "09:00", EndTime: "09:30", Duration: 30, Status: models.AppointmentScheduled,
	})
	repo.Create(&models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-15",
		StartTime: "10:00", EndTime: "10:30", Duration: 30, Status: models.AppointmentCancelled,
	})

	count, err := repo.CountByDate("2025-01-15")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// Only scheduled should count
	if count != 1 {
		t.Errorf("expected 1 (only scheduled), got: %d", count)
	}
}

func TestAppointmentRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)
	patient := createTestPatient(t, db)

	appt := &models.Appointment{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		PatientID: patient.ID, AppointmentDate: "2025-01-15",
		StartTime: "10:00", EndTime: "10:30", Duration: 30, Status: models.AppointmentScheduled,
		Purpose: "Cleaning",
	}
	repo.Create(appt)

	appt.Status = models.AppointmentCompleted
	appt.Notes = "Completed successfully"
	err := repo.Update(appt)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	found, _ := repo.FindByID(appt.ID)
	if found.Status != models.AppointmentCompleted {
		t.Errorf("expected completed status, got: %s", found.Status)
	}
	if found.Notes != "Completed successfully" {
		t.Errorf("expected notes, got: %s", found.Notes)
	}
}

// === INVOICE REPOSITORY TESTS ===

func TestInvoiceRepo_CreateAndFind(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInvoiceRepository(db)
	patient := createTestPatient(t, db)

	invoice := &models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-001",
		PatientID:     patient.ID,
		InvoiceDate:   "2025-01-15",
		SubTotal:      50000,
		TotalAmount:   50000,
		PaidAmount:    0,
		BalanceAmount: 50000,
		Status:        models.InvoiceIssued,
	}

	err := repo.Create(invoice)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	found, err := repo.FindByID(invoice.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if found.InvoiceNumber != "PV-2501-001" {
		t.Errorf("expected 'PV-2501-001', got: %s", found.InvoiceNumber)
	}
	if found.Patient.Name != "Test Patient" {
		t.Errorf("expected patient preloaded, got: %s", found.Patient.Name)
	}
}

func TestInvoiceRepo_List_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInvoiceRepository(db)
	patient := createTestPatient(t, db)

	// Create invoices with different statuses
	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-001", PatientID: patient.ID,
		InvoiceDate: "2025-01-15", SubTotal: 10000, TotalAmount: 10000,
		PaidAmount: 10000, BalanceAmount: 0, Status: models.InvoicePaid,
	})
	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-002", PatientID: patient.ID,
		InvoiceDate: "2025-01-16", SubTotal: 20000, TotalAmount: 20000,
		PaidAmount: 0, BalanceAmount: 20000, Status: models.InvoiceIssued,
	})
	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-003", PatientID: patient.ID,
		InvoiceDate: "2025-02-01", SubTotal: 30000, TotalAmount: 30000,
		PaidAmount: 15000, BalanceAmount: 15000, Status: models.InvoicePartial,
	})

	// Filter by status
	invoices, total, err := repo.List(1, 10, InvoiceFilters{Status: string(models.InvoiceIssued)})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if total != 1 {
		t.Errorf("expected 1 issued invoice, got: %d", total)
	}
	if len(invoices) != 1 {
		t.Errorf("expected 1 result, got: %d", len(invoices))
	}

	// Filter by date range
	invoices, total, _ = repo.List(1, 10, InvoiceFilters{StartDate: "2025-01-15", EndDate: "2025-01-16"})
	if total != 2 {
		t.Errorf("expected 2 invoices in Jan 15-16, got: %d", total)
	}

	// All invoices
	_, total, _ = repo.List(1, 10, InvoiceFilters{})
	if total != 3 {
		t.Errorf("expected 3 total invoices, got: %d", total)
	}
}

func TestInvoiceRepo_GetLastInvoiceNumber(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInvoiceRepository(db)
	patient := createTestPatient(t, db)

	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-001", PatientID: patient.ID,
		InvoiceDate: "2025-01-15", SubTotal: 10000, TotalAmount: 10000,
		Status: models.InvoicePaid,
	})
	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-003", PatientID: patient.ID,
		InvoiceDate: "2025-01-16", SubTotal: 20000, TotalAmount: 20000,
		Status: models.InvoiceIssued,
	})

	last, err := repo.GetLastInvoiceNumber("PV", "2501")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if last != "PV-2501-003" {
		t.Errorf("expected 'PV-2501-003', got: %s", last)
	}

	// Non-existent prefix/month
	last, err = repo.GetLastInvoiceNumber("PV", "2502")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if last != "" {
		t.Errorf("expected empty string for non-existent month, got: %s", last)
	}
}

func TestInvoiceRepo_GetTotalOutstanding(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInvoiceRepository(db)
	patient := createTestPatient(t, db)

	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-001", PatientID: patient.ID,
		InvoiceDate: "2025-01-15", SubTotal: 10000, TotalAmount: 10000,
		PaidAmount: 10000, BalanceAmount: 0, Status: models.InvoicePaid,
	})
	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-002", PatientID: patient.ID,
		InvoiceDate: "2025-01-16", SubTotal: 20000, TotalAmount: 20000,
		PaidAmount: 0, BalanceAmount: 20000, Status: models.InvoiceIssued,
	})
	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-003", PatientID: patient.ID,
		InvoiceDate: "2025-01-17", SubTotal: 30000, TotalAmount: 30000,
		PaidAmount: 10000, BalanceAmount: 20000, Status: models.InvoicePartial,
	})

	outstanding, err := repo.GetTotalOutstanding()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	// 20000 (issued) + 20000 (partial) = 40000
	if outstanding != 40000 {
		t.Errorf("expected 40000, got: %d", outstanding)
	}
}

func TestInvoiceRepo_GetOutstandingByPatient(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInvoiceRepository(db)
	patient1 := createTestPatient(t, db)

	patient2 := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Other Patient",
		Phone:     "9111111111",
		Gender:    "female",
	}
	db.Create(patient2)

	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-001", PatientID: patient1.ID,
		InvoiceDate: "2025-01-15", SubTotal: 10000, TotalAmount: 10000,
		BalanceAmount: 10000, Status: models.InvoiceIssued,
	})
	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2501-002", PatientID: patient2.ID,
		InvoiceDate: "2025-01-15", SubTotal: 50000, TotalAmount: 50000,
		BalanceAmount: 50000, Status: models.InvoiceIssued,
	})

	outstanding, err := repo.GetOutstandingByPatient(patient1.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if outstanding != 10000 {
		t.Errorf("expected 10000, got: %d", outstanding)
	}
}
