package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"clinmitra/internal/models"
	"clinmitra/internal/repository"

	"golang.org/x/sys/windows/registry"
)

// WhatsAppService handles WhatsApp message composition and delivery detection.
type WhatsAppService struct {
	clinicRepo repository.ClinicRepository
}

// NewWhatsAppService creates a WhatsAppService.
func NewWhatsAppService(clinicRepo repository.ClinicRepository) *WhatsAppService {
	return &WhatsAppService{clinicRepo: clinicRepo}
}

// WhatsAppTemplates holds the configurable message templates.
type WhatsAppTemplates struct {
	WelcomeTemplate string `json:"welcomeTemplate"`
	InvoiceTemplate string `json:"invoiceTemplate"`
	Enabled         bool   `json:"enabled"`
}

// WhatsAppMessageResult is returned when a message is prepared for sending.
type WhatsAppMessageResult struct {
	Phone            string `json:"phone"`
	Message          string `json:"message"`
	WhatsAppURL      string `json:"whatsAppUrl"`
	WebURL           string `json:"webUrl"`
	IsDesktopPresent bool   `json:"isDesktopPresent"`
}

// DefaultWelcomeTemplate is the default welcome message template.
const DefaultWelcomeTemplate = `Hello {{patient_name}}! 👋

Welcome to {{clinic_name}}! We're delighted to have you as our patient.

Dr. {{doctor_name}} and our team are here to provide you with the best dental care.

For appointments or queries, reach us at {{clinic_phone}}.

Thank you for choosing us! 🦷`

// DefaultInvoiceTemplate is the default invoice/payment message template.
const DefaultInvoiceTemplate = `Hi {{patient_name}},

Here's your payment receipt from {{clinic_name}}:

🧾 Invoice: {{invoice_number}}
📅 Date: {{invoice_date}}
💰 Total: ₹{{total_amount}}
✅ Paid: ₹{{paid_amount}}
📊 Balance: ₹{{balance_amount}}

Payment Method: {{payment_method}}

Thank you for your payment!
For any queries, call us at {{clinic_phone}}.`

// placeholderRegex matches {{placeholder_name}} patterns.
var placeholderRegex = regexp.MustCompile(`\{\{(\w+)\}\}`)

// RenderTemplate replaces placeholders in template with values from the data map.
func RenderTemplate(template string, data map[string]string) string {
	return placeholderRegex.ReplaceAllStringFunc(template, func(match string) string {
		key := match[2 : len(match)-2] // strip {{ and }}
		if val, ok := data[key]; ok {
			return val
		}
		return match // leave unresolved placeholders as-is
	})
}

// GetWhatsAppTemplates retrieves the current WhatsApp templates from settings.
func (s *WhatsAppService) GetWhatsAppTemplates() (*WhatsAppTemplates, error) {
	settings, err := s.clinicRepo.Get()
	if err != nil {
		return nil, err
	}
	templates := &WhatsAppTemplates{
		WelcomeTemplate: settings.WhatsAppWelcomeTemplate,
		InvoiceTemplate: settings.WhatsAppInvoiceTemplate,
		Enabled:         settings.WhatsAppEnabled,
	}
	if templates.WelcomeTemplate == "" {
		templates.WelcomeTemplate = DefaultWelcomeTemplate
	}
	if templates.InvoiceTemplate == "" {
		templates.InvoiceTemplate = DefaultInvoiceTemplate
	}
	return templates, nil
}

// PrepareWelcomeMessage renders a welcome message for a newly registered patient.
func (s *WhatsAppService) PrepareWelcomeMessage(patient *models.Patient) (*WhatsAppMessageResult, error) {
	settings, err := s.clinicRepo.Get()
	if err != nil {
		return nil, err
	}

	template := settings.WhatsAppWelcomeTemplate
	if template == "" {
		template = DefaultWelcomeTemplate
	}

	data := map[string]string{
		"patient_name": patient.Name,
		"clinic_name":  settings.ClinicName,
		"doctor_name":  settings.DoctorName,
		"clinic_phone": settings.Phone,
	}

	message := RenderTemplate(template, data)
	phone := cleanPhoneForWhatsApp(patient.Phone)

	return buildWhatsAppResult(phone, message), nil
}

// PrepareInvoiceMessage renders an invoice/payment message.
func (s *WhatsAppService) PrepareInvoiceMessage(invoice *models.Invoice, paymentMethod string) (*WhatsAppMessageResult, error) {
	settings, err := s.clinicRepo.Get()
	if err != nil {
		return nil, err
	}

	template := settings.WhatsAppInvoiceTemplate
	if template == "" {
		template = DefaultInvoiceTemplate
	}

	data := map[string]string{
		"patient_name":   invoice.Patient.Name,
		"clinic_name":    settings.ClinicName,
		"doctor_name":    settings.DoctorName,
		"clinic_phone":   settings.Phone,
		"invoice_number": invoice.InvoiceNumber,
		"invoice_date":   invoice.InvoiceDate,
		"total_amount":   formatPaiseToRupees(invoice.TotalAmount),
		"paid_amount":    formatPaiseToRupees(invoice.PaidAmount),
		"balance_amount": formatPaiseToRupees(invoice.BalanceAmount),
		"payment_method": paymentMethod,
	}

	message := RenderTemplate(template, data)
	phone := cleanPhoneForWhatsApp(invoice.Patient.Phone)

	return buildWhatsAppResult(phone, message), nil
}

// IsWhatsAppInstalled checks if WhatsApp Desktop is installed on Windows.
func IsWhatsAppInstalled() bool {
	// Check common WhatsApp Desktop installation paths
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData != "" {
		paths := []string{
			filepath.Join(localAppData, "WhatsApp", "WhatsApp.exe"),
			filepath.Join(localAppData, "Programs", "whatsapp-desktop", "WhatsApp.exe"),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return true
			}
		}
	}

	// Check if whatsapp:// protocol is registered in Windows Registry
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\whatsapp`, registry.READ)
	if err == nil {
		key.Close()
		return true
	}

	// Check HKLM as well (system-wide install)
	key, err = registry.OpenKey(registry.LOCAL_MACHINE, `Software\Classes\whatsapp`, registry.READ)
	if err == nil {
		key.Close()
		return true
	}

	return false
}

// cleanPhoneForWhatsApp converts a phone number to international format for WhatsApp.
// Indian numbers: 91XXXXXXXXXX (no +, no spaces)
func cleanPhoneForWhatsApp(phone string) string {
	// Remove all non-digit characters
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, phone)

	// If it's a 10-digit Indian number, prefix with 91
	if len(digits) == 10 {
		return "91" + digits
	}

	// If it already has country code (12 digits starting with 91)
	if len(digits) == 12 && strings.HasPrefix(digits, "91") {
		return digits
	}

	// Return as-is for other formats
	return digits
}

// formatPaiseToRupees converts paise (int64) to a formatted rupee string.
func formatPaiseToRupees(paise int64) string {
	rupees := float64(paise) / 100.0
	return fmt.Sprintf("%.2f", rupees)
}

// buildWhatsAppResult constructs the URLs for both desktop and web WhatsApp.
func buildWhatsAppResult(phone, message string) *WhatsAppMessageResult {
	encodedMsg := url.QueryEscape(message)

	return &WhatsAppMessageResult{
		Phone:            phone,
		Message:          message,
		WhatsAppURL:      fmt.Sprintf("whatsapp://send?phone=%s&text=%s", phone, encodedMsg),
		WebURL:           fmt.Sprintf("https://wa.me/%s?text=%s", phone, encodedMsg),
		IsDesktopPresent: IsWhatsAppInstalled(),
	}
}

// BuildWhatsAppMessage is a public wrapper for building a WhatsApp message result
// from a raw phone number and message string.
func BuildWhatsAppMessage(phone, message string) *WhatsAppMessageResult {
	cleanedPhone := cleanPhoneForWhatsApp(phone)
	return buildWhatsAppResult(cleanedPhone, message)
}
