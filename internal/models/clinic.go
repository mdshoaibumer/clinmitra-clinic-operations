package models

type ClinicSettings struct {
	ID            string `gorm:"type:text;primaryKey" json:"id"`
	ClinicName    string `gorm:"type:text;not null" json:"clinicName"`
	DoctorName    string `gorm:"type:text;not null" json:"doctorName"`
	Address       string `gorm:"type:text" json:"address"`
	City          string `gorm:"type:text" json:"city"`
	State         string `gorm:"type:text" json:"state"`
	Pincode       string `gorm:"type:text" json:"pincode"`
	Phone         string `gorm:"type:text" json:"phone"`
	Email         string `gorm:"type:text" json:"email"`
	GSTIN         string `gorm:"type:text" json:"gstin"`
	GSTEnabled    bool   `gorm:"default:false" json:"gstEnabled"`
	GSTRate       int    `gorm:"default:18" json:"gstRate"` // percentage
	InvoicePrefix string `gorm:"type:text;default:'PV'" json:"invoicePrefix"`
	LogoPath      string `gorm:"type:text" json:"logoPath"`
	SetupComplete bool   `gorm:"default:false" json:"setupComplete"`
	AutoBackup    bool   `gorm:"default:true" json:"autoBackup"`
	BackupPath    string `gorm:"type:text" json:"backupPath"`
	CreatedAt     int64  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     int64  `gorm:"autoUpdateTime" json:"updatedAt"`
}
