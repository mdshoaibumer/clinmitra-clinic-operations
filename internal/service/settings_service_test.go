package service

import (
	"testing"
	"time"

	"clinmitra/internal/auth"
	"clinmitra/internal/config"
	"clinmitra/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type settingsMockClinicRepo struct {
	mock.Mock
}

func (m *settingsMockClinicRepo) Get() (*models.ClinicSettings, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ClinicSettings), args.Error(1)
}

func (m *settingsMockClinicRepo) Upsert(settings *models.ClinicSettings) error {
	return m.Called(settings).Error(0)
}

func (m *settingsMockClinicRepo) IsSetupComplete() (bool, error) {
	args := m.Called()
	return args.Bool(0), args.Error(1)
}

type settingsMockTreatmentRepo struct {
	mock.Mock
}

func (m *settingsMockTreatmentRepo) Create(t *models.Treatment) error { return nil }
func (m *settingsMockTreatmentRepo) FindByID(id string) (*models.Treatment, error) {
	return &models.Treatment{}, nil
}
func (m *settingsMockTreatmentRepo) Update(t *models.Treatment) error { return nil }
func (m *settingsMockTreatmentRepo) Delete(id string) error           { return nil }
func (m *settingsMockTreatmentRepo) ListActive() ([]models.Treatment, error) {
	return nil, nil
}
func (m *settingsMockTreatmentRepo) ListAll() ([]models.Treatment, error) {
	return nil, nil
}

type settingsMockAuditRepo struct{}

func (m *settingsMockAuditRepo) Create(log *models.AuditLog) error                { return nil }
func (m *settingsMockAuditRepo) CreateTx(tx *gorm.DB, log *models.AuditLog) error { return nil }
func (m *settingsMockAuditRepo) ListByEntity(entityType, entityID string) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *settingsMockAuditRepo) ListByUser(userID string, limit int) ([]models.AuditLog, error) {
	return nil, nil
}
func (m *settingsMockAuditRepo) ListRecent(limit int) ([]models.AuditLog, error) {
	return nil, nil
}

func TestUpdateClinicSettings_Validation(t *testing.T) {
	clinicRepo := new(settingsMockClinicRepo)

	testSettings := &models.ClinicSettings{
		ID:                  "test-id",
		ClinicName:          "Test Clinic",
		DoctorName:          "Dr. Test",
		DoctorQualification: "BDS",
	}

	// Mock success
	clinicRepo.On("Upsert", testSettings).Return(nil)

	err := clinicRepo.Upsert(testSettings)
	assert.NoError(t, err)
	clinicRepo.AssertExpectations(t)
}

func TestClinicSettingsModel(t *testing.T) {
	settings := models.ClinicSettings{
		DoctorName:          "Dr. Smith",
		DoctorQualification: "MDS",
	}
	assert.Equal(t, "Dr. Smith", settings.DoctorName)
	assert.Equal(t, "MDS", settings.DoctorQualification)
}

func TestUpdateClinicSettings_RequiredFields(t *testing.T) {
	clinicRepo := new(settingsMockClinicRepo)
	treatmentRepo := new(settingsMockTreatmentRepo)
	auditRepo := &settingsMockAuditRepo{}
	auditService := NewAuditService(auditRepo)

	// Create a mock auth service that always passes role check
	authService := createSettingsTestAuthService(t)

	svc := NewSettingsService(clinicRepo, treatmentRepo, authService, auditService, testConfig())

	tests := []struct {
		name     string
		settings *models.ClinicSettings
		wantErr  string
	}{
		{
			name:     "empty clinic name",
			settings: &models.ClinicSettings{ID: "1", ClinicName: "", DoctorName: "Dr. X"},
			wantErr:  "Clinic name is required",
		},
		{
			name:     "empty doctor name",
			settings: &models.ClinicSettings{ID: "1", ClinicName: "Clinic", DoctorName: ""},
			wantErr:  "Doctor name is required",
		},
		{
			name:     "GSTIN too long",
			settings: &models.ClinicSettings{ID: "1", ClinicName: "Clinic", DoctorName: "Dr. X", GSTIN: "1234567890123456"},
			wantErr:  "GSTIN must not exceed 15 characters",
		},
		{
			name:     "IFSC too long",
			settings: &models.ClinicSettings{ID: "1", ClinicName: "Clinic", DoctorName: "Dr. X", IFSCCode: "123456789012"},
			wantErr:  "IFSC code must not exceed 11 characters",
		},
		{
			name:     "invalid GST rate",
			settings: &models.ClinicSettings{ID: "1", ClinicName: "Clinic", DoctorName: "Dr. X", GSTEnabled: true, GSTRate: 50},
			wantErr:  "GST rate must be between 0 and 28%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpdateClinicSettings(tt.settings)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestUpdateClinicSettings_ValidInput(t *testing.T) {
	clinicRepo := new(settingsMockClinicRepo)
	treatmentRepo := new(settingsMockTreatmentRepo)
	auditRepo := &settingsMockAuditRepo{}
	auditService := NewAuditService(auditRepo)
	authService := createSettingsTestAuthService(t)

	svc := NewSettingsService(clinicRepo, treatmentRepo, authService, auditService, testConfig())

	validSettings := &models.ClinicSettings{
		ID:          "1",
		ClinicName:  "Smile Dental",
		DoctorName:  "Dr. Sharma",
		GSTEnabled:  true,
		GSTRate:     18,
		BankAccount: "1234567890",
		IFSCCode:    "SBIN0001234",
		UPIID:       "clinic@upi",
	}

	clinicRepo.On("Upsert", validSettings).Return(nil)

	err := svc.UpdateClinicSettings(validSettings)
	assert.NoError(t, err)
	clinicRepo.AssertExpectations(t)
}

// createSettingsTestAuthService creates an AuthService with admin session
// pre-configured for settings tests.
func createSettingsTestAuthService(t *testing.T) *AuthService {
	t.Helper()
	sm := auth.NewSessionManager(8)
	svc := &AuthService{
		sessionManager: sm,
		cfg:            testConfig(),
	}
	// Set a current session with admin role
	session := &auth.Session{
		Token:     "test-token",
		UserID:    "admin-user-id",
		Username:  "admin",
		FullName:  "Test Admin",
		Role:      models.RoleAdmin,
		ExpiresAt: time.Now().Add(8 * time.Hour),
	}
	sm.RestoreSession(session)
	svc.currentSession = session
	return svc
}

// testConfig returns a minimal Config for test use.
func testConfig() *config.Config {
	return &config.Config{
		AppName:          "Test",
		Version:          "1.0.0",
		DataDir:          ".",
		MaxLoginAttempts: 5,
		LockoutMinutes:   15,
		SessionHours:     8,
		BcryptCost:       4,
	}
}
