package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"TriceraPass/internal/models"

	"github.com/go-chi/chi/v5"
)

func (app *application) UploadProfileImage(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("uid")
	file, handler, err := r.FormFile("image")
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer file.Close()

	filename := userID + filepath.Ext(handler.Filename)
	directoryPath := "./static/profile"

	// Ensure the directory exists
	err = os.MkdirAll(directoryPath, fs.ModePerm)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	filepath := fmt.Sprintf("%s/%s", directoryPath, filename)

	newFile, err := os.Create(filepath)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, file)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	image := models.ProfileImage{Filename: filename, FilePath: filepath, UserID: userID}
	imageID, err := app.Repository.InsertProfileImage(&image)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{Data: imageID, Message: "Profile image successfully uploaded"}
	_ = app.writeJSON(w, http.StatusCreated, response)
}

func (app *application) ServeStaticProfileImage(w http.ResponseWriter, r *http.Request) {
	fileName := chi.URLParam(r, "filename")
	profileImage, err := app.Repository.GetProfileImageByFilename(fileName)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	ProfileImageFilename := profileImage.Filename
	staticDir := "./static/profile"
	filePath := filepath.Join(staticDir, ProfileImageFilename)
	http.ServeFile(w, r, filePath)
}

func (app *application) GetProfileImageByFilename(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	profileImage, err := app.Repository.GetProfileImageByFilename(filename)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{Data: profileImage}
	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) GetProfileImageByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	profileImage, err := app.Repository.GetProfileImageByUserID(userID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{
		Data: profileImage,
	}
	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) DeleteProfileImageByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	profileImage, err := app.Repository.GetProfileImageByUserID(userID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.Repository.DeleteProfileImageByFilename(profileImage.Filename)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{
		Message: "Profile image successfully deleted",
		Error:   false,
	}

	_ = app.writeJSON(w, http.StatusOK, response)
}
