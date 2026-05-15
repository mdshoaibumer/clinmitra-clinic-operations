package models

import "time"

type UserRole string

const (
	RoleAdmin        UserRole = "admin"
	RoleDoctor       UserRole = "doctor"
	RoleReceptionist UserRole = "receptionist"
)

type User struct {
	BaseModel
	Username     string     `gorm:"type:text;uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"type:text;not null" json:"-"`
	FullName     string     `gorm:"type:text;not null" json:"fullName"`
	Role         UserRole   `gorm:"type:text;not null;default:'admin'" json:"role"`
	IsActive     bool       `gorm:"default:true" json:"isActive"`
	LastLoginAt  *time.Time `gorm:"type:datetime" json:"lastLoginAt"`
}
