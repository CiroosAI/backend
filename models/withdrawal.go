package models

import "time"

type Withdrawal struct {
	ID            uint         `gorm:"primaryKey" json:"id"`
	UserID        uint         `gorm:"not null;index" json:"user_id"`
	BankAccountID uint         `gorm:"not null;index" json:"bank_account_id"`
	Amount        float64      `gorm:"type:decimal(15,2);not null" json:"amount"`
	Charge        float64      `gorm:"type:decimal(15,2);not null;default:0.00" json:"charge"`
	FinalAmount   float64      `gorm:"type:decimal(15,2);not null" json:"final_amount"`
	OrderID       string       `gorm:"type:varchar(191);not null;uniqueIndex" json:"order_id"`
	Status        string       `gorm:"type:enum('Success','Pending','Failed');not null;default:'Pending'" json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	BankAccount   *BankAccount `gorm:"foreignKey:BankAccountID" json:"bank_account,omitempty"`
}

func (Withdrawal) TableName() string {
	return "withdrawals"
}
