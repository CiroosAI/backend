package models

type BankAccount struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	UserID        uint   `gorm:"not null;index" json:"user_id"`
	BankID        uint   `gorm:"not null;index" json:"bank_id"`
	AccountName   string `gorm:"size:100;not null" json:"account_name"`
	AccountNumber string `gorm:"size:50;not null" json:"account_number"`
	Bank          *Bank  `gorm:"foreignKey:BankID" json:"bank,omitempty"`
}

func (BankAccount) TableName() string {
	return "bank_accounts"
}
