package models

type PaymentMethod string

const (
	PaymentCash     PaymentMethod = "cash"
	PaymentUPI      PaymentMethod = "upi"
	PaymentCard     PaymentMethod = "card"
	PaymentTransfer PaymentMethod = "bank_transfer"
	PaymentOther    PaymentMethod = "other"
)

type Payment struct {
	BaseModel
	InvoiceID   string        `gorm:"type:text;not null;index" json:"invoiceId"`
	Amount      int64         `gorm:"type:integer;not null" json:"amount"` // paise
	Method      PaymentMethod `gorm:"type:text;not null" json:"method"`
	PaymentDate string        `gorm:"type:text;not null;index" json:"paymentDate"` // YYYY-MM-DD
	Reference   string        `gorm:"type:text" json:"reference"`                  // UPI ref, card last 4, etc.
	Notes       string        `gorm:"type:text" json:"notes"`
	ReceivedBy  string        `gorm:"type:text" json:"receivedBy"`

	// Relationships
	Invoice Invoice `gorm:"foreignKey:InvoiceID" json:"invoice,omitempty"`
}
