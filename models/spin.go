package models

import "time"

type SpinPrize struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Amount       float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Code         string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"`
	Chance       int       `gorm:"not null" json:"chance"`
	ChanceWeight int       `gorm:"not null" json:"chance_weight"`
	Status       string    `gorm:"type:enum('Active','Inactive');not null;default:'Active'" json:"status"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (SpinPrize) TableName() string {
	return "spin_prizes"
}
