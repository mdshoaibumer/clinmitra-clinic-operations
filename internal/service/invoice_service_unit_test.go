package service

import (
	"errors"
	"testing"

	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// --- Mocks for InvoiceService unit tests (void/payment logic) ---

type voidMockInvoiceRepo struct {
	invoices map[string]*models.Invoice
}

func newVoidMockInvoiceRepo() *voidMockInvoiceRepo {
	return &voidMockInvoiceRepo{invoices: make(map[string]*models.Invoice)}
}

func (m *voidMockInvoiceRepo) Create(invoice *models.Invoice) error {
	m.invoices[invoice.ID] = invoice
	return nil
}
func (m *voidMockInvoiceRepo) FindByID(id string) (*models.Invoice, error) {
	inv, ok := m.invoices[id]
	if !ok {
		return nil, utils.ErrNotFound
	}
	return inv, nil
}
func (m *voidMockInvoiceRepo) Update(invoice *models.Invoice) error {
	m.invoices[invoice.ID] = invoice
	return nil
}
func (m *voidMockInvoiceRepo) List(page, pageSize int, filters repository.InvoiceFilters) ([]models.Invoice, int64, error) {
	return nil, 0, nil
}
func (m *voidMockInvoiceRepo) ListByPatient(patientID string) ([]models.Invoice, error) {
	return nil, nil
}
func (m *voidMockInvoiceRepo) GetLastInvoiceNumber(prefix, yearMonth string) (string, error) {
	return "", nil
}
func (m *voidMockInvoiceRepo) GetOutstandingByPatient(patientID string) (int64, error) { return 0, nil }
func (m *voidMockInvoiceRepo) GetTotalOutstanding() (int64, error)                     { return 0, nil }
func (m *voidMockInvoiceRepo) GetRevenueByDateRange(startDate, endDate string) (int64, error) {
	return 0, nil
}
func (m *voidMockInvoiceRepo) GetTotalInvoicedByDateRange(startDate, endDate string) (int64, error) {
	return 0, nil
}
func (m *voidMockInvoiceRepo) GetOutstandingByDateRange(startDate, endDate string) (int64, error) {
	return 0, nil
}

func newTestInvoiceServiceForVoid(t *testing.T) (*InvoiceService, *voidMockInvoiceRepo) {
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
	db.AutoMigrate(&models.Invoice{})

	invoiceRepo := newVoidMockInvoiceRepo()
	auditService := newUnitTestAuditService()

	// Admin auth service
	authService := &AuthService{currentSession: testAdminSession()}

	svc := &InvoiceService{
		invoiceRepo:  invoiceRepo,
		authService:  authService,
		auditService: auditService,
		db:           db,
	}
	return svc, invoiceRepo
}

// --- VoidInvoice Tests ---

func TestInvoiceService_VoidInvoice_Success(t *testing.T) {
	svc, invoiceRepo := newTestInvoiceServiceForVoid(t)

	invoiceRepo.invoices["inv-1"] = &models.Invoice{
		BaseModel:     models.BaseModel{ID: "inv-1"},
		InvoiceNumber: "PV-202605-0001",
		Status:        models.InvoiceIssued,
		TotalAmount:   50000,
		PaidAmount:    0,
		BalanceAmount: 50000,
	}

	err := svc.VoidInvoice("inv-1", "Created by mistake")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	updated := invoiceRepo.invoices["inv-1"]
	if updated.Status != models.InvoiceVoid {
		t.Errorf("expected status 'void', got: %s", updated.Status)
	}
	if updated.VoidReason != "Created by mistake" {
		t.Errorf("expected reason 'Created by mistake', got: %s", updated.VoidReason)
	}
	if updated.BalanceAmount != 0 {
		t.Errorf("expected balance 0, got: %d", updated.BalanceAmount)
	}
}

func TestInvoiceService_VoidInvoice_AlreadyVoid(t *testing.T) {
	svc, invoiceRepo := newTestInvoiceServiceForVoid(t)

	invoiceRepo.invoices["inv-1"] = &models.Invoice{
		BaseModel: models.BaseModel{ID: "inv-1"},
		Status:    models.InvoiceVoid,
	}

	err := svc.VoidInvoice("inv-1", "Again")
	if err == nil {
		t.Fatal("expected error when voiding already voided invoice")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
}

func TestInvoiceService_VoidInvoice_HasPayments(t *testing.T) {
	svc, invoiceRepo := newTestInvoiceServiceForVoid(t)

	invoiceRepo.invoices["inv-1"] = &models.Invoice{
		BaseModel:     models.BaseModel{ID: "inv-1"},
		Status:        models.InvoicePartial,
		TotalAmount:   50000,
		PaidAmount:    20000,
		BalanceAmount: 30000,
	}

	err := svc.VoidInvoice("inv-1", "Want to void")
	if err == nil {
		t.Fatal("expected error when voiding invoice with payments")
	}
	var appErr *utils.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got: %T", err)
	}
}

func TestInvoiceService_VoidInvoice_EmptyReason(t *testing.T) {
	svc, invoiceRepo := newTestInvoiceServiceForVoid(t)

	invoiceRepo.invoices["inv-1"] = &models.Invoice{
		BaseModel:  models.BaseModel{ID: "inv-1"},
		Status:     models.InvoiceIssued,
		PaidAmount: 0,
	}

	err := svc.VoidInvoice("inv-1", "")
	if err == nil {
		t.Fatal("expected error for empty void reason")
	}
}

func TestInvoiceService_VoidInvoice_NotFound(t *testing.T) {
	svc, _ := newTestInvoiceServiceForVoid(t)

	err := svc.VoidInvoice("nonexistent", "reason")
	if err == nil {
		t.Fatal("expected error for nonexistent invoice")
	}
}

// --- RecordPayment Validation Tests ---

func TestInvoiceService_RecordPayment_InvalidMethod(t *testing.T) {
	svc, invoiceRepo := newTestInvoiceServiceForVoid(t)

	invoiceRepo.invoices["inv-1"] = &models.Invoice{
		BaseModel:     models.BaseModel{ID: "inv-1"},
		Status:        models.InvoiceIssued,
		TotalAmount:   50000,
		PaidAmount:    0,
		BalanceAmount: 50000,
	}

	_, err := svc.RecordPayment(RecordPaymentInput{
		InvoiceID:   "inv-1",
		Amount:      10000,
		Method:      "bitcoin", // Invalid
		PaymentDate: "2026-05-22",
	})

	if err == nil {
		t.Fatal("expected error for invalid payment method")
	}
}

func TestInvoiceService_RecordPayment_ZeroAmount(t *testing.T) {
	svc, invoiceRepo := newTestInvoiceServiceForVoid(t)

	invoiceRepo.invoices["inv-1"] = &models.Invoice{
		BaseModel:     models.BaseModel{ID: "inv-1"},
		Status:        models.InvoiceIssued,
		TotalAmount:   50000,
		PaidAmount:    0,
		BalanceAmount: 50000,
	}

	_, err := svc.RecordPayment(RecordPaymentInput{
		InvoiceID:   "inv-1",
		Amount:      0,
		Method:      "cash",
		PaymentDate: "2026-05-22",
	})

	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestInvoiceService_RecordPayment_VoidedInvoice(t *testing.T) {
	svc, invoiceRepo := newTestInvoiceServiceForVoid(t)

	invoiceRepo.invoices["inv-1"] = &models.Invoice{
		BaseModel: models.BaseModel{ID: "inv-1"},
		Status:    models.InvoiceVoid,
	}

	_, err := svc.RecordPayment(RecordPaymentInput{
		InvoiceID:   "inv-1",
		Amount:      10000,
		Method:      "cash",
		PaymentDate: "2026-05-22",
	})

	if err == nil {
		t.Fatal("expected error for voided invoice")
	}
}
