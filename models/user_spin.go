package models

import "time"

type UserSpin struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	UserID  uint      `gorm:"not null;index" json:"user_id"`
	PrizeID uint      `gorm:"not null;index" json:"prize_id"`
	Amount  float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Code    string    `gorm:"type:varchar(20);not null" json:"code"`
	WonAt   time.Time `json:"won_at"`
}
