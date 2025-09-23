package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"project/database"
	"project/utils"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshHandler exchanges a valid refresh token for a new access token and rotated refresh token
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON body"})
		return
	}
	if req.RefreshToken == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "refresh_token is required"})
		return
	}

	// validate existing refresh token
	rt, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Invalid refresh token"})
		return
	}

	// rotate: revoke old token and create new one in a transaction
	tx := database.DB.Begin()
	rt.Revoked = true
	if err := tx.Save(rt).Error; err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}
	newJTI, _, err := utils.GenerateRefreshToken(rt.UserID)
	if err != nil {
		tx.Rollback()
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}
	if err := tx.Commit().Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}

	// issue new access token
	accessToken, err := utils.GenerateAccessToken(rt.UserID, "user")
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{
		Success: true,
		Message: "Successfully",
		Data: map[string]interface{}{
			"access_token":  accessToken,
			"access_expire": time.Now().Add(15 * time.Minute).UTC().Format(time.RFC3339),
			"refresh_token": newJTI,
		},
	})
}
