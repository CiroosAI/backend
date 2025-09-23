package admins

import (
	"net/http"
	"strconv"

	"project/database"
	"project/models"
	"project/utils"
)

type BankAccountResponse struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	UserName      string `json:"username"`
	Phone         string `json:"phone"`
	BankID        uint   `json:"bank_id"`
	BankName      string `json:"bank_name"`
	AccountName   string `json:"account_name"`
	AccountNumber string `json:"account_number"`
}

func GetBankAccounts(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	userId := r.URL.Query().Get("userId")
	bankId := r.URL.Query().Get("bankId")
	search := r.URL.Query().Get("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Start query
	db := database.DB
	query := db.Model(&models.BankAccount{}).
		Joins("JOIN users ON bank_accounts.user_id = users.id").
		Joins("JOIN banks ON bank_accounts.bank_id = banks.id")

	// Apply filters
	if userId != "" {
		query = query.Where("bank_accounts.user_id = ?", userId)
	}
	if bankId != "" {
		query = query.Where("bank_accounts.bank_id = ?", bankId)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("users.name LIKE ? OR users.number LIKE ? or bank_accounts.account_name LIKE ? or bank_accounts.account_number LIKE ?", like, like, like, like)
	}

	// Get bank accounts with joined details
	type BankAccountWithDetails struct {
		models.BankAccount
		UserName string
		Phone    string
		BankName string
	}

	var bankAccounts []BankAccountWithDetails
	query.Select("bank_accounts.*, users.name as user_name, users.number as phone, banks.name as bank_name").
		Offset(offset).
		Limit(limit).
		Find(&bankAccounts)

	// Transform to response format
	var response []BankAccountResponse
	for _, ba := range bankAccounts {
		response = append(response, BankAccountResponse{
			ID:            ba.ID,
			UserID:        ba.UserID,
			UserName:      ba.UserName,
			Phone:         ba.Phone,
			BankID:        ba.BankID,
			BankName:      ba.BankName,
			AccountName:   ba.AccountName,
			AccountNumber: ba.AccountNumber,
		})
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}
