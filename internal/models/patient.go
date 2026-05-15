package models

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type Patient struct {
	BaseModel
	Name           string `gorm:"type:text;not null;index" json:"name"`
	Phone          string `gorm:"type:text;not null;index" json:"phone"`
	Email          string `gorm:"type:text" json:"email"`
	Gender         Gender `gorm:"type:text;not null" json:"gender"`
	Age            int    `gorm:"type:integer" json:"age"`
	DateOfBirth    string `gorm:"type:text" json:"dateOfBirth"`
	Address        string `gorm:"type:text" json:"address"`
	City           string `gorm:"type:text" json:"city"`
	BloodGroup     string `gorm:"type:text" json:"bloodGroup"`
	MedicalHistory string `gorm:"type:text" json:"medicalHistory"`
	Allergies      string `gorm:"type:text" json:"allergies"`
	Notes          string `gorm:"type:text" json:"notes"`
	CreatedBy      string `gorm:"type:text" json:"createdBy"`

	// Relationships
	Appointments      []Appointment      `gorm:"foreignKey:PatientID" json:"appointments,omitempty"`
	Invoices          []Invoice          `gorm:"foreignKey:PatientID" json:"invoices,omitempty"`
	PatientTreatments []PatientTreatment `gorm:"foreignKey:PatientID" json:"patientTreatments,omitempty"`
}
