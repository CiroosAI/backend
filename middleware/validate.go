package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"project/utils"
)

// ValidateJSON decodes JSON payload into dst and runs validator.ValidateStruct.
// It also enforces a request timeout (via ctx) and expects Content-Type: application/json.
func ValidateJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if ct := r.Header.Get("Content-Type"); ct != "application/json" && ct != "application/json; charset=utf-8" {
		utils.WriteJSON(w, http.StatusUnsupportedMediaType, utils.APIResponse{Success: false, Message: "Content-Type must be application/json"})
		return http.ErrNotSupported
	}
	// apply a short timeout for parsing/validation
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	r = r.WithContext(ctx)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Invalid JSON body"})
		return err
	}
	if err := utils.ValidateStruct(dst); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.APIResponse{Success: false, Message: "Validation failed", Data: err.Error()})
		return err
	}
	return nil
}
