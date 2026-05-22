package handler

import (
	"context"
	"log/slog"
	"strings"

	"clinmitra/internal/models"
	"clinmitra/internal/service"
	"clinmitra/internal/utils"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// WhatsAppHandler provides Wails-bound methods for WhatsApp messaging.
type WhatsAppHandler struct {
	whatsappService *service.WhatsAppService
	invoiceService  *service.InvoiceService
	ctx             context.Context
}

// NewWhatsAppHandler creates a WhatsAppHandler.
func NewWhatsAppHandler(whatsappService *service.WhatsAppService, invoiceService *service.InvoiceService) *WhatsAppHandler {
	return &WhatsAppHandler{
		whatsappService: whatsappService,
		invoiceService:  invoiceService,
	}
}

// SetContext stores the Wails runtime context for browser URL opening.
func (h *WhatsAppHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

// GetWhatsAppTemplates returns the current WhatsApp message templates.
func (h *WhatsAppHandler) GetWhatsAppTemplates() (*service.WhatsAppTemplates, error) {
	result, err := h.whatsappService.GetWhatsAppTemplates()
	return result, safeError(err)
}

// PrepareWelcomeMessage renders a welcome message for a patient.
func (h *WhatsAppHandler) PrepareWelcomeMessage(patient *models.Patient) (*service.WhatsAppMessageResult, error) {
	slog.Info("preparing WhatsApp welcome message", "patient", patient.Name)
	result, err := h.whatsappService.PrepareWelcomeMessage(patient)
	return result, safeError(err)
}

// PrepareInvoiceMessage renders an invoice payment message.
func (h *WhatsAppHandler) PrepareInvoiceMessage(invoiceID string, paymentMethod string) (*service.WhatsAppMessageResult, error) {
	slog.Info("preparing WhatsApp invoice message", "invoiceId", invoiceID)
	invoice, err := h.invoiceService.GetInvoice(invoiceID)
	if err != nil {
		return nil, safeError(err)
	}
	result, err := h.whatsappService.PrepareInvoiceMessage(invoice, paymentMethod)
	return result, safeError(err)
}

// IsWhatsAppInstalled checks if WhatsApp Desktop is available locally.
func (h *WhatsAppHandler) IsWhatsAppInstalled() bool {
	return service.IsWhatsAppInstalled()
}

// OpenWhatsApp opens the WhatsApp URL using the appropriate method.
// If desktop app is installed, uses whatsapp:// protocol.
// Otherwise opens wa.me link in the default browser.
func (h *WhatsAppHandler) OpenWhatsApp(phone string, message string) error {
	slog.Info("opening WhatsApp", "phone", phone)

	// Build the message result directly from the provided phone and message
	result := service.BuildWhatsAppMessage(phone, message)

	var targetURL string
	if result.IsDesktopPresent {
		targetURL = result.WhatsAppURL
	} else {
		targetURL = result.WebURL
	}

	if h.ctx != nil {
		wailsRuntime.BrowserOpenURL(h.ctx, targetURL)
	}
	return nil
}

// SendViaWhatsApp opens WhatsApp with the given pre-composed URL.
// Only allows whatsapp:// and https://wa.me/ URLs to prevent URL injection.
func (h *WhatsAppHandler) SendViaWhatsApp(waURL string) error {
	if err := validateWhatsAppURL(waURL); err != nil {
		slog.Warn("rejected invalid WhatsApp URL", "error", err.Error())
		return safeError(err)
	}
	slog.Info("sending via WhatsApp", "url_prefix", waURL[:min(len(waURL), 30)])
	if h.ctx != nil {
		wailsRuntime.BrowserOpenURL(h.ctx, waURL)
	}
	return nil
}

// validateWhatsAppURL ensures only safe WhatsApp URLs are opened.
func validateWhatsAppURL(u string) error {
	if strings.HasPrefix(u, "whatsapp://send?") {
		return nil
	}
	if strings.HasPrefix(u, "https://wa.me/") {
		return nil
	}
	return utils.ValidationError("Invalid WhatsApp URL: only whatsapp:// and https://wa.me/ URLs are allowed")
}
