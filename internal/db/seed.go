package db

import (
	"practivo/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var defaultTreatments = []struct {
	Name         string
	Code         string
	DefaultPrice int64 // in paise
	Category     string
}{
	{"Consultation", "CONSULT", 50000, "General"},
	{"Dental Cleaning (Scaling)", "SCALING", 150000, "Preventive"},
	{"Tooth Extraction (Simple)", "EXT-SIMPLE", 100000, "Surgery"},
	{"Tooth Extraction (Surgical)", "EXT-SURG", 300000, "Surgery"},
	{"Root Canal Treatment (Anterior)", "RCT-ANT", 500000, "Endodontics"},
	{"Root Canal Treatment (Posterior)", "RCT-POST", 800000, "Endodontics"},
	{"Dental Filling (Composite)", "FILL-COMP", 150000, "Restorative"},
	{"Dental Filling (GIC)", "FILL-GIC", 100000, "Restorative"},
	{"Crown (Metal)", "CROWN-MET", 400000, "Prosthodontics"},
	{"Crown (Ceramic)", "CROWN-CER", 800000, "Prosthodontics"},
	{"Crown (Zirconia)", "CROWN-ZIR", 1200000, "Prosthodontics"},
	{"Bridge (per unit)", "BRIDGE", 500000, "Prosthodontics"},
	{"Denture (Complete)", "DENT-COMP", 1500000, "Prosthodontics"},
	{"Denture (Partial)", "DENT-PART", 800000, "Prosthodontics"},
	{"Teeth Whitening", "WHITEN", 500000, "Cosmetic"},
	{"Dental Implant", "IMPLANT", 3500000, "Implantology"},
	{"Orthodontic Consultation", "ORTHO-CON", 100000, "Orthodontics"},
	{"Braces (Metal)", "BRACES-MET", 3000000, "Orthodontics"},
	{"Braces (Ceramic)", "BRACES-CER", 4500000, "Orthodontics"},
	{"Wisdom Tooth Removal", "WISDOM", 500000, "Surgery"},
	{"Fluoride Treatment", "FLUORIDE", 80000, "Preventive"},
	{"Pit & Fissure Sealant", "SEALANT", 100000, "Preventive"},
	{"Gum Treatment (Curettage)", "GUM-CUR", 200000, "Periodontics"},
	{"Gum Surgery (Flap)", "GUM-FLAP", 500000, "Periodontics"},
	{"X-Ray (Single)", "XRAY-S", 30000, "Diagnostic"},
	{"X-Ray (OPG)", "XRAY-OPG", 50000, "Diagnostic"},
	{"Temporary Filling", "FILL-TEMP", 50000, "Restorative"},
	{"Re-cementation", "RECEMENT", 50000, "Restorative"},
	{"Post & Core", "POST-CORE", 300000, "Endodontics"},
	{"Mouth Guard", "MGUARD", 200000, "Preventive"},
}

// SeedTreatments populates the treatments table with default dental
// procedures if the table is empty. Inserts in batches of 10.
// This is idempotent — subsequent calls after the initial seed are no-ops.
func SeedTreatments(db *gorm.DB) error {
	var count int64
	db.Model(&models.Treatment{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	treatments := make([]models.Treatment, len(defaultTreatments))
	for i, t := range defaultTreatments {
		treatments[i] = models.Treatment{
			ID:           uuid.New().String(),
			Name:         t.Name,
			Code:         t.Code,
			DefaultPrice: t.DefaultPrice,
			Category:     t.Category,
			IsActive:     true,
		}
	}

	return db.CreateInBatches(treatments, 10).Error
}
