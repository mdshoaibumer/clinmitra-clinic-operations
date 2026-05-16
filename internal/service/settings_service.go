package service

import (
	"clinmitra/internal/config"
	"clinmitra/internal/models"
	"clinmitra/internal/repository"
	"clinmitra/internal/utils"

	"github.com/google/uuid"
)

type SetupInput struct {
	ClinicName         string `json:"clinicName"`
	DoctorName         string `json:"doctorName"`
	DoctorQualification string `json:"doctorQualification"`
	Address            string `json:"address"`
	City               string `json:"city"`
	State              string `json:"state"`
	Pincode            string `json:"pincode"`
	Phone              string `json:"phone"`
	Email              string `json:"email"`
	GSTIN              string `json:"gstin"`
	GSTEnabled         bool   `json:"gstEnabled"`
	InvoicePrefix      string `json:"invoicePrefix"`
	AdminUsername      string `json:"adminUsername"`
	AdminPassword      string `json:"adminPassword"`
	AdminFullName      string `json:"adminFullName"`
}

type SettingsService struct {
	clinicRepo    repository.ClinicRepository
	treatmentRepo repository.TreatmentRepository
	authService   *AuthService
	auditService  *AuditService
	cfg           *config.Config
}

// NewSettingsService creates a SettingsService for managing clinic configuration
// and treatment catalog.
func NewSettingsService(
	clinicRepo repository.ClinicRepository,
	treatmentRepo repository.TreatmentRepository,
	authService *AuthService,
	auditService *AuditService,
	cfg *config.Config,
) *SettingsService {
	return &SettingsService{
		clinicRepo:    clinicRepo,
		treatmentRepo: treatmentRepo,
		authService:   authService,
		auditService:  auditService,
		cfg:           cfg,
	}
}

// IsSetupComplete checks whether the initial clinic setup has been done.
func (s *SettingsService) IsSetupComplete() (bool, error) {
	return s.clinicRepo.IsSetupComplete()
}

// CompleteSetup performs first-time setup: validates input, creates the admin
// user, and saves clinic settings. Fails if setup was already completed.
func (s *SettingsService) CompleteSetup(input SetupInput) error {
	done, err := s.clinicRepo.IsSetupComplete()
	if err != nil {
		return err
	}
	if done {
		return utils.ErrSetupAlreadyDone
	}

	// Validate
	if err := utils.ValidateRequired("Clinic name", input.ClinicName); err != nil {
		return err
	}
	if err := utils.ValidateRequired("Doctor name", input.DoctorName); err != nil {
		return err
	}
	if err := utils.ValidateRequired("Phone", input.Phone); err != nil {
		return err
	}

	// Create admin user
	if err := s.authService.CreateInitialAdmin(input.AdminUsername, input.AdminPassword, input.AdminFullName); err != nil {
		return err
	}

	// Save clinic settings
	prefix := input.InvoicePrefix
	if prefix == "" {
		prefix = "PV"
	}

	settings := &models.ClinicSettings{
		ID:                 uuid.New().String(),
		ClinicName:         input.ClinicName,
		DoctorName:         input.DoctorName,
		DoctorQualification: input.DoctorQualification,
		Address:            input.Address,
		City:               input.City,
		State:              input.State,
		Pincode:            input.Pincode,
		Phone:              input.Phone,
		Email:              input.Email,
		GSTIN:              input.GSTIN,
		GSTEnabled:         input.GSTEnabled,
		GSTRate:            18,
		InvoicePrefix:      prefix,
		SetupComplete:      true,
		AutoBackup:         true,
		BackupPath:         s.cfg.BackupDir,
	}

	return s.clinicRepo.Upsert(settings)
}

// GetClinicSettings returns the current clinic configuration.
func (s *SettingsService) GetClinicSettings() (*models.ClinicSettings, error) {
	return s.clinicRepo.Get()
}

// UpdateClinicSettings saves updated clinic settings and logs the change.
// Requires admin role.
func (s *SettingsService) UpdateClinicSettings(settings *models.ClinicSettings) error {
	if err := s.authService.RequireRole(models.RoleAdmin); err != nil {
		return err
	}
	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditUpdate, "clinic_settings", settings.ID, nil, settings)
	return s.clinicRepo.Upsert(settings)
}

// ListTreatments returns all active dental treatments.
func (s *SettingsService) ListTreatments() ([]models.Treatment, error) {
	return s.treatmentRepo.ListActive()
}

// ListAllTreatments returns all treatments including deactivated ones.
func (s *SettingsService) ListAllTreatments() ([]models.Treatment, error) {
	return s.treatmentRepo.ListAll()
}

// CreateTreatment validates and creates a new treatment/procedure entry.
// Requires admin or doctor role.
func (s *SettingsService) CreateTreatment(name, code, category, description string, defaultPrice int64) (*models.Treatment, error) {
	if err := s.authService.RequireRole(models.RoleAdmin, models.RoleDoctor); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired("Name", name); err != nil {
		return nil, err
	}
	// Code is optional
	if code == "" {
		code = "CUSTOM"
	}
	if err := utils.ValidatePositiveAmount("Default price", defaultPrice); err != nil {
		return nil, err
	}

	treatment := &models.Treatment{
		ID:           uuid.New().String(),
		Name:         name,
		Code:         code,
		DefaultPrice: defaultPrice,
		Category:     category,
		Description:  description,
		IsActive:     true,
	}

	if err := s.treatmentRepo.Create(treatment); err != nil {
		return nil, err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditCreate, "treatment", treatment.ID, nil, treatment)
	return treatment, nil
}

// UpdateTreatment modifies an existing treatment's details and logs the change.
// Requires admin or doctor role.
func (s *SettingsService) UpdateTreatment(id, name, code, category, description string, defaultPrice int64) error {
	if err := s.authService.RequireRole(models.RoleAdmin, models.RoleDoctor); err != nil {
		return err
	}
	treatment, err := s.treatmentRepo.FindByID(id)
	if err != nil {
		return utils.ErrNotFound
	}

	old := *treatment
	treatment.Name = name
	treatment.Code = code
	treatment.Category = category
	treatment.Description = description
	treatment.DefaultPrice = defaultPrice

	if err := s.treatmentRepo.Update(treatment); err != nil {
		return err
	}

	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditUpdate, "treatment", id, old, treatment)
	return nil
}

// DeleteTreatment soft-deletes a treatment by marking it inactive.
// Requires admin role.
func (s *SettingsService) DeleteTreatment(id string) error {
	if err := s.authService.RequireRole(models.RoleAdmin); err != nil {
		return err
	}
	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditDelete, "treatment", id, nil, nil)
	return s.treatmentRepo.Delete(id)
}

// SaveLogo stores the base64-encoded logo in clinic settings.
// Requires admin role.
func (s *SettingsService) SaveLogo(base64Data string) error {
	if err := s.authService.RequireRole(models.RoleAdmin); err != nil {
		return err
	}
	settings, err := s.clinicRepo.Get()
	if err != nil {
		return err
	}
	settings.LogoBase64 = base64Data
	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditUpdate, "clinic_logo", settings.ID, nil, "logo_updated")
	return s.clinicRepo.Upsert(settings)
}

// RemoveLogo clears the logo from clinic settings.
// Requires admin role.
func (s *SettingsService) RemoveLogo() error {
	if err := s.authService.RequireRole(models.RoleAdmin); err != nil {
		return err
	}
	settings, err := s.clinicRepo.Get()
	if err != nil {
		return err
	}
	settings.LogoBase64 = ""
	s.auditService.LogAction(s.authService.GetCurrentUserID(), models.AuditUpdate, "clinic_logo", settings.ID, nil, "logo_removed")
	return s.clinicRepo.Upsert(settings)
}
