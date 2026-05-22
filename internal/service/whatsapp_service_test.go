package service

import (
	"testing"

	"clinmitra/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]string
		expected string
	}{
		{
			name:     "simple replacement",
			template: "Hello {{patient_name}}!",
			data:     map[string]string{"patient_name": "John"},
			expected: "Hello John!",
		},
		{
			name:     "multiple replacements",
			template: "Hi {{patient_name}}, welcome to {{clinic_name}}!",
			data:     map[string]string{"patient_name": "Priya", "clinic_name": "Smile Dental"},
			expected: "Hi Priya, welcome to Smile Dental!",
		},
		{
			name:     "unresolved placeholder remains",
			template: "Hello {{patient_name}}, your ID is {{patient_id}}.",
			data:     map[string]string{"patient_name": "Raj"},
			expected: "Hello Raj, your ID is {{patient_id}}.",
		},
		{
			name:     "no placeholders",
			template: "Hello World!",
			data:     map[string]string{"patient_name": "Test"},
			expected: "Hello World!",
		},
		{
			name:     "empty template",
			template: "",
			data:     map[string]string{"patient_name": "Test"},
			expected: "",
		},
		{
			name:     "all invoice fields",
			template: "Invoice {{invoice_number}} | Total: ₹{{total_amount}} | Paid: ₹{{paid_amount}} | Balance: ₹{{balance_amount}}",
			data: map[string]string{
				"invoice_number": "PV-001",
				"total_amount":   "5000.00",
				"paid_amount":    "3000.00",
				"balance_amount": "2000.00",
			},
			expected: "Invoice PV-001 | Total: ₹5000.00 | Paid: ₹3000.00 | Balance: ₹2000.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTemplate(tt.template, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanPhoneForWhatsApp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "10 digit Indian number",
			input:    "9876543210",
			expected: "919876543210",
		},
		{
			name:     "with +91 prefix",
			input:    "+919876543210",
			expected: "919876543210",
		},
		{
			name:     "with spaces",
			input:    "98765 43210",
			expected: "919876543210",
		},
		{
			name:     "with dashes",
			input:    "987-654-3210",
			expected: "919876543210",
		},
		{
			name:     "already 12 digits with 91",
			input:    "919876543210",
			expected: "919876543210",
		},
		{
			name:     "with +91 and spaces",
			input:    "+91 98765 43210",
			expected: "919876543210",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPhoneForWhatsApp(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatPaiseToRupees(t *testing.T) {
	tests := []struct {
		name     string
		paise    int64
		expected string
	}{
		{"zero", 0, "0.00"},
		{"100 paise = 1 rupee", 100, "1.00"},
		{"5000 rupees", 500000, "5000.00"},
		{"with paise fraction", 15050, "150.50"},
		{"large amount", 10000000, "100000.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPaiseToRupees(tt.paise)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildWhatsAppResult(t *testing.T) {
	result := buildWhatsAppResult("919876543210", "Hello World!")

	assert.Equal(t, "919876543210", result.Phone)
	assert.Equal(t, "Hello World!", result.Message)
	assert.Contains(t, result.WhatsAppURL, "whatsapp://send?phone=919876543210")
	assert.Contains(t, result.WhatsAppURL, "text=Hello+World")
	assert.Contains(t, result.WebURL, "https://wa.me/919876543210")
	assert.Contains(t, result.WebURL, "text=Hello+World")
}

// whatsAppMockClinicRepo implements ClinicRepository for testing WhatsApp service.
type whatsAppMockClinicRepo struct {
	settings *models.ClinicSettings
	getErr   error
}

func (m *whatsAppMockClinicRepo) Get() (*models.ClinicSettings, error) {
	return m.settings, m.getErr
}

func (m *whatsAppMockClinicRepo) Upsert(settings *models.ClinicSettings) error {
	m.settings = settings
	return nil
}

func (m *whatsAppMockClinicRepo) IsSetupComplete() (bool, error) {
	return m.settings != nil && m.settings.SetupComplete, nil
}

func TestWhatsAppService_PrepareWelcomeMessage(t *testing.T) {
	repo := &whatsAppMockClinicRepo{
		settings: &models.ClinicSettings{
			ClinicName:              "Smile Dental Clinic",
			DoctorName:              "Dr. Sharma",
			Phone:                   "9876543210",
			WhatsAppEnabled:         true,
			WhatsAppWelcomeTemplate: "Hello {{patient_name}}, welcome to {{clinic_name}}! Call {{clinic_phone}}.",
		},
	}

	svc := NewWhatsAppService(repo)

	patient := &models.Patient{
		Name:  "Priya Patel",
		Phone: "8765432109",
	}

	result, err := svc.PrepareWelcomeMessage(patient)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "918765432109", result.Phone)
	assert.Contains(t, result.Message, "Priya Patel")
	assert.Contains(t, result.Message, "Smile Dental Clinic")
	assert.Contains(t, result.Message, "9876543210")
	assert.Contains(t, result.WhatsAppURL, "whatsapp://send")
	assert.Contains(t, result.WebURL, "https://wa.me/918765432109")
}

func TestWhatsAppService_PrepareWelcomeMessage_DefaultTemplate(t *testing.T) {
	repo := &whatsAppMockClinicRepo{
		settings: &models.ClinicSettings{
			ClinicName:              "Test Clinic",
			DoctorName:              "Dr. Test",
			Phone:                   "9000000000",
			WhatsAppEnabled:         true,
			WhatsAppWelcomeTemplate: "", // Empty - should use default
		},
	}

	svc := NewWhatsAppService(repo)

	patient := &models.Patient{
		Name:  "Test Patient",
		Phone: "8000000000",
	}

	result, err := svc.PrepareWelcomeMessage(patient)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.Message, "Test Patient")
	assert.Contains(t, result.Message, "Test Clinic")
	assert.Contains(t, result.Message, "Dr. Test")
	assert.Contains(t, result.Message, "9000000000")
}

func TestWhatsAppService_PrepareInvoiceMessage(t *testing.T) {
	repo := &whatsAppMockClinicRepo{
		settings: &models.ClinicSettings{
			ClinicName:              "Dental Care",
			DoctorName:              "Dr. Kumar",
			Phone:                   "9111111111",
			WhatsAppEnabled:         true,
			WhatsAppInvoiceTemplate: "Hi {{patient_name}}, Invoice {{invoice_number}}: ₹{{total_amount}}. Paid: ₹{{paid_amount}}. Balance: ₹{{balance_amount}}. Method: {{payment_method}}.",
		},
	}

	svc := NewWhatsAppService(repo)

	invoice := &models.Invoice{
		InvoiceNumber: "PV-042",
		InvoiceDate:   "2026-05-22",
		TotalAmount:   500000, // 5000.00
		PaidAmount:    300000, // 3000.00
		BalanceAmount: 200000, // 2000.00
		Patient: models.Patient{
			Name:  "Ramesh Gupta",
			Phone: "7777777777",
		},
	}

	result, err := svc.PrepareInvoiceMessage(invoice, "UPI")
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "917777777777", result.Phone)
	assert.Contains(t, result.Message, "Ramesh Gupta")
	assert.Contains(t, result.Message, "PV-042")
	assert.Contains(t, result.Message, "5000.00")
	assert.Contains(t, result.Message, "3000.00")
	assert.Contains(t, result.Message, "2000.00")
	assert.Contains(t, result.Message, "UPI")
}

func TestWhatsAppService_GetWhatsAppTemplates(t *testing.T) {
	t.Run("returns custom templates", func(t *testing.T) {
		repo := &whatsAppMockClinicRepo{
			settings: &models.ClinicSettings{
				WhatsAppEnabled:         true,
				WhatsAppWelcomeTemplate: "Custom welcome",
				WhatsAppInvoiceTemplate: "Custom invoice",
			},
		}
		svc := NewWhatsAppService(repo)

		templates, err := svc.GetWhatsAppTemplates()
		require.NoError(t, err)
		assert.True(t, templates.Enabled)
		assert.Equal(t, "Custom welcome", templates.WelcomeTemplate)
		assert.Equal(t, "Custom invoice", templates.InvoiceTemplate)
	})

	t.Run("returns defaults when templates empty", func(t *testing.T) {
		repo := &whatsAppMockClinicRepo{
			settings: &models.ClinicSettings{
				WhatsAppEnabled:         true,
				WhatsAppWelcomeTemplate: "",
				WhatsAppInvoiceTemplate: "",
			},
		}
		svc := NewWhatsAppService(repo)

		templates, err := svc.GetWhatsAppTemplates()
		require.NoError(t, err)
		assert.True(t, templates.Enabled)
		assert.Equal(t, DefaultWelcomeTemplate, templates.WelcomeTemplate)
		assert.Equal(t, DefaultInvoiceTemplate, templates.InvoiceTemplate)
	})

	t.Run("returns error on repo failure", func(t *testing.T) {
		repo := &whatsAppMockClinicRepo{
			getErr: assert.AnError,
		}
		svc := NewWhatsAppService(repo)

		templates, err := svc.GetWhatsAppTemplates()
		assert.Error(t, err)
		assert.Nil(t, templates)
	})
}
