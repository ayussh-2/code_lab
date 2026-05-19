package models

import "time"

type User struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	Email                 string     `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password              string     `gorm:"not null;size:255" json:"-"`
	Name                  string     `gorm:"size:100" json:"first_name"`
	Username              string     `gorm:"uniqueIndex;size:32" json:"username"`
	AvatarURL             string     `gorm:"size:512" json:"avatar_url"`
	Bio                   string     `gorm:"type:text" json:"bio"`
	Rating                int        `gorm:"not null;default:1500" json:"rating"`
	Role                  string     `gorm:"not null;size:50;default:user" json:"role"`
	EmailVerified         bool       `gorm:"not null;default:true" json:"email_verified"`
	RefreshTokenHash      string     `gorm:"size:255" json:"-"`
	RefreshTokenExpiresAt *time.Time `json:"-"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
