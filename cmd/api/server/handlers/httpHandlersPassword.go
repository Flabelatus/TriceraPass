package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/controllers"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/internal/models"
	"fmt"
	"net/http"
	"os"
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
