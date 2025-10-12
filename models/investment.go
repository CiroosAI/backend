package models

import "time"

type Investment struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"not null;index" json:"user_id"`
	ProductID     uint       `gorm:"not null;index" json:"product_id"`
	CategoryID    uint       `gorm:"not null;index" json:"category_id"`
	Amount        float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	DailyProfit   float64    `gorm:"type:decimal(15,2);not null" json:"daily_profit"`
	Duration      int        `gorm:"not null" json:"duration"`
	TotalPaid     int        `gorm:"not null;default:0" json:"total_paid"`
	TotalReturned float64    `gorm:"type:decimal(15,2);not null;default:0.00" json:"total_returned"`
	LastReturnAt  *time.Time `json:"last_return_at,omitempty"`
	NextReturnAt  *time.Time `json:"next_return_at,omitempty"`
	OrderID       string     `gorm:"type:varchar(191);not null;uniqueIndex" json:"order_id"`
	Status        string     `gorm:"type:enum('Pending','Running','Completed','Suspended','Cancelled');default:'Pending'" json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	
	// Relations
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (Investment) TableName() string {
	return "investments"
}
