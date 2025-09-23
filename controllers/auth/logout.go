package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"project/database"
	"project/utils"

	"github.com/golang-jwt/jwt/v5"
)

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutHandler revokes a specific refresh token and (optionally) the access token jti from Authorization header
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON body"})
		return
	}
	if req.RefreshToken == "" {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "refresh_token is required"})
		return
	}

	// Attempt to revoke access-token jti if Authorization header is present
	authz := r.Header.Get("Authorization")
	if authz != "" && strings.HasPrefix(authz, "Bearer ") {
		tokenStr := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
		if tkn, err := utils.ValidateToken(tokenStr); err == nil && tkn != nil {
			if claims, ok := tkn.Claims.(jwt.MapClaims); ok {
				if jtiRaw, ok := claims["jti"].(string); ok && jtiRaw != "" {
					// determine TTL from exp if possible
					var ttl time.Duration = 0
					if expRaw, ok := claims["exp"]; ok {
						switch v := expRaw.(type) {
						case float64:
							expTime := time.Unix(int64(v), 0)
							ttl = time.Until(expTime)
						case int64:
							expTime := time.Unix(v, 0)
							ttl = time.Until(expTime)
						case int:
							expTime := time.Unix(int64(v), 0)
							ttl = time.Until(expTime)
						}
					}
					if ttl < 0 {
						ttl = 0
					}
					_ = utils.RevokeJTI(jtiRaw, ttl)
				}
			}
		}
		// ignore errors parsing access token; still proceed to revoke refresh token
	}

	if database.DB == nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}
	if err := database.DB.Model(&modelRefreshToken{}).Where("id = ?", req.RefreshToken).Update("revoked", true).Error; err != nil {
		// If row not found return success to avoid token enumeration
		utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Logged out"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "Logged out"})
}

// modelRefreshToken is a light local struct to avoid import cycles
type modelRefreshToken struct {
	ID string `json:"id"`
}
