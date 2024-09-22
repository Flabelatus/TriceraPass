package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID        string    `gorm:"type:uuid;primary_key"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"deleted_at"`
	UserName  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Mode      Mode      `gorm:"foreignKey:UserID" json:"mode,omitempty"`
}

type Mode struct {
	gorm.Model
	Name   string `json:"mode_name"`
	UserID string `json:"user_id"`
}

type UserConfirmation struct {
	ID        string    `gorm:"type:uuid;primary_key" json:"id"`
	UserID    string    `json:"user_id"`
	ExpiredAt int64     `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
	Confirmed bool      `json:"confirmed"`
}

type PasswordRestToken struct {
	ID        string    `gorm:"type:uuid;primary_key" json:"id"`
	UserID    string    `json:"user_id"`
	ExpiredAt int64     `json:"expired_at"`
	CreatedAt time.Time `json:"created_at"`
	TokenUsed bool      `json:"token_used"`
}

func (pwToken *PasswordRestToken) IsTokenExpired() bool {
	// Convert the Unix timestamp to a time.Time object
	expirationTime := time.Unix(pwToken.ExpiredAt, 0).UTC()

	// Get the current time in UTC
	currentTime := time.Now().UTC()

	// Check if the current time is after the expiration time
	return currentTime.After(expirationTime)
}

func (pwToken *PasswordRestToken) SetTokenExpire() {
	if !pwToken.IsTokenExpired() {
		pwToken.ExpiredAt = time.Now().UTC().UnixNano()
	}
}

// Define a function to check if a given time has expired
func (uc *UserConfirmation) IsExpired() bool {
	// Convert the Unix timestamp to a time.Time object
	expirationTime := time.Unix(uc.ExpiredAt, 0).UTC()

	// Get the current time in UTC
	currentTime := time.Now().UTC()

	// Check if the current time is after the expiration time
	return currentTime.After(expirationTime)
}

func (uc *UserConfirmation) SetExpire() {
	if !uc.IsExpired() {
		uc.ExpiredAt = time.Now().UTC().UnixNano()
	}
}

func (u *User) PasswordMatches(plainText string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainText))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			// invalid password
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
