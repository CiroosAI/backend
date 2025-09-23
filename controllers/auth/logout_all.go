package auth

import (
	"net/http"
	"strings"
	"time"

	"project/database"
	"project/utils"

	"github.com/golang-jwt/jwt/v5"
)

// LogoutAllHandler revokes all refresh tokens for the authenticated user
func LogoutAllHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := utils.GetUserID(r)
	if !ok {
		utils.WriteJSON(w, http.StatusUnauthorized, utils.APIResponse{Success: false, Message: "Unauthorized"})
		return
	}

	// Best-effort: revoke current access token jti if present
	authz := r.Header.Get("Authorization")
	if authz != "" && strings.HasPrefix(authz, "Bearer ") {
		tokenStr := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
		if tkn, err := utils.ValidateToken(tokenStr); err == nil && tkn != nil {
			if claims, ok := tkn.Claims.(jwt.MapClaims); ok {
				if jtiRaw, ok := claims["jti"].(string); ok && jtiRaw != "" {
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
	}

	if database.DB == nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}
	if err := database.DB.Model(&map[string]interface{}{}).Where("user_id = ?", uid).Table("refresh_tokens").Update("revoked", true).Error; err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.APIResponse{Success: false, Message: "Server error"})
		return
	}
	utils.WriteJSON(w, http.StatusOK, utils.APIResponse{Success: true, Message: "All sessions revoked"})
}
