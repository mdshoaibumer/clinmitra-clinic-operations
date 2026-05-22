package repository

import (
	"clinmitra/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	UpdateLastLogin(id string) error
	Count() (int64, error)
}

type ClinicRepository interface {
	Get() (*models.ClinicSettings, error)
	Upsert(settings *models.ClinicSettings) error
	IsSetupComplete() (bool, error)
}

type PatientRepository interface {
	Create(patient *models.Patient) error
	FindByID(id string) (*models.Patient, error)
	Update(patient *models.Patient) error
	Delete(id string) error
	List(page, pageSize int, search string) ([]models.Patient, int64, error)
	FindByPhone(phone string) (*models.Patient, error)
	Count() (int64, error)
	CountSince(sinceDate string) (int64, error)
}

type TreatmentRepository interface {
	Create(treatment *models.Treatment) error
	FindByID(id string) (*models.Treatment, error)
	Update(treatment *models.Treatment) error
	Delete(id string) error
	ListActive() ([]models.Treatment, error)
	ListAll() ([]models.Treatment, error)
}

type AppointmentRepository interface {
	Create(appointment *models.Appointment) error
	FindByID(id string) (*models.Appointment, error)
	Update(appointment *models.Appointment) error
	Delete(id string) error
	ListByDate(date string) ([]models.Appointment, error)
	ListByDateRange(startDate, endDate string) ([]models.Appointment, error)
	ListByPatient(patientID string) ([]models.Appointment, error)
	FindConflicting(date, startTime, endTime, excludeID string) (*models.Appointment, error)
	CountByDate(date string) (int64, error)
}

type InvoiceRepository interface {
	Create(invoice *models.Invoice) error
	FindByID(id string) (*models.Invoice, error)
	Update(invoice *models.Invoice) error
	List(page, pageSize int, filters InvoiceFilters) ([]models.Invoice, int64, error)
	ListByPatient(patientID string) ([]models.Invoice, error)
	GetLastInvoiceNumber(prefix, yearMonth string) (string, error)
	GetOutstandingByPatient(patientID string) (int64, error)
	GetTotalOutstanding() (int64, error)
	GetRevenueByDateRange(startDate, endDate string) (int64, error)
	GetTotalInvoicedByDateRange(startDate, endDate string) (int64, error)
	GetOutstandingByDateRange(startDate, endDate string) (int64, error)
}

type InvoiceFilters struct {
	Status    string
	StartDate string
	EndDate   string
	PatientID string
	Search    string
}

type InvoiceItemRepository interface {
	CreateBatch(items []models.InvoiceItem) error
	FindByInvoiceID(invoiceID string) ([]models.InvoiceItem, error)
}

type PaymentRepository interface {
	Create(payment *models.Payment) error
	FindByInvoiceID(invoiceID string) ([]models.Payment, error)
	GetTotalByInvoice(invoiceID string) (int64, error)
	GetCollectionByDate(date string) (int64, error)
	GetCollectionByDateRange(startDate, endDate string) (int64, error)
	ListByDateRange(startDate, endDate string) ([]models.Payment, error)
}

type PatientTreatmentRepository interface {
	Create(pt *models.PatientTreatment) error
	CreateBatch(pts []models.PatientTreatment) error
	ListByPatient(patientID string) ([]models.PatientTreatment, error)
}

type AuditRepository interface {
	Create(log *models.AuditLog) error
	CreateTx(tx *gorm.DB, log *models.AuditLog) error
	ListByEntity(entityType, entityID string) ([]models.AuditLog, error)
	ListByUser(userID string, limit int) ([]models.AuditLog, error)
	ListRecent(limit int) ([]models.AuditLog, error)
}
