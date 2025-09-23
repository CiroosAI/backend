package models

import "time"

type Product struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Minimum    float64   `gorm:"type:decimal(15,2);not null" json:"minimum"`
	Maximum    float64   `gorm:"type:decimal(15,2);not null" json:"maximum"`
	Percentage float64   `gorm:"type:decimal(5,2);not null" json:"percentage"`
	Duration   int       `gorm:"not null" json:"duration"`
	Status     string    `gorm:"type:enum('Active','Inactive');default:'Active'" json:"status"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

func (Product) TableName() string {
	return "products"
}
