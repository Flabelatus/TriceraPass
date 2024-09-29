package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/controllers"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/internal/models"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// User struct without senative data
type RegularUserResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserName  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

// GetUserByID retrieves a user by their ID and sends the user information in the response.
// It fetches the user from the database and returns the result.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for the get user by ID route.
func GetUserByID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract user ID from the URL parameters
		idParam := chi.URLParam(r, "user_id")
		fmt.Println(idParam)

		// Fetch the user from the database
		user, err := app.Repository.GetUserByID(idParam)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		var response utils.JSONResponse

		// Check if the user is an admin
		if user.Mode.Name == "admin" {
			// Return the full user data for admin
			response = utils.JSONResponse{
				Data: user,
			}
		} else {
			// Return the regular user data without sensitive information
			response = utils.JSONResponse{
				Data: RegularUserResponse{
					ID:        user.ID,
					CreatedAt: user.CreatedAt,
					UserName:  user.UserName,
					FirstName: user.FirstName,
					LastName:  user.LastName,
					Email:     user.Email,
				},
			}
		}

		// Write the JSON response
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// GetUserByEmail retrieves a user by their email and returns the user information as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that fetches a user by their email.
func GetUserByEmail(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userEmail := chi.URLParam(r, "user_email")

		// Fetch user by email from the database
		user, err := app.Repository.GetUserByEmail(userEmail)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		var response utils.JSONResponse

		if user.Mode.Name == "admin" {
			response = utils.JSONResponse{
				Data: user,
			}
		} else {
			response = utils.JSONResponse{
				Data: RegularUserResponse{
					ID:        user.ID,
					CreatedAt: user.CreatedAt,
					UserName:  user.UserName,
					FirstName: user.FirstName,
					LastName:  user.LastName,
					Email:     user.Email,
				},
			}
		}
		fmt.Print(response)
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// AddMissingCreationDate adds a missing creation date to a user if it doesn't exist and updates the user in the database.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that updates the user's creation date if missing.
func AddMissingCreationDate(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := chi.URLParam(r, "user_id")
		user, err := app.Repository.GetUserByID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		var response utils.JSONResponse

		if user.CreatedAt.IsZero() {
			// If CreatedAt is zero, set it to a new date
			user.CreatedAt = time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
		}

		// Update the user in the database
		updatedUser, err := app.Repository.UpdateUser(userID, user)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		if updatedUser.CreatedAt != user.CreatedAt {
			utils.ErrorJSON(w, errors.New("updated user's created at does not match"))
			return
		}

		// Construct and return the response
		response.Data = userID
		if user.CreatedAt.IsZero() {
			response.Message = "user's creation date successfully set"
		} else {
			response.Message = "user already has a valid date"
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// UpdateUser updates a user's information in the database based on the provided user ID and request payload.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that updates the user's information.
func Updateuser(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := chi.URLParam(r, "user_id")
		var payload *models.User

		err := utils.ReadJSON(w, r, &payload)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		user, err := app.Repository.GetUserByID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Update user information
		user.UserName = payload.UserName
		user.Email = payload.Email

		_, err = app.Repository.UpdateUser(userID, user)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		response := utils.JSONResponse{Data: userID, Message: "user successfully updated"}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

func DeleteOwnUserData(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the identity from jwt claims
		_, claims, err := app.Auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID := claims.Subject

		// Get the user based on the user_id in the url params
		userIdFromUrl := chi.URLParam(r, "user_id")

		if userID != userIdFromUrl {
			utils.ErrorJSON(w, errors.New("the user id does not match with your identity"))
			return
		}

		// Delete the user (this will also delete the associated Mode due to cascading)
		err = app.Repository.DeleteUserByID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Delete the user's profile image as well from the database
		err = app.Repository.DeleteProfileImageByUserID(userID)
		if err != nil {
			log.Println("Deleting the user profile image from database")
			utils.ErrorJSON(w, err)
			return
		}

		// Delete the image file as well
		err = controllers.DeleteProfileImageFile(app.Root, userID)
		if err != nil {
			log.Println("Deleting the user profile image file")
			utils.ErrorJSON(w, fmt.Errorf("could not delete the image file: %v", err))
			return
		}

		resp := utils.JSONResponse{
			Data: "Successfully deleted the user and the associated mode",
		}

		_ = utils.WriteJSON(w, http.StatusAccepted, resp)
	}
}
