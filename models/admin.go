package models

import (
	"time"

	"project/database"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"` // Password won't be included in JSON responses
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique"`
	Role      string    `json:"role" gorm:"default:admin"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate handles password hashing before saving to database
func (a *Admin) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.Password = string(hashedPassword)
	return nil
}

// ValidatePassword checks if the provided password matches the hashed password
func (a *Admin) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	return err == nil
}

// GetAdminByUsername retrieves an admin by username
func GetAdminByUsername(username string) (*Admin, error) {
	var admin Admin
	result := database.DB.Where("username = ? AND is_active = ?", username, true).First(&admin)
	if result.Error != nil {
		return nil, result.Error
	}
	return &admin, nil
}
