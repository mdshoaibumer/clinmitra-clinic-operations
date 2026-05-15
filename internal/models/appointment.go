package models

type AppointmentStatus string

const (
	AppointmentScheduled AppointmentStatus = "scheduled"
	AppointmentCompleted AppointmentStatus = "completed"
	AppointmentCancelled AppointmentStatus = "cancelled"
	AppointmentNoShow    AppointmentStatus = "no_show"
)

type Appointment struct {
	BaseModel
	PatientID       string            `gorm:"type:text;not null;index" json:"patientId"`
	AppointmentDate string            `gorm:"type:text;not null;index:idx_appt_date_status;index" json:"appointmentDate"` // YYYY-MM-DD
	StartTime       string            `gorm:"type:text;not null;index:idx_appt_time_range" json:"startTime"`              // HH:MM
	EndTime         string            `gorm:"type:text;not null;index:idx_appt_time_range" json:"endTime"`                // HH:MM
	Duration        int               `gorm:"type:integer;default:30" json:"duration"`                                    // minutes
	Status          AppointmentStatus `gorm:"type:text;not null;default:'scheduled';index:idx_appt_date_status" json:"status"`
	Purpose         string            `gorm:"type:text" json:"purpose"`
	Notes           string            `gorm:"type:text" json:"notes"`
	CancelReason    string            `gorm:"type:text" json:"cancelReason"`
	CreatedBy       string            `gorm:"type:text" json:"createdBy"`

	// Relationships
	Patient Patient `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
}
