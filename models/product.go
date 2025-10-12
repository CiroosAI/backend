package models

import "time"

type Product struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CategoryID    uint      `gorm:"column:category_id;not null;index" json:"category_id"`
	Name          string    `gorm:"column:name;size:100;not null" json:"name"`
	Amount        float64   `gorm:"column:amount;type:decimal(15,2);not null" json:"amount"`
	DailyProfit   float64   `gorm:"column:daily_profit;type:decimal(15,2);not null" json:"daily_profit"`
	Duration      int       `gorm:"column:duration;not null" json:"duration"`
	RequiredVIP   int       `gorm:"column:required_vip;default:0" json:"required_vip"`
	PurchaseLimit int       `gorm:"column:purchase_limit;default:0" json:"purchase_limit"` // 0 = unlimited
	Status        string    `gorm:"column:status;type:enum('Active','Inactive');default:'Active'" json:"status"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
	
	// Relations
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (Product) TableName() string {
	return "products"
}
