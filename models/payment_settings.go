package models

import (
	"strconv"
	"strings"
)

type PaymentSettings struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	PakasirAPIKey  string  `gorm:"size:191" json:"PAKASIR_API_KEY"`
	PakasirProject string  `gorm:"size:191" json:"PAKASIR_PROJECT"`
	DepositAmount  float64 `gorm:"type:decimal(15,2)" json:"DEPOSIT_AMOUNT"`
	BankName       string  `gorm:"size:100" json:"BANK_NAME"`
	BankCode       string  `gorm:"size:50" json:"BANK_CODE"`
	AccountNumber  string  `gorm:"size:100" json:"ACCOUNT_NUMBER"`
	AccountName    string  `gorm:"size:100" json:"ACCOUNT_NAME"`
	WithdrawAmount float64 `gorm:"type:decimal(15,2)" json:"WITHDRAW_AMOUNT"`
	WishlistID     string  `gorm:"type:text" json:"WISHLIST_ID"` // CSV of user IDs, e.g. "2,3,4,5,6"
}

func (PaymentSettings) TableName() string { return "payment_settings" }

// IsUserInWishlist checks if the given userID is present in the CSV list.
func (ps *PaymentSettings) IsUserInWishlist(userID uint) bool {
	if ps == nil || strings.TrimSpace(ps.WishlistID) == "" {
		return false
	}
	parts := strings.Split(ps.WishlistID, ",")
	needle := fmtUint(userID)
	for _, p := range parts {
		if strings.TrimSpace(p) == needle {
			return true
		}
	}
	return false
}

func fmtUint(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}
