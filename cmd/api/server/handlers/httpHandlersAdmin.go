package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/internal/models"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// Type to cast for admin user
type UserDataAsAdmin struct {
	ID        string      `gorm:"type:uuid;primary_key"`
	CreatedAt time.Time   `json:"created_at"`
	DeletedAt time.Time   `json:"deleted_at"`
	UserName  string      `json:"username"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Email     string      `json:"email"`
	Mode      models.Mode `gorm:"foreignKey:UserID" json:"mode,omitempty"`
}

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
		users, err := app.Repository.GetAllUsers()
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		var usersInResponse []UserDataAsAdmin

		for _, u := range users {
			userInResponse := UserDataAsAdmin{
				ID:        u.ID,
				CreatedAt: u.CreatedAt,
				DeletedAt: u.DeletedAt,
				UserName:  u.UserName,
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Email:     u.Email,
				Mode:      u.Mode,
			}

			usersInResponse = append(usersInResponse, userInResponse)
		}

		response := utils.JSONResponse{
			Data: usersInResponse,
		}

		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}
