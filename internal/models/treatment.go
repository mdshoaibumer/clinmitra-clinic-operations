package models

type Treatment struct {
	ID           string `gorm:"type:text;primaryKey" json:"id"`
	Name         string `gorm:"type:text;not null" json:"name"`
	Code         string `gorm:"type:text;uniqueIndex" json:"code"`
	DefaultPrice int64  `gorm:"type:integer;not null" json:"defaultPrice"` // in paise
	Category     string `gorm:"type:text" json:"category"`
	Description  string `gorm:"type:text" json:"description"`
	IsActive     bool   `gorm:"default:true" json:"isActive"`
	CreatedAt    int64  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    int64  `gorm:"autoUpdateTime" json:"updatedAt"`
}

type PatientTreatment struct {
	BaseModel
	PatientID     string `gorm:"type:text;not null;index" json:"patientId"`
	TreatmentID   string `gorm:"type:text;not null" json:"treatmentId"`
	InvoiceID     string `gorm:"type:text" json:"invoiceId"`
	TreatmentDate string `gorm:"type:text;not null;index" json:"treatmentDate"`
	ToothNumber   string `gorm:"type:text" json:"toothNumber"`
	Notes         string `gorm:"type:text" json:"notes"`
	PerformedBy   string `gorm:"type:text" json:"performedBy"`

	// Relationships
	Patient   Patient   `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	Treatment Treatment `gorm:"foreignKey:TreatmentID" json:"treatment,omitempty"`
}
