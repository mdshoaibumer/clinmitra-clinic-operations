package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&Patient{}, &User{}, &Appointment{}, &Invoice{}, &InvoiceItem{}, &Payment{}, &Treatment{}, &PatientTreatment{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestBaseModel_BeforeCreate_GeneratesUUID(t *testing.T) {
	db := setupTestDB(t)

	patient := &Patient{
		Name:   "Test Patient",
		Phone:  "9876543210",
		Gender: GenderMale,
	}

	if err := db.Create(patient).Error; err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if patient.ID == "" {
		t.Error("expected ID to be generated, got empty string")
	}

	// UUID format check: 8-4-4-4-12
	if len(patient.ID) != 36 {
		t.Errorf("expected UUID (36 chars), got %d chars: %q", len(patient.ID), patient.ID)
	}
}

func TestBaseModel_BeforeCreate_PreservesExistingID(t *testing.T) {
	db := setupTestDB(t)

	customID := "custom-id-12345"
	patient := &Patient{
		BaseModel: BaseModel{ID: customID},
		Name:      "Test Patient",
		Phone:     "9876543210",
		Gender:    GenderMale,
	}

	if err := db.Create(patient).Error; err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if patient.ID != customID {
		t.Errorf("expected preserved ID %q, got %q", customID, patient.ID)
	}
}

func TestBaseModel_SoftDelete(t *testing.T) {
	db := setupTestDB(t)

	patient := &Patient{
		Name:   "Delete Me",
		Phone:  "8765432100",
		Gender: GenderFemale,
	}
	db.Create(patient)

	// Soft delete
	if err := db.Delete(patient).Error; err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	// Should not find with normal query
	var found Patient
	err := db.First(&found, "id = ?", patient.ID).Error
	if err == nil {
		t.Error("expected record not found after soft delete")
	}

	// Should find with Unscoped
	err = db.Unscoped().First(&found, "id = ?", patient.ID).Error
	if err != nil {
		t.Errorf("expected to find soft-deleted record with Unscoped: %v", err)
	}
}

func TestErrRecordNotFound(t *testing.T) {
	if ErrRecordNotFound != gorm.ErrRecordNotFound {
		t.Error("ErrRecordNotFound should equal gorm.ErrRecordNotFound")
	}
}

func TestGenderConstants(t *testing.T) {
	if GenderMale != "male" {
		t.Errorf("GenderMale = %q, want 'male'", GenderMale)
	}
	if GenderFemale != "female" {
		t.Errorf("GenderFemale = %q, want 'female'", GenderFemale)
	}
	if GenderOther != "other" {
		t.Errorf("GenderOther = %q, want 'other'", GenderOther)
	}
}

func TestAppointmentStatusConstants(t *testing.T) {
	if AppointmentScheduled != "scheduled" {
		t.Errorf("AppointmentScheduled = %q", AppointmentScheduled)
	}
	if AppointmentCompleted != "completed" {
		t.Errorf("AppointmentCompleted = %q", AppointmentCompleted)
	}
	if AppointmentCancelled != "cancelled" {
		t.Errorf("AppointmentCancelled = %q", AppointmentCancelled)
	}
	if AppointmentNoShow != "no_show" {
		t.Errorf("AppointmentNoShow = %q", AppointmentNoShow)
	}
}

func TestInvoiceStatusConstants(t *testing.T) {
	if InvoiceIssued != "issued" {
		t.Errorf("InvoiceIssued = %q", InvoiceIssued)
	}
	if InvoicePartial != "partial" {
		t.Errorf("InvoicePartial = %q", InvoicePartial)
	}
	if InvoicePaid != "paid" {
		t.Errorf("InvoicePaid = %q", InvoicePaid)
	}
	if InvoiceVoid != "void" {
		t.Errorf("InvoiceVoid = %q", InvoiceVoid)
	}
}

func TestUserRoleConstants(t *testing.T) {
	if RoleAdmin != "admin" {
		t.Errorf("RoleAdmin = %q", RoleAdmin)
	}
	if RoleDoctor != "doctor" {
		t.Errorf("RoleDoctor = %q", RoleDoctor)
	}
	if RoleReceptionist != "receptionist" {
		t.Errorf("RoleReceptionist = %q", RoleReceptionist)
	}
}

func TestPaymentMethodConstants(t *testing.T) {
	if PaymentCash != "cash" {
		t.Errorf("PaymentCash = %q", PaymentCash)
	}
	if PaymentUPI != "upi" {
		t.Errorf("PaymentUPI = %q", PaymentUPI)
	}
	if PaymentCard != "card" {
		t.Errorf("PaymentCard = %q", PaymentCard)
	}
	if PaymentTransfer != "bank_transfer" {
		t.Errorf("PaymentTransfer = %q", PaymentTransfer)
	}
	if PaymentOther != "other" {
		t.Errorf("PaymentOther = %q", PaymentOther)
	}
}

func TestAuditActionConstants(t *testing.T) {
	if AuditCreate != "CREATE" {
		t.Errorf("AuditCreate = %q", AuditCreate)
	}
	if AuditUpdate != "UPDATE" {
		t.Errorf("AuditUpdate = %q", AuditUpdate)
	}
	if AuditDelete != "DELETE" {
		t.Errorf("AuditDelete = %q", AuditDelete)
	}
	if AuditLogin != "LOGIN" {
		t.Errorf("AuditLogin = %q", AuditLogin)
	}
	if AuditLogout != "LOGOUT" {
		t.Errorf("AuditLogout = %q", AuditLogout)
	}
	if AuditBackup != "BACKUP" {
		t.Errorf("AuditBackup = %q", AuditBackup)
	}
}

func TestPatient_Relationships(t *testing.T) {
	db := setupTestDB(t)

	patient := &Patient{
		Name:   "Relationship Test",
		Phone:  "7654321098",
		Gender: GenderMale,
		Age:    30,
	}
	db.Create(patient)

	// Create appointment for patient
	appt := &Appointment{
		BaseModel:       BaseModel{ID: "appt-1"},
		PatientID:       patient.ID,
		AppointmentDate: "2026-12-01",
		StartTime:       "10:00",
		EndTime:         "10:30",
		Duration:        30,
		Status:          AppointmentScheduled,
	}
	db.Create(appt)

	// Load patient with appointments
	var loaded Patient
	db.Preload("Appointments").First(&loaded, "id = ?", patient.ID)
	if len(loaded.Appointments) != 1 {
		t.Errorf("expected 1 appointment, got %d", len(loaded.Appointments))
	}
}
