package models

import "time"

type Transaction struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `gorm:"not null;index" json:"user_id"`
	Amount           float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Charge           float64   `gorm:"type:decimal(15,2);not null;default:0.00" json:"charge"`
	OrderID          string    `gorm:"type:varchar(191);not null;uniqueIndex" json:"order_id"`
	TransactionFlow  string    `gorm:"type:enum('debit','credit');not null" json:"transaction_flow"`
	TransactionType  string    `gorm:"type:varchar(50);not null" json:"transaction_type"`
	Message          *string   `gorm:"type:text" json:"message,omitempty"`
	Status           string    `gorm:"type:enum('Success','Pending','Failed');not null;default:'Pending'" json:"status"`
	CreatedAt        time.Time `json:"-"`
	UpdatedAt        time.Time `json:"-"`
}

func (Transaction) TableName() string {
	return "transactions"
}
