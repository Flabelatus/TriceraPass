package handlers

import (
	"TriceraPass/cmd/api/utils"
	"TriceraPass/cmd/api/application"
	"TriceraPass/internal/models"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
)

func UploadProfileImage(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := r.URL.Query().Get("uid")
		file, handler, err := r.FormFile("image")
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		defer file.Close()

		filename := userID + filepath.Ext(handler.Filename)
		directoryPath := "./static/profile"

		// Ensure the directory exists
		err = os.MkdirAll(directoryPath, fs.ModePerm)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

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

		image := models.ProfileImage{Filename: filename, FilePath: filepath, UserID: userID}
		imageID, err := app.Repository.InsertProfileImage(&image)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		response := utils.JSONResponse{Data: imageID, Message: "Profile image successfully uploaded"}
		_ = utils.WriteJSON(w, http.StatusCreated, response)
	}
}

func ServeStaticProfileImage(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileName := chi.URLParam(r, "filename")
		profileImage, err := app.Repository.GetProfileImageByFilename(fileName)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		ProfileImageFilename := profileImage.Filename
		staticDir := "./static/profile"
		filePath := filepath.Join(staticDir, ProfileImageFilename)
		http.ServeFile(w, r, filePath)
	}
}

func GetProfileImageByFilename(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filename := chi.URLParam(r, "filename")
		profileImage, err := app.Repository.GetProfileImageByFilename(filename)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		response := utils.JSONResponse{Data: profileImage}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

func GetProfileImageByUserID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		profileImage, err := app.Repository.GetProfileImageByUserID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		response := utils.JSONResponse{
			Data: profileImage,
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

func DeleteProfileImageByUserID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		profileImage, err := app.Repository.GetProfileImageByUserID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		err = app.Repository.DeleteProfileImageByFilename(profileImage.Filename)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		response := utils.JSONResponse{
			Message: "Profile image successfully deleted",
			Error:   false,
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}
