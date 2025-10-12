package models

import "time"

type Category struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	ProfitType  string    `gorm:"type:enum('locked','unlocked');default:'unlocked'" json:"profit_type"`
	Status      string    `gorm:"type:enum('Active','Inactive');default:'Active'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Category) TableName() string {
	return "categories"
}

