package models

import (
	"crypto/rand"
	"fmt"
	"time"
)

type RefreshToken struct {
	ID        string    `gorm:"primaryKey;type:char(36)" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"created_at"`
}

func NewRefreshToken(userID uint, ttlDays int) (*RefreshToken, error) {
	id, err := generateRandomID(32)
	if err != nil {
		return nil, err
	}
	return &RefreshToken{
		ID:        id,
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Duration(ttlDays) * 24 * time.Hour),
		Revoked:   false,
		CreatedAt: time.Now(),
	}, nil
}

func generateRandomID(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	const hex = "0123456789abcdef"
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = hex[int(b[i])%len(hex)]
	}
	return fmt.Sprintf("rt_%s", string(out)), nil
}
