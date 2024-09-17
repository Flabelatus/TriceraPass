package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/controllers"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/internal/models"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// PasswordResetPayload represents the payload for resetting a user's password.
type PasswordResetPayload struct {
	UserID      string `json:"user_id"`      // The ID of the user.
	NewPassword string `json:"new_password"` // The new password for the user.
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

		errChan := make(chan error)
		modeChan := make(chan []models.Mode)

		go func() {
			modes, err := app.Repository.GetAllModes()
			if err != nil {
				errChan <- err
			} else {
				modeChan <- modes
			}
		}()

		select {
		case modes := <-modeChan:
			_ = utils.WriteJSON(w, http.StatusOK, modes)
		case err := <-errChan:
			utils.ErrorJSON(w, err)
		}
	}
}

// GetUsersModeByUserID retrieves the mode of a user by their user ID and returns it as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that fetches the user mode by user ID.
func GetUsersModeByUserID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		IDParam := chi.URLParam(r, "user_id")

		errChan := make(chan error)
		userChan := make(chan *models.User)

		go func() {
			user, err := app.Repository.GetUserByID(IDParam)
			if err != nil {
				errChan <- err
			} else {
				userChan <- user
			}
		}()

		select {
		case user := <-userChan:
			mode := user.Mode.Name
			_ = utils.WriteJSON(w, http.StatusOK, mode)
		case err := <-errChan:
			utils.ErrorJSON(w, fmt.Errorf("error getting the user mode - %v", err))
		}
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

// CreateUserMode creates a new user mode and inserts it into the database.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that creates a new user mode.
func CreateUserMode(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var mode *models.Mode

		err := utils.ReadJSON(w, r, &mode)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		modeChanel := make(chan *models.Mode)
		errChanel := make(chan error)

		mode.CreatedAt = time.Now()
		mode.UpdatedAt = time.Now()

		go func() {
			err := app.Repository.CreateMode(mode)
			if err != nil {
				errChanel <- err
			} else {
				modeChanel <- mode
			}
		}()

		select {
		case mode := <-modeChanel:
			response := utils.JSONResponse{
				Message: "mode created",
				Data:    mode,
			}
			_ = utils.WriteJSON(w, http.StatusAccepted, response)
		case err := <-errChanel:
			utils.ErrorJSON(w, err)
		}
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

		IDParam := chi.URLParam(r, "mode_id")
		modeID, err := strconv.Atoi(IDParam)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		var payload *models.Mode
		err = utils.ReadJSON(w, r, &payload)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		errChan := make(chan error)
		modeChan := make(chan *models.Mode)

		go func() {
			mode, err := app.Repository.GetModeByID(int(modeID))
			if err != nil {
				errChan <- err
			}
			mode.Name = payload.Name
			mode.UpdatedAt = time.Now()

			err = app.Repository.UpdateMode(modeID, mode)
			if err != nil {
				errChan <- err
			} else {
				modeChan <- mode
			}
		}()

		select {
		case mode := <-modeChan:
			response := utils.JSONResponse{
				Message: "mode successfully updated",
				Data:    mode,
			}
			_ = utils.WriteJSON(w, http.StatusOK, response)
		case err := <-errChan:
			utils.ErrorJSON(w, err)
		}
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

		IDParam := chi.URLParam(r, "mode_id")
		modeID, err := strconv.Atoi(IDParam)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		errChan := make(chan error)

		go func() {
			err := app.Repository.DeleteModeByID(modeID)
			if err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
		}()

		err = <-errChan
		if err != nil {
			utils.ErrorJSON(w, err)
		}
		response := utils.JSONResponse{
			Error:   false,
			Message: "successfully removed user mode",
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
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

		response := utils.JSONResponse{
			Data: user,
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// DeleteUser deletes a user from the database by their user ID and returns a success message.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function that deletes a user by their user ID.
func DeleteUser(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		IDParam := chi.URLParam(r, "user_id")
		errChan := make(chan error)

		go func() {
			err := app.Repository.DeleteUserByID(IDParam)
			if err != nil {
				errChan <- err
			} else {
				errChan <- nil
			}
		}()

		// Check for errors and return a response
		err := <-errChan
		if err != nil {
			utils.ErrorJSON(w, err)
		}

		response := utils.JSONResponse{
			Error:   false,
			Message: "successfully removed user",
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

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

// SendPasswordResetEmail generates a password reset token and sends an email to the user containing the reset link.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for sending a password reset email.
func SendPasswordResetEmail(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Payload to read user ID from the request body
		type UserPayload struct {
			UserID string `json:"user_id"`
		}

		var payload *UserPayload
		err := utils.ReadJSON(w, r, &payload)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error reading payload - %v", err))
			return
		}

		// Generate a password reset token
		passwordToken := models.PasswordRestToken{
			ID:        uuid.NewString(),
			UserID:    payload.UserID,
			ExpiredAt: 10,
			CreatedAt: time.Now(),
			TokenUsed: false,
		}

		tokenID, err := app.Repository.InsertPasswordToken(&passwordToken)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		fmt.Printf("password reset token generated with id: %s", tokenID)

		// Load environment variables
		_ = godotenv.Load()

		// Fetch the user from the database
		user, err := app.Repository.GetUserByID(payload.UserID)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error getting user from database - %v", err))
			return
		}

		// Send the password reset email using Mailgun
		apiKey := os.Getenv("MAILGUN_API_KEY")
		domain := os.Getenv("MAILGUN_DOMAIN")

		_, err = controllers.SendEmail(domain, apiKey, user.Email, user.UserName, payload.UserID, "password", 1)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error sending email - %v", err))
			return
		}

		// Respond with the token ID
		response := utils.JSONResponse{
			Data:    tokenID,
			Message: "password reset link was sent to your email",
		}

		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// SendForgottenPasswordEmail handles the process of sending a password reset email to a user who forgot their password.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for sending a forgotten password email.
func SendForgottenPasswordEmail(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Payload to read the email address from the request body
		type PayloadEmail struct {
			Email string `json:"email"`
		}

		var emailInPayload *PayloadEmail
		if err := utils.ReadJSON(w, r, &emailInPayload); err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Fetch the user by their email address
		user, err := app.Repository.GetUserByEmail(emailInPayload.Email)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Generate a password reset token
		passwordToken := models.PasswordRestToken{
			ID:        uuid.NewString(),
			UserID:    user.ID,
			ExpiredAt: 10,
			CreatedAt: time.Now(),
			TokenUsed: false,
		}

		tokenID, err := app.Repository.InsertPasswordToken(&passwordToken)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		fmt.Printf("password reset token generated with id: %s", tokenID)

		// Load environment variables
		_ = godotenv.Load()

		// Send the password reset email using Mailgun
		apiKey := os.Getenv("MAILGUN_API_KEY")
		domain := os.Getenv("MAILGUN_DOMAIN")

		_, err = controllers.SendEmail(domain, apiKey, user.Email, user.UserName, user.ID, "password", 1)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error sending email - %v", err))
			return
		}

		// Respond with the token ID
		response := utils.JSONResponse{
			Data:    tokenID,
			Message: "password reset link was sent to your email",
		}

		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// ChangePasswordByUserID handles the process of changing a user's password by their user ID.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for changing a user's password.
func ChangePasswordByUserID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Payload to read the new password and user ID from the request body
		var payload *PasswordResetPayload
		err := utils.ReadJSON(w, r, &payload)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Fetch the existing password
		oldPassword, err := app.Repository.GetUserPasswordByID(payload.UserID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Verify that the new password is not the same as the old one
		isDuplicate, _ := controllers.VerifyPasswordNonDuplicate(oldPassword, payload.NewPassword)
		if isDuplicate {
			utils.ErrorJSON(w, fmt.Errorf("the new password cannot be the same as your existing one"))
			return
		}

		// Update the user's password in the database
		err = app.Repository.ChangePasswordByUserID(payload.UserID, payload.NewPassword)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Send a password change notification email (async)
		go func() {
			_ = godotenv.Load()
			apiKey := os.Getenv("MAILGUN_API_KEY")
			domain := os.Getenv("MAILGUN_DOMAIN")

			user, err := app.Repository.GetUserByID(payload.UserID)
			if err != nil {
				fmt.Println("error getting user:", err)
				return
			}

			_, err = controllers.SendEmail(domain, apiKey, user.Email, user.UserName, user.ID, "passwordChange", 10)
			if err != nil {
				fmt.Println("error sending email:", err)
				return
			}
		}()

		// Respond with success
		response := utils.JSONResponse{Message: "password changed successfully"}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// SetTokenToUsed marks a password reset token as used and sets its expiration.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for marking a password reset token as used.
func SetTokenToUsed(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")

		// Fetch the last password reset token for the user
		passwordToken, err := app.Repository.GetLastPasswordTokenByUserID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Check if the token has already been used or expired
		if passwordToken != nil && passwordToken.IsTokenExpired() {
			if passwordToken.TokenUsed {
				utils.ErrorJSON(w, fmt.Errorf("this password reset token is already used"))
				return
			}

			// Mark the token as used and expire it
			passwordToken.TokenUsed = true
			passwordToken.SetTokenExpire()

			// Update the token in the database
			err = app.Repository.SetTokenToUsed(passwordToken.ID, passwordToken)
			if err != nil {
				utils.ErrorJSON(w, err)
				return
			}

			// Respond with success
			response := utils.JSONResponse{Message: "password reset token was successfully used"}
			_ = utils.WriteJSON(w, http.StatusOK, response)
		}
	}
}

// FetchPasswordTokenByUserID retrieves the most recent password reset token for a user and returns it as a JSON response.
//
// Parameters:
// - app: A pointer to the application context containing repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for fetching a password reset token by user ID.
func FetchPasswordTokenByUserID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		passwordToken, err := app.Repository.GetLastPasswordTokenByUserID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		response := utils.JSONResponse{
			Data: passwordToken,
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}
