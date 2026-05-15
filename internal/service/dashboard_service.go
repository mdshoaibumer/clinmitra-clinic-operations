package service

import (
	"fmt"
	"time"

	"practivo/internal/repository"
	"practivo/internal/utils"
)

type DashboardStats struct {
	TodayAppointments int64 `json:"todayAppointments"`
	TotalPatients     int64 `json:"totalPatients"`
	TodayRevenue      int64 `json:"todayRevenue"`     // paise
	MonthRevenue      int64 `json:"monthRevenue"`     // paise
	TotalOutstanding  int64 `json:"totalOutstanding"` // paise
	PatientsThisMonth int64 `json:"patientsThisMonth"`
}

type DailyReport struct {
	Date            string           `json:"date"`
	TotalCollection int64            `json:"totalCollection"` // paise
	Payments        []PaymentSummary `json:"payments"`
}

type PaymentSummary struct {
	InvoiceNumber string `json:"invoiceNumber"`
	PatientName   string `json:"patientName"`
	Amount        int64  `json:"amount"` // paise
	Method        string `json:"method"`
}

type MonthlyReport struct {
	Year             int   `json:"year"`
	Month            int   `json:"month"`
	TotalRevenue     int64 `json:"totalRevenue"`     // paise
	TotalInvoiced    int64 `json:"totalInvoiced"`    // paise
	TotalOutstanding int64 `json:"totalOutstanding"` // paise
}

type OutstandingEntry struct {
	PatientID   string `json:"patientId"`
	PatientName string `json:"patientName"`
	Phone       string `json:"phone"`
	Amount      int64  `json:"amount"` // paise
}

type DashboardService struct {
	invoiceRepo     repository.InvoiceRepository
	paymentRepo     repository.PaymentRepository
	appointmentRepo repository.AppointmentRepository
	patientRepo     repository.PatientRepository
}

// NewDashboardService creates a DashboardService for aggregating statistics
// and generating financial reports.
func NewDashboardService(
	invoiceRepo repository.InvoiceRepository,
	paymentRepo repository.PaymentRepository,
	appointmentRepo repository.AppointmentRepository,
	patientRepo repository.PatientRepository,
) *DashboardService {
	return &DashboardService{
		invoiceRepo:     invoiceRepo,
		paymentRepo:     paymentRepo,
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
	}
}

// GetDashboardStats returns aggregated metrics for the dashboard:
// today's appointments, total patients, today/month revenue, and outstanding.
// Queries run sequentially since SQLite uses a single connection.
func (s *DashboardService) GetDashboardStats() (*DashboardStats, error) {
	today := utils.TodayDate()
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")

	todayAppointments, err := s.appointmentRepo.CountByDate(today)
	if err != nil {
		return nil, fmt.Errorf("dashboard stats failed: %w", err)
	}

	totalPatients, err := s.patientRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("dashboard stats failed: %w", err)
	}

	todayRevenue, err := s.paymentRepo.GetCollectionByDate(today)
	if err != nil {
		return nil, fmt.Errorf("dashboard stats failed: %w", err)
	}

	monthRevenue, err := s.paymentRepo.GetCollectionByDateRange(monthStart, today)
	if err != nil {
		return nil, fmt.Errorf("dashboard stats failed: %w", err)
	}

	totalOutstanding, err := s.invoiceRepo.GetTotalOutstanding()
	if err != nil {
		return nil, fmt.Errorf("dashboard stats failed: %w", err)
	}

	return &DashboardStats{
		TodayAppointments: todayAppointments,
		TotalPatients:     totalPatients,
		TodayRevenue:      todayRevenue,
		MonthRevenue:      monthRevenue,
		TotalOutstanding:  totalOutstanding,
	}, nil
}

// GetDailyReport returns a collection report for a specific date,
// including individual payment details with invoice and patient info.
func (s *DashboardService) GetDailyReport(date string) (*DailyReport, error) {
	if date == "" {
		date = utils.TodayDate()
	}

	totalCollection, err := s.paymentRepo.GetCollectionByDate(date)
	if err != nil {
		return nil, err
	}

	payments, err := s.paymentRepo.ListByDateRange(date, date)
	if err != nil {
		return nil, err
	}

	summaries := make([]PaymentSummary, len(payments))
	for i, p := range payments {
		summaries[i] = PaymentSummary{
			InvoiceNumber: p.Invoice.InvoiceNumber,
			PatientName:   p.Invoice.Patient.Name,
			Amount:        p.Amount,
			Method:        string(p.Method),
		}
	}

	return &DailyReport{
		Date:            date,
		TotalCollection: totalCollection,
		Payments:        summaries,
	}, nil
}

// GetMonthlyReport returns revenue and outstanding totals for a given month.
func (s *DashboardService) GetMonthlyReport(year, month int) (*MonthlyReport, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	endDate := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.Local).Format("2006-01-02")

	revenue, err := s.paymentRepo.GetCollectionByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	outstanding, err := s.invoiceRepo.GetTotalOutstanding()
	if err != nil {
		return nil, err
	}

	return &MonthlyReport{
		Year:             year,
		Month:            month,
		TotalRevenue:     revenue,
		TotalOutstanding: outstanding,
	}, nil
}
