package handlers

import (
	"TriceraPass/cmd/api/utils"
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/controllers"
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

type PasswordResetPayload struct {
	UserID      string `json:"user_id"`
	NewPassword string `json:"new_password"`
}

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

		// Now update the user in the database
		updatedUser, err := app.Repository.UpdateUser(userID, user)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Check if the update was successful
		if updatedUser.CreatedAt != user.CreatedAt {
			utils.ErrorJSON(w, errors.New("updated user's created at does not match"))
			return
		}

		// Construct response
		response.Data = userID
		if user.CreatedAt.IsZero() {
			response.Message = "users creation date successfully set"
		} else {
			response.Message = "user already has a valid date"
		}

		// Write response
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

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

func CreateUserMode(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var mode *models.Mode
		// read response
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

func GetUserByEmail(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userEmail := chi.URLParam(r, "user_email")
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

func GetConfirmationByID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		confirmationID := chi.URLParam(r, "confirmation_id")
		confirmationModel, err := app.Repository.GetConfirmationByID(confirmationID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		response := utils.JSONResponse{Data: confirmationModel}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

func GetMostRecentConfirmation(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := chi.URLParam(r, "user_id")
		confirmation, err := app.Repository.GetLastConfirmation(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		response := utils.JSONResponse{Data: confirmation}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

func ConfirmUser(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := chi.URLParam(r, "user_id")

		confirmation, err := app.Repository.GetLastConfirmation(userID)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error 1 - %v", err))
			return
		}
		fmt.Println(confirmation.IsExpired())
		if confirmation != nil && !confirmation.IsExpired() {
			if confirmation.Confirmed {
				utils.ErrorJSON(w, fmt.Errorf("user is already confirmed"))
				return
			}

			confirmation.Confirmed = true
			confirmation.SetExpire()

			err = app.Repository.ConfirmUser(confirmation.ID, confirmation)
			if err != nil {
				utils.ErrorJSON(w, fmt.Errorf("error 2 - %v", err))
				return
			}

			fmt.Println(confirmation.IsExpired())

			response := utils.JSONResponse{Message: "user confirmed successfully"}
			_ = utils.WriteJSON(w, http.StatusOK, response)
		}
	}
}

func GetLastConfirmation(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		confrimation, err := app.Repository.GetLastConfirmation(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		response := utils.JSONResponse{
			Data: confrimation,
		}

		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

func SendPasswordResetEmail(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read payload
		type UserPaylaod struct {
			UserID string `json:"user_id"`
		}

		var payload *UserPaylaod
		err := utils.ReadJSON(w, r, &payload)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error reading payload - %v", err))
			return
		}

		// generate a token
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

		_ = godotenv.Load()

		user, err := app.Repository.GetUserByID(payload.UserID)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error getting user from database - %v", err))
			return
		}

		// Send Mailgun here
		apiKey := os.Getenv("MAILGUN_API_KEY")
		domain := os.Getenv("MAILGUN_DOMAIN")

		_, err = controllers.SendEmail(domain, apiKey, user.Email, user.UserName, payload.UserID, "password", 1)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error sending email - %v", err))
			return
		}

		response := utils.JSONResponse{
			Data:    tokenID,
			Message: "password reset link was sent to your email",
		}

		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

func SendForgottenPasswordEmail(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type PayloadEmail struct {
			Email string `json:"email"`
		}

		var emailInPayload *PayloadEmail
		if err := utils.ReadJSON(w, r, &emailInPayload); err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		user, err := app.Repository.GetUserByEmail(emailInPayload.Email)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// generate a token
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

		_ = godotenv.Load()

		apiKey := os.Getenv("MAILGUN_API_KEY")
		domain := os.Getenv("MAILGUN_DOMAIN")

		_, err = controllers.SendEmail(domain, apiKey, user.Email, user.UserName, user.ID, "password", 1)
		if err != nil {
			utils.ErrorJSON(w, fmt.Errorf("error sending email - %v", err))
			return
		}

		response := utils.JSONResponse{
			Data:    tokenID,
			Message: "password reset link was sent to your email",
		}

		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

func ChangePasswordByUserID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read payload
		var payload *PasswordResetPayload
		err := utils.ReadJSON(w, r, &payload)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// get the existing password
		oldPassword, err := app.Repository.GetUserPasswordByID(payload.UserID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// verify that the new password and old one are not the same
		isDuplicate, _ := controllers.VerifyPasswordNonDuplicate(oldPassword, payload.NewPassword)

		if isDuplicate {
			utils.ErrorJSON(w, fmt.Errorf("the new password can not be the same as your existing one"))
			return
		}

		// change the password
		err = app.Repository.ChangePasswordByUserID(payload.UserID, payload.NewPassword)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// write response
		response := utils.JSONResponse{Message: "password changed successfully"}
		_ = utils.WriteJSON(w, http.StatusOK, response)
		go func() {
			// Load environment variables
			_ = godotenv.Load()

			apiKey := os.Getenv("MAILGUN_API_KEY")
			domain := os.Getenv("MAILGUN_DOMAIN")

			user, err := app.Repository.GetUserByID(payload.UserID)
			if err != nil {
				// Log the error, but do not return it to prevent disrupting the response
				fmt.Println("error getting user:", err)
				return
			}

			// Send the email with a delay of 10 seconds
			_, err = controllers.SendEmail(domain, apiKey, user.Email, user.UserName, user.ID, "passwordChange", 10)
			if err != nil {
				// Log the error, but do not return it to prevent disrupting the response
				fmt.Println("error sending email:", err)
				return
			}
		}()
	}
}

func SetTokenToUsed(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "user_id")
		passwordToken, err := app.Repository.GetLastPasswordTokenByUserID(userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		if passwordToken != nil && passwordToken.IsTokenExpired() {
			if passwordToken.TokenUsed {
				utils.ErrorJSON(w, fmt.Errorf("this password reset token is already used"))
				return
			}
			passwordToken.TokenUsed = true
			passwordToken.SetTokenExpire()
			err = app.Repository.SetTokenToUsed(passwordToken.ID, passwordToken)
			if err != nil {
				utils.ErrorJSON(w, err)
				return
			}
			response := utils.JSONResponse{Message: "password reset token was successfully used"}
			_ = utils.WriteJSON(w, http.StatusOK, response)
		}
	}
}

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
