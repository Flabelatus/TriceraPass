package controllers

import (
	"golang.org/x/crypto/bcrypt"
)

func HashAPassword(password string) (string, error) {
	// Generate a hashed version of the password using bcrypt
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func VerifyPasswordNonDuplicate(oldPassword, newPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(newPassword))
	if err != nil {
		return false, err
	}

	return true, nil
}
