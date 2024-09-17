package controllers

import (
	"golang.org/x/crypto/bcrypt"
)

// HashAPassword generates a bcrypt hash of a plain-text password.
// The hash is generated using a cost factor of 14, providing strong security.
//
// Parameters:
// - password: The plain-text password to hash.
//
// Returns:
// - string: The bcrypt hash of the password.
// - error: An error if hashing fails.
func HashAPassword(password string) (string, error) {
	// Generate a hashed version of the password using bcrypt
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// VerifyPasswordNonDuplicate compares a stored hashed password with a plain-text password.
// It ensures that the old and new passwords are not the same by checking if the new password matches the stored hash.
//
// Parameters:
// - oldPassword: The bcrypt hashed version of the old password.
// - newPassword: The plain-text new password to verify against the old hash.
//
// Returns:
// - bool: True if the passwords match, false otherwise.
// - error: An error if the comparison fails.
func VerifyPasswordNonDuplicate(oldPassword, newPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(newPassword))
	if err != nil {
		return false, err
	}

	return true, nil
}
