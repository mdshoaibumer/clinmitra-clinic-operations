package handler

import (
	"practivo/internal/service"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

// NewDashboardHandler creates a DashboardHandler backed by the given service.
func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetDashboardStats returns aggregated statistics for the dashboard
// (today's appointments, revenue, total patients, outstanding balance).
func (h *DashboardHandler) GetDashboardStats() (*service.DashboardStats, error) {
	result, err := h.dashboardService.GetDashboardStats()
	return result, safeError(err)
}

// GetDailyReport returns a daily collection report with payment details.
func (h *DashboardHandler) GetDailyReport(date string) (*service.DailyReport, error) {
	result, err := h.dashboardService.GetDailyReport(date)
	return result, safeError(err)
}

// GetMonthlyReport returns a monthly revenue and outstanding summary.
func (h *DashboardHandler) GetMonthlyReport(year, month int) (*service.MonthlyReport, error) {
	result, err := h.dashboardService.GetMonthlyReport(year, month)
	return result, safeError(err)
}
