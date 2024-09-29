package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/internal/models"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

// UploadProfileImage handles the upload of a user's profile image.
// It retrieves the image file from the request, saves it to a static directory, and stores the file information in the database.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for the image upload route.
func UploadProfileImage(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get the user ID from the query parameters
		userID := r.URL.Query().Get("uid")

		// Get the image file from the form data
		file, handler, err := r.FormFile("image")
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		defer file.Close()

		// Set the filename based on the user ID and the file extension
		filename := userID + filepath.Ext(handler.Filename)
		directoryPath := "./static/profile"

		// Ensure the directory exists
		err = os.MkdirAll(directoryPath, fs.ModePerm)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Create the file path and save the uploaded image
		filepath := fmt.Sprintf("%s/%s", directoryPath, filename)
		newFile, err := os.Create(filepath)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		defer newFile.Close()

		_, err = io.Copy(newFile, file)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Insert the profile image record into the database
		image := models.ProfileImage{Filename: filename, FilePath: filepath, UserID: userID}
		imageID, err := app.Repository.InsertProfileImage(&image)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Respond with success and the image ID
		response := utils.JSONResponse{Data: imageID, Message: "Profile image successfully uploaded"}
		_ = utils.WriteJSON(w, http.StatusCreated, response)
	}
}

// ServeStaticProfileImage serves a static profile image by filename from the server's file system.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for serving profile images.
func ServeStaticProfileImage(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get the filename from the URL parameters
		fileName := chi.URLParam(r, "filename")

		// Fetch the profile image information from the database
		profileImage, err := app.Repository.GetProfileImageByFilename(fileName)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Construct the file path and serve the file
		staticDir := fmt.Sprintf("%s/static/profile", app.Root)
		filePath := filepath.Join(staticDir, profileImage.Filename)
		http.ServeFile(w, r, filePath)
	}
}

// GetProfileImageByFilename retrieves profile image information by its filename from the database and returns it as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for retrieving profile image metadata by filename.
func GetProfileImageByFilename(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get the filename from the URL parameters
		filename := chi.URLParam(r, "filename")

		// Fetch the profile image information from the database
		profileImage, err := app.Repository.GetProfileImageByFilename(filename)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Return the profile image information as JSON
		response := utils.JSONResponse{Data: profileImage}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// GetProfileImageByUserID retrieves profile image information by user ID from the database and returns it as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for retrieving profile image metadata by user ID.
func GetProfileImageByUserID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get the user ID from the URL parameters
		userID := chi.URLParam(r, "user_id")

		// Fetch the profile image information from the database
		profileImage, err := app.Repository.GetProfileImageByUserID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Return the profile image information as JSON
		response := utils.JSONResponse{
			Data: profileImage,
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// DeleteProfileImageByUserID deletes a user's profile image from the server and database by their user ID.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for deleting profile images by user ID.
func DeleteProfileImageByUserID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Get the user ID from the URL parameters
		userID := chi.URLParam(r, "user_id")

		// Fetch the profile image information from the database
		profileImage, err := app.Repository.GetProfileImageByUserID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Delete the profile image file from the server and database
		err = app.Repository.DeleteProfileImageByFilename(profileImage.Filename)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Return a success response
		response := utils.JSONResponse{
			Message: "Profile image successfully deleted",
			Error:   false,
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}
