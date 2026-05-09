package models

import "time"

const (
	OTPPurposePasswordReset = "password_reset"
	OTPPurposeEmailVerify   = "email_verify"
)

type OTP struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     *uint      `gorm:"index" json:"user_id,omitempty"`
	Email      string     `gorm:"index;not null;size:255" json:"email"`
	CodeHash   string     `gorm:"not null;size:255" json:"-"`
	Purpose    string     `gorm:"not null;size:50;index" json:"purpose"`
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expires_at"`
	ConsumedAt *time.Time `json:"consumed_at,omitempty"`
	Attempts   int        `gorm:"not null;default:0" json:"attempts"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
