package models

import "time"

type AuditAction string

const (
	AuditCreate AuditAction = "CREATE"
	AuditUpdate AuditAction = "UPDATE"
	AuditDelete AuditAction = "DELETE"
	AuditLogin  AuditAction = "LOGIN"
	AuditLogout AuditAction = "LOGOUT"
	AuditBackup AuditAction = "BACKUP"
)

type AuditLog struct {
	ID         string      `gorm:"type:text;primaryKey" json:"id"`
	UserID     string      `gorm:"type:text;index" json:"userId"`
	Action     AuditAction `gorm:"type:text;not null" json:"action"`
	EntityType string      `gorm:"type:text;not null;index:idx_entity" json:"entityType"`
	EntityID   string      `gorm:"type:text;index:idx_entity" json:"entityId"`
	OldValue   string      `gorm:"type:text" json:"oldValue"` // JSON
	NewValue   string      `gorm:"type:text" json:"newValue"` // JSON
	IPAddress  string      `gorm:"type:text" json:"ipAddress"`
	CreatedAt  time.Time   `gorm:"autoCreateTime;index" json:"createdAt"`
}
