package models

import "time"

type RatingHistory struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	Rating     int       `gorm:"not null" json:"rating"`
	RecordedAt time.Time `gorm:"not null;index" json:"recorded_at"`
}
