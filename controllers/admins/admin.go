package admins

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"project/database"
	"project/models"
	"project/utils"
)

// GET /admin/profile
func GetAdminProfile(w http.ResponseWriter, r *http.Request) {
	// Extract Bearer token and validate to get admin ID
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized: No token provided",
		})
		return
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	_, claims, err := utils.ValidateAccessToken(tokenString)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized: Invalid token",
		})
		return
	}

	// Get admin ID from claims
	var adminID int64
	if rawID, ok := claims["id"]; ok {
		switch v := rawID.(type) {
		case float64:
			adminID = int64(v)
		case int64:
			adminID = v
		case int:
			adminID = int64(v)
		case string:
			var n int64
			_, _ = fmt.Sscanf(v, "%d", &n)
			adminID = n
		}
	}
	if adminID == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized: Invalid subject",
		})
		return
	}

	var admin models.Admin
	if err := database.DB.First(&admin, adminID).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
			Success: false,
			Message: "Admin tidak ditemukan",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data:    admin,
	})
}

type updateAdminProfileRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

// PUT /admin/profile
func UpdateAdminProfile(w http.ResponseWriter, r *http.Request) {
	// Extract admin ID from Bearer token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized: No token provided",
		})
		return
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	_, claims, err := utils.ValidateAccessToken(tokenString)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized: Invalid token",
		})
		return
	}
	var adminID int64
	if rawID, ok := claims["id"]; ok {
		switch v := rawID.(type) {
		case float64:
			adminID = int64(v)
		case int64:
			adminID = v
		case int:
			adminID = int64(v)
		case string:
			var n int64
			_, _ = fmt.Sscanf(v, "%d", &n)
			adminID = n
		}
	}
	if adminID == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized: Invalid subject",
		})
		return
	}

	var req updateAdminProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	var admin models.Admin
	if err := database.DB.First(&admin, adminID).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
			Success: false,
			Message: "Admin tidak ditemukan",
		})
		return
	}

	updates := map[string]interface{}{}
	if strings.TrimSpace(req.Username) != "" {
		updates["username"] = strings.TrimSpace(req.Username)
	}
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Email) != "" {
		updates["email"] = strings.TrimSpace(req.Email)
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&admin).Updates(updates).Error; err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
				Success: false,
				Message: "Gagal memperbarui profil",
			})
			return
		}
		// reload
		database.DB.First(&admin, adminID)
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Profil berhasil diperbarui",
		Data:    admin,
	})
}

type updateAdminPasswordRequest struct {
	CurrentPassword      string `json:"current_password"`
	NewPassword          string `json:"new_password"`
	ConfirmationPassword string `json:"confirmation_password"`
}

// PUT /admin/password
func UpdateAdminPassword(w http.ResponseWriter, r *http.Request) {
	// Extract admin ID from Bearer token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized: No token provided",
		})
		return
	}
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	_, claims, err := utils.ValidateAccessToken(tokenString)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized: Invalid token",
		})
		return
	}
	var adminID int64
	if rawID, ok := claims["id"]; ok {
		switch v := rawID.(type) {
		case float64:
			adminID = int64(v)
		case int64:
			adminID = v
		case int:
			adminID = int64(v)
		case string:
			var n int64
			_, _ = fmt.Sscanf(v, "%d", &n)
			adminID = n
		}
	}
	if adminID == 0 {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Unauthorized: Invalid subject",
		})
		return
	}

	var req updateAdminPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	// Basic validation
	if strings.TrimSpace(req.CurrentPassword) == "" || strings.TrimSpace(req.NewPassword) == "" || strings.TrimSpace(req.ConfirmationPassword) == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Semua field password wajib diisi",
		})
		return
	}
	if req.NewPassword != req.ConfirmationPassword {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{
			Success: false,
			Message: "Konfirmasi password tidak cocok",
		})
		return
	}

	var admin models.Admin
	if err := database.DB.First(&admin, adminID).Error; err != nil {
		utils.WriteJSON(w, http.StatusNotFound, utils.APIResponse{
			Success: false,
			Message: "Admin tidak ditemukan",
		})
		return
	}

	// Verify current password
	if !admin.ValidatePassword(req.CurrentPassword) {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{
			Success: false,
			Message: "Password saat ini salah",
		})
		return
	}

	// Set new password and hash
	admin.Password = req.NewPassword
	if err := admin.HashPassword(); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal mengatur ulang password",
		})
		return
	}
	if err := database.DB.Model(&admin).Update("password", admin.Password).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{
			Success: false,
			Message: "Gagal menyimpan password baru",
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Password berhasil diperbarui",
	})
}
