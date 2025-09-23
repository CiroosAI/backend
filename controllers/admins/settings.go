package admins

import (
	"encoding/json"
	"net/http"
	"project/database"
	"project/models"
	"project/utils"
)

type SettingRequest struct {
	Name           string  `json:"name"`
	Logo           string  `json:"logo"`
	MinWithdraw    float64 `json:"min_withdraw"`
	MaxWithdraw    float64 `json:"max_withdraw"`
	WithdrawCharge float64 `json:"withdraw_charge"`
	Maintenance    bool    `json:"maintenance"`
	ClosedRegister bool    `json:"closed_register"`
	LinkCS         string  `json:"link_cs"`
	LinkGroup      string  `json:"link_group"`
	LinkApp        string  `json:"link_app"`
}

// GET /api/admin/settings
func GetSettingsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.DB

	var setting models.Setting
	if err := db.First(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Transform to response format
	response := map[string]interface{}{
		"name":            setting.Name,
		"logo":            setting.Logo,
		"min_withdraw":    setting.MinWithdraw,
		"max_withdraw":    setting.MaxWithdraw,
		"withdraw_charge": setting.WithdrawCharge,
		"maintenance":     setting.Maintenance,
		"closed_register": setting.ClosedRegister,
		"link_cs":         setting.LinkCS,
		"link_group":      setting.LinkGroup,
		"link_app":        setting.LinkApp,
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    response,
	})
}

// PUT /api/admin/settings
func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var req SettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	db := database.DB

	// Get current settings
	var setting models.Setting
	if err := db.First(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Update settings
	setting.Name = req.Name
	setting.Logo = req.Logo
	setting.MinWithdraw = req.MinWithdraw
	setting.MaxWithdraw = req.MaxWithdraw
	setting.WithdrawCharge = req.WithdrawCharge
	setting.Maintenance = req.Maintenance
	setting.ClosedRegister = req.ClosedRegister
	setting.LinkCS = req.LinkCS
	setting.LinkGroup = req.LinkGroup
	setting.LinkApp = req.LinkApp

	if err := db.Save(&setting).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Terjadi kesalahan sistem, silakan coba lagi",
		})
		return
	}

	// Transform to response format
	response := map[string]interface{}{
		"name":            setting.Name,
		"logo":            setting.Logo,
		"min_withdraw":    setting.MinWithdraw,
		"max_withdraw":    setting.MaxWithdraw,
		"withdraw_charge": setting.WithdrawCharge,
		"maintenance":     setting.Maintenance,
		"closed_register": setting.ClosedRegister,
		"link_cs":         setting.LinkCS,
		"link_group":      setting.LinkGroup,
		"link_app":        setting.LinkApp,
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Pengaturan berhasil diperbarui",
		Data:    response,
	})
}
