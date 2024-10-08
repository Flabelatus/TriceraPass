// Package handlers contains the HTTP handler functions responsible for
// processing incoming HTTP requests and returning appropriate responses.
// Handlers bridge the gap between the routing logic and the core application logic,
// and interact with services, repositories, and utilities.
package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/auth"
	"TriceraPass/cmd/api/controllers"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/internal/models"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type AuthHandlers struct {
}

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Substitution struct {
	Var   string `json:"var"`
	Value string `json:"value"`
}

type Variable struct {
	Email         string         `json:"email"`
	Substitutions []Substitution `json:"substitutions"`
}

type EmailPayload struct {
	From      EmailAddress   `json:"from"`
	To        []EmailAddress `json:"to"`
	Subject   string         `json:"subject"`
	Text      string         `json:"text"`
	Html      string         `json:"html"`
	Variables []Variable     `json:"variables"`
}

// Authenticate handles user authentication by verifying the email and password.
// If successful, it generates a new JWT token pair and returns it in the response.
//
// Parameters:
// - app: A pointer to the application context containing authentication logic and repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for the login route.
func Authenticate(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestPayload struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err := utils.ReadJSON(w, r, &requestPayload)
		if err != nil {
			utils.ErrorJSON(w, err, http.StatusBadRequest)
			return
		}

		// Fetch user by email
		user, err := app.Repository.GetUserByEmail(requestPayload.Email)
		if err != nil {
			utils.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
			return
		}

		// Check if the provided password matches the stored password hash
		valid, err := user.PasswordMatches(requestPayload.Password)
		if err != nil || !valid {
			utils.ErrorJSON(w, errors.New("invalid email or password"), http.StatusUnauthorized)
			return
		}

		// Create a JWT user and generate token pairs
		u := auth.JwtUser{
			ID:        user.ID,
			UserName:  user.UserName,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}

		tokens, err := app.Auth.GenerateTokenPair(&u)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Set refresh token in a cookie
		refreshCookie := app.Auth.GetRefreshCookie(tokens.RefreshToken)
		http.SetCookie(w, refreshCookie)

		utils.WriteJSON(w, http.StatusAccepted, tokens)
	}
}

// RefreshToken handles the process of refreshing a user's JWT tokens using the refresh token.
// It reads the refresh token from cookies, verifies it, and generates a new token pair.
//
// Parameters:
// - app: A pointer to the application context containing authentication logic and repositories.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for the refresh token route.
func RefreshToken(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, cookie := range r.Cookies() {
			if cookie.Name == app.Auth.CookieName {
				claims := &auth.Claims{}
				refreshToken := cookie.Value

				// Parse and verify the refresh token
				_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte(app.JWTSecret), nil
				})
				if err != nil {
					utils.ErrorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
					return
				}

				// Fetch user by ID from the token claims
				userID := claims.Subject
				user, err := app.Repository.GetUserByID(userID)
				if err != nil {
					utils.ErrorJSON(w, fmt.Errorf("unknown user - %v", err), http.StatusUnauthorized)
					return
				}

				// Generate new token pairs
				u := auth.JwtUser{
					ID:        user.ID,
					UserName:  user.UserName,
					FirstName: user.FirstName,
					LastName:  user.LastName,
				}
				tokenPairs, err := app.Auth.GenerateTokenPair(&u)
				if err != nil {
					utils.ErrorJSON(w, errors.New("error generating tokens"), http.StatusUnauthorized)
					return
				}

				// Set new refresh token in the cookie
				http.SetCookie(w, app.Auth.GetRefreshCookie(tokenPairs.RefreshToken))

				utils.WriteJSON(w, http.StatusOK, tokenPairs)
			}
		}
	}
}

// Logout handles user logout by expiring the refresh token cookie and returning a success response.
//
// Parameters:
// - app: A pointer to the application context containing authentication logic.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for the logout route.
func Logout(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, app.Auth.GetExpiredRefreshCookie())
		w.WriteHeader(http.StatusAccepted)
		response := utils.JSONResponse{
			Message: "successfully logged out",
		}
		_ = utils.WriteJSON(w, http.StatusOK, response)
	}
}

// RegisterNewUser handles the process of registering a new user.
// It reads the user data from the request body, hashes the user's password,
// saves the user to the database, and sends a confirmation email.
//
// Parameters:
// - app: A pointer to the application context containing repositories and services.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function for the user registration route.
func RegisterNewUser(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Load environment variables
		err := godotenv.Load()
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Parse the request body into a new User model
		newUser := models.User{ID: uuid.NewString(), CreatedAt: time.Now()}
		err = utils.ReadJSON(w, r, &newUser)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Hash the user's password
		hashedPassword, err := controllers.HashAPassword(newUser.Password)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		newUser.Password = hashedPassword
		newUser.CreatedAt = time.Now()

		// Save the user to the database
		userID, err := app.Repository.CreateUser(&newUser)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		filename, profilePath, err := controllers.UploadDefaultProfile(app.Root, userID)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Insert the profile image record into the database
		image := models.ProfileImage{Filename: filename, FilePath: profilePath, UserID: userID}
		_, err = app.Repository.InsertProfileImage(&image)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Return success response with user ID
		response := utils.JSONResponse{
			Error:   false,
			Message: "user successfully created",
			Data:    userID,
		}
		_ = utils.WriteJSON(w, http.StatusCreated, response)

		// Create a new user confirmation record
		newConfirmation := models.UserConfirmation{
			ID:        uuid.NewString(),
			CreatedAt: time.Now(),
			Confirmed: false,
			ExpiredAt: time.Now().UTC().Add(30 * time.Minute).Unix(),
			UserID:    userID,
		}

		fmt.Println("Is Expired: ", newConfirmation.IsExpired())

		// Insert the confirmation into the database
		_, err = app.Repository.InsertConfirmation(&newConfirmation)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		users, err := app.Repository.GetAllUsers()
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		var modeName string

		if len(users) > 1 {
			modeName = "default"

		} else {
			modeName = "admin"
		}

		// Create a mode record for the user
		userMode := models.Mode{
			Name:   modeName,
			UserID: newUser.ID,
		}

		err = app.Repository.CreateMode(&userMode)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Send confirmation email via Mailgun
		apiKey := os.Getenv("MAIL_SERVER_API_KEY")
		domain := os.Getenv("MAIL_SERVER_DOMAIN")

		_, err = controllers.SendEmail(domain, apiKey, newUser.Email, newUser.UserName, userID, "confirmation", 1)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
	}
}
