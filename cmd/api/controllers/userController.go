package controllers

import (
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"

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

func SelectRandomDefaultProfileImage(defaultDirPath string) (string, error) {

	pathToImages := fmt.Sprintf("%s/static/default", defaultDirPath)
	files, err := os.ReadDir(pathToImages)
	if err != nil {
		return "", err
	}

	RandomIndex := rand.Intn(len(files))
	selectedImage := files[RandomIndex]

	ImagePath := fmt.Sprintf("%s/%s", pathToImages, selectedImage.Name())

	// Return the picked image for the profile picture
	return ImagePath, nil
}

func UploadDefaultProfile(rootDir, userID string) (string, string, error) {

	// Default path for profile pictures
	directoryPath := fmt.Sprintf("%s/static/profile", rootDir)

	// Create a default profile image
	imagePath, err := SelectRandomDefaultProfileImage(rootDir)
	if err != nil {
		return "", "", err
	}

	filename := userID + filepath.Ext(imagePath)

	// Ensure the directory exists
	err = os.MkdirAll(directoryPath, fs.ModePerm)
	if err != nil {
		return "", "", err
	}

	profilePath := fmt.Sprintf("%s/%s", directoryPath, filename)

	newFile, err := os.Create(profilePath)
	if err != nil {
		return "", "", err
	}
	defer newFile.Close()

	file, err := os.Open(imagePath)
	if err != nil {
		return "", "", fmt.Errorf("error: file could not be opened: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		return "", "", fmt.Errorf("error: contents could not be copied: %v", err)
	}

	return filename, profilePath, nil
}

func DeleteProfileImageFile(rootDir, userID string) error {
	profileImagePath := fmt.Sprintf("%s/static/profile", rootDir)
	profileImageDir, err := os.ReadDir(profileImagePath)
	if err != nil {
		return err
	}

	ArbitraryPath := profileImageDir[0].Name()

	// Concatenate the user ID with the extension of a image file
	filename := userID + filepath.Ext(ArbitraryPath)

	for _, fp := range profileImageDir {
		if fp.Name() == filename {
			err = os.Remove(fmt.Sprintf("%s/%s", profileImagePath, fp.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
