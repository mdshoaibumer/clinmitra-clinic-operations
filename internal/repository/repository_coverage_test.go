package repository

import (
	"testing"

	"clinmitra/internal/models"

	"github.com/google/uuid"
)

// === AUDIT REPOSITORY TESTS ===

func TestAuditRepo_CreateAndList(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuditRepository(db)

	log := &models.AuditLog{
		UserID:     "user1",
		Action:     models.AuditCreate,
		EntityType: "patient",
		EntityID:   "patient1",
	}

	err := repo.Create(log)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if log.ID == "" {
		t.Fatal("expected ID to be set")
	}

	// ListByEntity
	logs, err := repo.ListByEntity("patient", "patient1")
	if err != nil {
		t.Fatalf("ListByEntity error: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1, got %d", len(logs))
	}

	// ListByUser
	logs, err = repo.ListByUser("user1", 10)
	if err != nil {
		t.Fatalf("ListByUser error: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1, got %d", len(logs))
	}

	// ListRecent
	logs, err = repo.ListRecent(10)
	if err != nil {
		t.Fatalf("ListRecent error: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1, got %d", len(logs))
	}
}

// === CLINIC REPOSITORY TESTS ===

func TestClinicRepo_GetUpsertIsSetup(t *testing.T) {
	db := setupTestDB(t)
	repo := NewClinicRepository(db)

	// Get before setup - should return not found
	_, err := repo.Get()
	if err == nil {
		t.Fatal("expected error for empty clinic")
	}

	// IsSetupComplete before setup
	complete, err := repo.IsSetupComplete()
	if err != nil {
		t.Fatalf("IsSetupComplete error: %v", err)
	}
	if complete {
		t.Fatal("expected not complete")
	}

	// Upsert - creates
	settings := &models.ClinicSettings{
		ClinicName:    "Test Clinic",
		DoctorName:    "Dr. Test",
		SetupComplete: true,
	}
	err = repo.Upsert(settings)
	if err != nil {
		t.Fatalf("Upsert error: %v", err)
	}

	// Get after setup
	fetched, err := repo.Get()
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if fetched.ClinicName != "Test Clinic" {
		t.Fatalf("expected Test Clinic, got %s", fetched.ClinicName)
	}

	// IsSetupComplete after setup
	complete, err = repo.IsSetupComplete()
	if err != nil {
		t.Fatalf("IsSetupComplete error: %v", err)
	}
	if !complete {
		t.Fatal("expected complete")
	}

	// Upsert - updates
	settings.ClinicName = "Updated Clinic"
	err = repo.Upsert(settings)
	if err != nil {
		t.Fatalf("Upsert update error: %v", err)
	}

	fetched, err = repo.Get()
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if fetched.ClinicName != "Updated Clinic" {
		t.Fatalf("expected Updated Clinic, got %s", fetched.ClinicName)
	}
}

// === TREATMENT REPOSITORY TESTS ===

func TestTreatmentRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTreatmentRepository(db)

	treatment := &models.Treatment{
		ID:           uuid.New().String(),
		Name:         "Filling",
		Code:         "FIL",
		Category:     "restorative",
		DefaultPrice: 50000,
		IsActive:     true,
	}

	// Create
	err := repo.Create(treatment)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// FindByID
	found, err := repo.FindByID(treatment.ID)
	if err != nil {
		t.Fatalf("FindByID error: %v", err)
	}
	if found.Name != "Filling" {
		t.Fatalf("expected Filling, got %s", found.Name)
	}

	// Update
	found.Name = "Composite Filling"
	err = repo.Update(found)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	// ListActive
	active, err := repo.ListActive()
	if err != nil {
		t.Fatalf("ListActive error: %v", err)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active, got %d", len(active))
	}

	// ListAll
	all, err := repo.ListAll()
	if err != nil {
		t.Fatalf("ListAll error: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1, got %d", len(all))
	}

	// Delete (soft)
	err = repo.Delete(treatment.ID)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	active, _ = repo.ListActive()
	if len(active) != 0 {
		t.Fatalf("expected 0 active after delete, got %d", len(active))
	}

	// ListAll still shows it
	all, _ = repo.ListAll()
	if len(all) != 1 {
		t.Fatalf("expected 1 in ListAll after soft delete, got %d", len(all))
	}
}

// === PAYMENT REPOSITORY TESTS ===

func TestPaymentRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	paymentRepo := NewPaymentRepository(db)
	invoiceRepo := NewInvoiceRepository(db)
	patientRepo := NewPatientRepository(db)

	// Create patient and invoice first
	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Test Patient",
		Phone:     "9876543210",
		Gender:    "male",
	}
	patientRepo.Create(patient)

	invoice := &models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2605-0001",
		PatientID:     patient.ID,
		InvoiceDate:   "2026-05-20",
		SubTotal:      50000,
		TotalAmount:   50000,
		BalanceAmount: 50000,
		Status:        models.InvoiceIssued,
	}
	invoiceRepo.Create(invoice)

	// Create payment
	payment := &models.Payment{
		BaseModel:   models.BaseModel{ID: uuid.New().String()},
		InvoiceID:   invoice.ID,
		Amount:      30000,
		Method:      "cash",
		PaymentDate: "2026-05-20",
	}
	err := paymentRepo.Create(payment)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// FindByInvoiceID
	payments, err := paymentRepo.FindByInvoiceID(invoice.ID)
	if err != nil {
		t.Fatalf("FindByInvoiceID error: %v", err)
	}
	if len(payments) != 1 {
		t.Fatalf("expected 1, got %d", len(payments))
	}

	// GetTotalByInvoice
	total, err := paymentRepo.GetTotalByInvoice(invoice.ID)
	if err != nil {
		t.Fatalf("GetTotalByInvoice error: %v", err)
	}
	if total != 30000 {
		t.Fatalf("expected 30000, got %d", total)
	}

	// GetCollectionByDate
	dailyTotal, err := paymentRepo.GetCollectionByDate("2026-05-20")
	if err != nil {
		t.Fatalf("GetCollectionByDate error: %v", err)
	}
	if dailyTotal != 30000 {
		t.Fatalf("expected 30000, got %d", dailyTotal)
	}

	// GetCollectionByDateRange
	rangeTotal, err := paymentRepo.GetCollectionByDateRange("2026-05-01", "2026-05-31")
	if err != nil {
		t.Fatalf("GetCollectionByDateRange error: %v", err)
	}
	if rangeTotal != 30000 {
		t.Fatalf("expected 30000, got %d", rangeTotal)
	}

	// ListByDateRange
	payments, err = paymentRepo.ListByDateRange("2026-05-01", "2026-05-31")
	if err != nil {
		t.Fatalf("ListByDateRange error: %v", err)
	}
	if len(payments) != 1 {
		t.Fatalf("expected 1, got %d", len(payments))
	}
}

// === PATIENT TREATMENT REPOSITORY TESTS ===

func TestPatientTreatmentRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPatientTreatmentRepository(db)
	patientRepo := NewPatientRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Test Patient",
		Phone:     "9876543210",
		Gender:    "male",
	}
	patientRepo.Create(patient)

	// Create single
	pt := &models.PatientTreatment{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		PatientID:     patient.ID,
		TreatmentID:   uuid.New().String(),
		InvoiceID:     uuid.New().String(),
		TreatmentDate: "2026-05-20",
		ToothNumber:   "11",
	}
	err := repo.Create(pt)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	// CreateBatch
	batch := []models.PatientTreatment{
		{BaseModel: models.BaseModel{ID: uuid.New().String()}, PatientID: patient.ID, TreatmentID: uuid.New().String(), InvoiceID: uuid.New().String(), TreatmentDate: "2026-05-20"},
		{BaseModel: models.BaseModel{ID: uuid.New().String()}, PatientID: patient.ID, TreatmentID: uuid.New().String(), InvoiceID: uuid.New().String(), TreatmentDate: "2026-05-21"},
	}
	err = repo.CreateBatch(batch)
	if err != nil {
		t.Fatalf("CreateBatch error: %v", err)
	}

	// ListByPatient
	treatments, err := repo.ListByPatient(patient.ID)
	if err != nil {
		t.Fatalf("ListByPatient error: %v", err)
	}
	if len(treatments) != 3 {
		t.Fatalf("expected 3, got %d", len(treatments))
	}
}

// === INVOICE ITEM REPOSITORY TESTS ===

func TestInvoiceItemRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	itemRepo := NewInvoiceItemRepository(db)
	invoiceRepo := NewInvoiceRepository(db)
	patientRepo := NewPatientRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Test Patient",
		Phone:     "9876543210",
		Gender:    "male",
	}
	patientRepo.Create(patient)

	invoice := &models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2605-0001",
		PatientID:     patient.ID,
		InvoiceDate:   "2026-05-20",
		SubTotal:      50000,
		TotalAmount:   50000,
		BalanceAmount: 50000,
		Status:        models.InvoiceIssued,
	}
	invoiceRepo.Create(invoice)

	// CreateBatch
	items := []models.InvoiceItem{
		{BaseModel: models.BaseModel{ID: uuid.New().String()}, InvoiceID: invoice.ID, Description: "Filling", Quantity: 1, UnitPrice: 30000, Amount: 30000},
		{BaseModel: models.BaseModel{ID: uuid.New().String()}, InvoiceID: invoice.ID, Description: "Cleaning", Quantity: 1, UnitPrice: 20000, Amount: 20000},
	}
	err := itemRepo.CreateBatch(items)
	if err != nil {
		t.Fatalf("CreateBatch error: %v", err)
	}

	// FindByInvoiceID
	found, err := itemRepo.FindByInvoiceID(invoice.ID)
	if err != nil {
		t.Fatalf("FindByInvoiceID error: %v", err)
	}
	if len(found) != 2 {
		t.Fatalf("expected 2, got %d", len(found))
	}
}

// === INVOICE REPO: Additional Coverage ===

func TestInvoiceRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInvoiceRepository(db)
	patientRepo := NewPatientRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Test Patient",
		Phone:     "9876543210",
		Gender:    "male",
	}
	patientRepo.Create(patient)

	invoice := &models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2605-0001",
		PatientID:     patient.ID,
		InvoiceDate:   "2026-05-20",
		SubTotal:      50000,
		TotalAmount:   50000,
		BalanceAmount: 50000,
		Status:        models.InvoiceIssued,
	}
	repo.Create(invoice)

	// Update
	invoice.Status = models.InvoicePaid
	invoice.BalanceAmount = 0
	err := repo.Update(invoice)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	found, _ := repo.FindByID(invoice.ID)
	if found.Status != models.InvoicePaid {
		t.Fatalf("expected paid, got %s", found.Status)
	}
}

func TestInvoiceRepo_ListByPatient(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInvoiceRepository(db)
	patientRepo := NewPatientRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Test",
		Phone:     "9876543210",
		Gender:    "male",
	}
	patientRepo.Create(patient)

	repo.Create(&models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2605-0001", PatientID: patient.ID,
		InvoiceDate: "2026-05-20", SubTotal: 50000, TotalAmount: 50000, BalanceAmount: 50000, Status: models.InvoiceIssued,
	})

	invoices, err := repo.ListByPatient(patient.ID)
	if err != nil {
		t.Fatalf("ListByPatient error: %v", err)
	}
	if len(invoices) != 1 {
		t.Fatalf("expected 1, got %d", len(invoices))
	}
}

func TestInvoiceRepo_GetRevenueByDateRange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewInvoiceRepository(db)
	patientRepo := NewPatientRepository(db)
	paymentRepo := NewPaymentRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Test",
		Phone:     "9876543210",
		Gender:    "male",
	}
	patientRepo.Create(patient)

	invoice := &models.Invoice{
		BaseModel:     models.BaseModel{ID: uuid.New().String()},
		InvoiceNumber: "PV-2605-0001", PatientID: patient.ID,
		InvoiceDate: "2026-05-20", SubTotal: 50000, TotalAmount: 50000, BalanceAmount: 0, Status: models.InvoicePaid,
	}
	repo.Create(invoice)

	// Need a payment record since GetRevenueByDateRange queries payments table
	paymentRepo.Create(&models.Payment{
		BaseModel:   models.BaseModel{ID: uuid.New().String()},
		InvoiceID:   invoice.ID,
		Amount:      50000,
		Method:      "cash",
		PaymentDate: "2026-05-20",
	})

	revenue, err := repo.GetRevenueByDateRange("2026-05-01", "2026-05-31")
	if err != nil {
		t.Fatalf("GetRevenueByDateRange error: %v", err)
	}
	if revenue != 50000 {
		t.Fatalf("expected 50000, got %d", revenue)
	}
}

// === APPOINTMENT REPO: Additional Coverage ===

func TestAppointmentRepo_DeleteAndListByPatient(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAppointmentRepository(db)
	patientRepo := NewPatientRepository(db)

	patient := &models.Patient{
		BaseModel: models.BaseModel{ID: uuid.New().String()},
		Name:      "Test",
		Phone:     "9876543210",
		Gender:    "male",
	}
	patientRepo.Create(patient)

	appt := &models.Appointment{
		BaseModel:       models.BaseModel{ID: uuid.New().String()},
		PatientID:       patient.ID,
		AppointmentDate: "2026-12-20",
		StartTime:       "09:00",
		EndTime:         "09:30",
		Duration:        30,
		Status:          models.AppointmentScheduled,
	}
	repo.Create(appt)

	// ListByPatient
	appts, err := repo.ListByPatient(patient.ID)
	if err != nil {
		t.Fatalf("ListByPatient error: %v", err)
	}
	if len(appts) != 1 {
		t.Fatalf("expected 1, got %d", len(appts))
	}

	// Delete
	err = repo.Delete(appt.ID)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	appts, _ = repo.ListByPatient(patient.ID)
	if len(appts) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(appts))
	}
}
