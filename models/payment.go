package models

import "time"

type Payment struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	InvestmentID   uint       `gorm:"not null;index" json:"investment_id"`
	ReferenceID    *string    `gorm:"type:varchar(191)" json:"reference_id,omitempty"`
	OrderID        string     `gorm:"type:varchar(191);not null;uniqueIndex" json:"order_id"`
	PaymentMethod  *string    `gorm:"type:varchar(16)" json:"payment_method,omitempty"`
	PaymentChannel *string    `gorm:"type:varchar(16)" json:"payment_channel,omitempty"`
	PaymentCode    *string    `gorm:"type:text" json:"payment_code,omitempty"`
	PaymentLink    *string    `gorm:"type:text" json:"payment_link,omitempty"`
	Status         string     `gorm:"type:varchar(16);default:'Pending'" json:"status"`
	ExpiredAt      *time.Time `json:"expired_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}
