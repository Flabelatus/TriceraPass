package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/internal/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// GetUsersModeByUserID retrieves the mode of a user by their user ID and returns it as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that fetches the user mode by user ID.
func GetUsersModeByUserID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract user ID from the URL parameters
		IDParam := chi.URLParam(r, "user_id")

		// Fetch the user
		user, err := app.Repository.GetUserByID(IDParam)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error getting the user mode - %v", err))
			return
		}

		// Extract the mode from the user object
		mode := user.Mode.Name

		// Write the mode as the JSON response
		_ = utils.WriteJSON(w, http.StatusOK, mode)
	}
}

// GetUserModeByID retrieves a user mode by its mode ID and returns it as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that fetches a user mode by its ID.
func GetUserModeByID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		IDParams := chi.URLParam(r, "mode_id")
		modeID, err := strconv.Atoi(IDParams)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		errChan := make(chan error)
		modeChan := make(chan *models.Mode)

		go func() {
			mode, err := app.Repository.GetModeByID(modeID)
			if err != nil {
				errChan <- err
			} else {
				modeChan <- mode
			}
		}()

		select {
		case mode := <-modeChan:
			_ = utils.WriteJSON(w, http.StatusOK, mode)
		case err := <-errChan:
			utils.ErrorJSON(w, err)
		}
	}
}

// GetAllUserModes retrieves all user modes from the database and returns them as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that fetches and returns all user modes.
func GetAllUserModes(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Fetch all modes
		modes, err := app.Repository.GetAllModes()
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Write the modes as the JSON response
		_ = utils.WriteJSON(w, http.StatusOK, modes)
	}
}

// CreateUserMode creates a new user mode and inserts it into the database.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that creates a new user mode.
func CreateUserMode(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Read the incoming mode data from the request body
		var mode *models.Mode
		err := utils.ReadJSON(w, r, &mode)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Set the creation and update timestamps
		mode.CreatedAt = time.Now()
		mode.UpdatedAt = time.Now()

		// Create the mode
		err = app.Repository.CreateMode(mode)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Prepare and send success response
		response := utils.JSONResponse{
			Message: "mode created",
			Data:    mode,
		}
		_ = utils.WriteJSON(w, http.StatusAccepted, response)
	}
}

// UpdateUserMode updates an existing user mode based on the provided mode ID and request payload.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that updates a user mode.
func UpdateUserMode(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract mode ID from the URL parameters
		IDParam := chi.URLParam(r, "mode_id")
		modeID, err := strconv.Atoi(IDParam)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Read the incoming payload from the request body
		var payload *models.Mode
		err = utils.ReadJSON(w, r, &payload)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Fetch the mode by mode ID
		mode, err := app.Repository.GetModeByID(int(modeID))
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Update the mode object with new data
		mode.Name = payload.Name
		mode.UpdatedAt = time.Now()

		// Update the mode in the repository
		err = app.Repository.UpdateMode(modeID, mode)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Prepare and send success response
		response := utils.JSONResponse{
			Message: "mode successfully updated",
			Data:    mode,
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// DeleteUserMode deletes a user mode by its mode ID from the database.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that deletes a user mode.
func DeleteUserMode(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract mode ID from the URL parameters
		IDParam := chi.URLParam(r, "mode_id")
		modeID, err := strconv.Atoi(IDParam)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Delete the mode
		err = app.Repository.DeleteModeByID(modeID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Prepare and send success response
		response := utils.JSONResponse{
			Error:   false,
			Message: "successfully removed user mode",
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}
