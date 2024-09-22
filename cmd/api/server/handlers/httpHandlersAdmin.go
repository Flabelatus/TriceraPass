package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/internal/models"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func AdminDeleteAllUsers(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// AdminDeleteUserByID deletes a user from the database by their user ID and returns a success message.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that deletes a user by their user ID.
func AdminDeleteUserByID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from the URL parameters
		IDParam := chi.URLParam(r, "user_id")

		// Delete the user by ID
		err := app.Repository.DeleteUserByID(IDParam)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Prepare and send success response
		response := utils.JSONResponse{
			Error:   false,
			Message: "successfully removed user",
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

func AdminCreateUser(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// GetAllUsers retrieves all users from the database and returns them as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that fetches and returns all users.
func GetAllUsers(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errChan := make(chan error)
		uChan := make(chan []models.User)

		go func() {
			users, err := app.Repository.GetAllUsers()
			if err != nil {
				errChan <- err
			} else {
				uChan <- users
			}
		}()

		select {
		case users := <-uChan:
			_ = utils.WriteJSON(w, http.StatusOK, users)
		case err := <-errChan:
			utils.ErrorJSON(w, err)
		}
	}
}
