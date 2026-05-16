package handler

import (
	"clinmitra/internal/models"
	"clinmitra/internal/service"
)

type InvoiceHandler struct {
	invoiceService *service.InvoiceService
}

// NewInvoiceHandler creates an InvoiceHandler backed by the given service.
func NewInvoiceHandler(invoiceService *service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

// CreateInvoice generates a new invoice with line items, GST calculation,
// and treatment history recording.
func (h *InvoiceHandler) CreateInvoice(input service.CreateInvoiceInput) (*models.Invoice, error) {
	result, err := h.invoiceService.CreateInvoice(input)
	return result, safeError(err)
}

// GetInvoice retrieves a single invoice by ID with items and payments.
func (h *InvoiceHandler) GetInvoice(id string) (*models.Invoice, error) {
	result, err := h.invoiceService.GetInvoice(id)
	return result, safeError(err)
}

// ListInvoices returns a paginated, filterable list of invoices.
func (h *InvoiceHandler) ListInvoices(page, pageSize int, status, startDate, endDate, patientID, search string) (*service.InvoiceListResponse, error) {
	page, pageSize = sanitizePagination(page, pageSize)
	search = sanitizeSearch(search)
	result, err := h.invoiceService.ListInvoices(page, pageSize, status, startDate, endDate, patientID, search)
	return result, safeError(err)
}

// RecordPayment records a payment against an invoice and updates its status.
func (h *InvoiceHandler) RecordPayment(input service.RecordPaymentInput) (*models.Payment, error) {
	result, err := h.invoiceService.RecordPayment(input)
	return result, safeError(err)
}

// VoidInvoice marks an unpaid invoice as void with a reason.
func (h *InvoiceHandler) VoidInvoice(id, reason string) error {
	return safeError(h.invoiceService.VoidInvoice(id, reason))
}

// GetPatientOutstanding returns the total unpaid balance for a patient.
func (h *InvoiceHandler) GetPatientOutstanding(patientID string) (int64, error) {
	result, err := h.invoiceService.GetPatientOutstanding(patientID)
	return result, safeError(err)
}

// GetPatientInvoices returns all invoices for a specific patient.
func (h *InvoiceHandler) GetPatientInvoices(patientID string) ([]models.Invoice, error) {
	result, err := h.invoiceService.GetPatientInvoices(patientID)
	return result, safeError(err)
}
