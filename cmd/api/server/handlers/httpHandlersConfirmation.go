package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/utils"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetConfirmationByID retrieves a confirmation record by its ID and returns it as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that fetches a confirmation record by its ID.
func GetConfirmationByID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		confirmationID := chi.URLParam(r, "confirmation_id")

		// Fetch the confirmation by its ID
		confirmationModel, err := app.Repository.GetConfirmationByID(confirmationID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Return the confirmation as JSON
		response := utils.JSONResponse{Data: confirmationModel}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// GetMostRecentConfirmation retrieves the most recent confirmation for a user by their user ID and returns it as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that fetches the most recent confirmation for a user.
func GetMostRecentConfirmation(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")

		// Fetch the most recent confirmation for the user
		confirmation, err := app.Repository.GetLastConfirmation(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Return the confirmation as JSON
		response := utils.JSONResponse{Data: confirmation}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// ConfirmUser confirms a user's account if the confirmation is valid and has not expired.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that confirms a user's account.
func ConfirmUser(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")

		// Fetch the most recent confirmation for the user
		confirmation, err := app.Repository.GetLastConfirmation(userID)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error 1 - %v", err))
			return
		}

		// Check if the confirmation is valid and not expired
		fmt.Println(confirmation.IsExpired())
		if confirmation != nil && !confirmation.IsExpired() {
			if confirmation.Confirmed {
				utils.ErrorJSON(w, fmt.Errorf("user is already confirmed"))
				return
			}

			// Mark the user as confirmed and expire the confirmation
			confirmation.Confirmed = true
			confirmation.SetExpire()

			err = app.Repository.ConfirmUser(confirmation.ID, confirmation)
			if err != nil {
				utils.ErrorJSON(w, fmt.Errorf("error 2 - %v", err))
				return
			}

			fmt.Println(confirmation.IsExpired())

			// Return a success message
			response := utils.JSONResponse{Message: "user confirmed successfully"}
			_ = utils.WriteJSON(w, http.StatusOK, response)
		}
	}
}

// GetLastConfirmation retrieves the most recent confirmation for a user by their user ID and returns it as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for fetching the last confirmation for a user.
func GetLastConfirmation(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		confirmation, err := app.Repository.GetLastConfirmation(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		response := utils.JSONResponse{
			Data: confirmation,
		}

		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}
