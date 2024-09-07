package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"TriceraPass/internal/controllers"
	"TriceraPass/internal/models"
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

func (app *application) GetAllUserModes(w http.ResponseWriter, r *http.Request) {
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
		_ = app.writeJSON(w, http.StatusOK, modes)
	case err := <-errChan:
		app.errorJSON(w, err)
	}
}

func (app *application) GetUsersModeByUserID(w http.ResponseWriter, r *http.Request) {
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
		_ = app.writeJSON(w, http.StatusOK, mode)
	case err := <-errChan:
		app.errorJSON(w, fmt.Errorf("error getting the user mode - %v", err))
	}
}

func (app *application) GetUserModeByID(w http.ResponseWriter, r *http.Request) {
	IDParams := chi.URLParam(r, "mode_id")
	modeID, err := strconv.Atoi(IDParams)
	if err != nil {
		app.errorJSON(w, err)
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
		_ = app.writeJSON(w, http.StatusOK, mode)
	case err := <-errChan:
		app.errorJSON(w, err)
	}
}

func (app *application) AddMissingCreationDate(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	user, err := app.Repository.GetUserByID(userID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var response JSONResponse

	if user.CreatedAt.IsZero() {
		// If CreatedAt is zero, set it to a new date
		user.CreatedAt = time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC)
	}

	// Now update the user in the database
	updatedUser, err := app.Repository.UpdateUser(userID, user)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// Check if the update was successful
	if updatedUser.CreatedAt != user.CreatedAt {
		app.errorJSON(w, errors.New("updated user's created at does not match"))
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
	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) Updateuser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	var payload *models.User

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	user, err := app.Repository.GetUserByID(userID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	user.UserName = payload.UserName
	user.Email = payload.Email

	_, err = app.Repository.UpdateUser(userID, user)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{Data: userID, Message: "user successfully updated"}
	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) CreateUserMode(w http.ResponseWriter, r *http.Request) {
	var mode *models.Mode
	// read response
	err := app.readJSON(w, r, &mode)
	if err != nil {
		app.errorJSON(w, err)
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
		response := JSONResponse{
			Message: "mode created",
			Data:    mode,
		}
		_ = app.writeJSON(w, http.StatusAccepted, response)
	case err := <-errChanel:
		app.errorJSON(w, err)
	}
}

func (app *application) UpdateUserMode(w http.ResponseWriter, r *http.Request) {
	IDParam := chi.URLParam(r, "mode_id")
	modeID, err := strconv.Atoi(IDParam)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload *models.Mode
	err = app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, err)
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
		response := JSONResponse{
			Message: "mode successfully updated",
			Data:    mode,
		}
		_ = app.writeJSON(w, http.StatusOK, response)
	case err := <-errChan:
		app.errorJSON(w, err)

	}
}

func (app *application) DeleteUserMode(w http.ResponseWriter, r *http.Request) {
	IDParam := chi.URLParam(r, "mode_id")
	modeID, err := strconv.Atoi(IDParam)
	if err != nil {
		app.errorJSON(w, err)
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
		app.errorJSON(w, err)
	}
	response := JSONResponse{
		Error:   false,
		Message: "successfully removed user mode",
	}
	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) GetAllUsers(w http.ResponseWriter, r *http.Request) {
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
		_ = app.writeJSON(w, http.StatusOK, users)
	case err := <-errChan:
		app.errorJSON(w, err)
	}
}

func (app *application) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	userEmail := chi.URLParam(r, "user_email")
	user, err := app.Repository.GetUserByEmail(userEmail)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{
		Data: user,
	}

	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request) {
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
		app.errorJSON(w, err)
	}
	response := JSONResponse{
		Error:   false,
		Message: "successfully removed user",
	}
	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) GetConfirmationByID(w http.ResponseWriter, r *http.Request) {
	confirmationID := chi.URLParam(r, "confirmation_id")
	confirmationModel, err := app.Repository.GetConfirmationByID(confirmationID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{Data: confirmationModel}
	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) GetMostRecentConfirmation(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	confirmation, err := app.Repository.GetLastConfirmation(userID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{Data: confirmation}
	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) ConfirmUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")

	confirmation, err := app.Repository.GetLastConfirmation(userID)
	if err != nil {
		app.errorJSON(w, fmt.Errorf("error 1 - %v", err))
		return
	}
	fmt.Println(confirmation.IsExpired())
	if confirmation != nil && !confirmation.IsExpired() {
		if confirmation.Confirmed {
			app.errorJSON(w, fmt.Errorf("user is already confirmed"))
			return
		}

		confirmation.Confirmed = true
		confirmation.SetExpire()

		err = app.Repository.ConfirmUser(confirmation.ID, confirmation)
		if err != nil {
			app.errorJSON(w, fmt.Errorf("error 2 - %v", err))
			return
		}

		fmt.Println(confirmation.IsExpired())

		response := JSONResponse{Message: "user confirmed successfully"}
		_ = app.writeJSON(w, http.StatusOK, response)
	}
}

func (app *application) GetLastConfirmation(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	confrimation, err := app.Repository.GetLastConfirmation(userID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{
		Data: confrimation,
	}

	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) SendPasswordResetEmail(w http.ResponseWriter, r *http.Request) {

	// read payload
	type UserPaylaod struct {
		UserID string `json:"user_id"`
	}

	var payload *UserPaylaod
	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, fmt.Errorf("error reading payload - %v", err))
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
		app.errorJSON(w, err)
		return
	}

	fmt.Printf("password reset token generated with id: %s", tokenID)

	_ = godotenv.Load()

	user, err := app.Repository.GetUserByID(payload.UserID)
	if err != nil {
		app.errorJSON(w, fmt.Errorf("error getting user from database - %v", err))
		return
	}

	// Send Mailgun here
	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")

	_, err = controllers.SendEmail(domain, apiKey, user.Email, user.UserName, payload.UserID, "password", 1)
	if err != nil {
		app.errorJSON(w, fmt.Errorf("error sending email - %v", err))
		return
	}

	response := JSONResponse{
		Data:    tokenID,
		Message: "password reset link was sent to your email",
	}

	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) SendForgottenPasswordEmail(w http.ResponseWriter, r *http.Request) {
	type PayloadEmail struct {
		Email string `json:"email"`
	}

	var emailInPayload *PayloadEmail
	if err := app.readJSON(w, r, &emailInPayload); err != nil {
		app.errorJSON(w, err)
		return
	}

	user, err := app.Repository.GetUserByEmail(emailInPayload.Email)
	if err != nil {
		app.errorJSON(w, err)
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
		app.errorJSON(w, err)
		return
	}

	fmt.Printf("password reset token generated with id: %s", tokenID)

	_ = godotenv.Load()

	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")

	_, err = controllers.SendEmail(domain, apiKey, user.Email, user.UserName, user.ID, "password", 1)
	if err != nil {
		app.errorJSON(w, fmt.Errorf("error sending email - %v", err))
		return
	}

	response := JSONResponse{
		Data:    tokenID,
		Message: "password reset link was sent to your email",
	}

	_ = app.writeJSON(w, http.StatusOK, response)
}

func (app *application) ChangePasswordByUserID(w http.ResponseWriter, r *http.Request) {

	// read payload
	var payload *PasswordResetPayload
	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// get the existing password
	oldPassword, err := app.Repository.GetUserPasswordByID(payload.UserID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// verify that the new password and old one are not the same
	isDuplicate, _ := controllers.VerifyPasswordNonDuplicate(oldPassword, payload.NewPassword)

	if isDuplicate {
		app.errorJSON(w, fmt.Errorf("the new password can not be the same as your existing one"))
		return
	}

	// change the password
	err = app.Repository.ChangePasswordByUserID(payload.UserID, payload.NewPassword)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// write response
	response := JSONResponse{Message: "password changed successfully"}
	_ = app.writeJSON(w, http.StatusOK, response)
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

func (app *application) SetTokenToUsed(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	passwordToken, err := app.Repository.GetLastPasswordTokenByUserID(userID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	if passwordToken != nil && passwordToken.IsTokenExpired() {
		if passwordToken.TokenUsed {
			app.errorJSON(w, fmt.Errorf("this password reset token is already used"))
			return
		}

		passwordToken.TokenUsed = true
		passwordToken.SetTokenExpire()
		err = app.Repository.SetTokenToUsed(passwordToken.ID, passwordToken)
		if err != nil {
			app.errorJSON(w, err)
			return
		}
		response := JSONResponse{Message: "password reset token was successfully used"}
		_ = app.writeJSON(w, http.StatusOK, response)
	}
}

func (app *application) FetchPasswordTokenByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	passwordToken, err := app.Repository.GetLastPasswordTokenByUserID(userID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	response := JSONResponse{
		Data: passwordToken,
	}

	_ = app.writeJSON(w, http.StatusOK, response)
}
