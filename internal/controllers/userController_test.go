package controllers

import (
	"testing"
)

// Test HashAPassword function
func TestHashAPassword(t *testing.T) {
	password := "mysecurepassword"
	hashedPassword, err := HashAPassword(password)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// bcrypt hashes are always 60 characters long
	if len(hashedPassword) != 60 {
		t.Errorf("expected hashed password to be 60 characters long, got %d", len(hashedPassword))
	}

	// Verify that the hashed password is not the same as the original password
	if hashedPassword == password {
		t.Errorf("expected hashed password to be different from the original password")
	}
}

// Test VerifyPasswordNonDuplicate function
func TestVerifyPasswordNonDuplicate(t *testing.T) {
	password := "mysecurepassword"
	hashedPassword, err := HashAPassword(password)
	if err != nil {
		t.Fatalf("Error hashing password: %v", err)
	}

	// Test case: Correct password match
	match, err := VerifyPasswordNonDuplicate(hashedPassword, password)
	if err != nil || !match {
		t.Errorf("expected passwords to match, got match: %v, error: %v", match, err)
	}

	// Test case: Mismatched password
	wrongPassword := "anotherpassword"
	match, err = VerifyPasswordNonDuplicate(hashedPassword, wrongPassword)
	if err == nil || match {
		t.Errorf("expected passwords to not match, got match: %v, error: %v", match, err)
	}
}
