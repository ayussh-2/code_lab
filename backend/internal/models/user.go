package models

import "time"

type User struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	Email                 string     `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password              string     `gorm:"not null;size:255" json:"-"`
	Name                  string     `gorm:"size:100" json:"first_name"`
	RefreshTokenHash      string     `gorm:"size:255" json:"-"`
	RefreshTokenExpiresAt *time.Time `json:"-"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
