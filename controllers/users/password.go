package users

import (
	"encoding/json"
	"net/http"

	"project/database"
	"project/models"
	"project/utils"

	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordRequest struct {
	CurrentPassword      string `json:"current_password"`
	Password             string `json:"password"`
	ConfirmationPassword string `json:"confirmation_password"`
}

func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok || uid == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid request"})
		return
	}
	if len(req.Password) < 6 {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Kata sandi harus minimal 6 karakter"})
		return
	}
	if req.Password != req.ConfirmationPassword {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Konfirmasi kata sandi tidak cocok"})
		return
	}
	db := database.DB
	var user models.User
	if err := db.First(&user, uid).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{Success: false, Message: "User not found"})
		return
	}
	// Validate current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Kata sandi saat ini tidak cocok"})
		return
	}
	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Failed to hash password"})
		return
	}
	if err := db.Model(&user).Update("password", string(hash)).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Failed to update password"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Kata sandi berhasil diubah"})
}
