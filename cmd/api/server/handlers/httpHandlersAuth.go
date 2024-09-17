package handlers

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/auth"
	"TriceraPass/cmd/api/controllers"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/internal/models"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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

// Home returns an HTTP handler function that serves the home page or API status information.
// It dynamically loads configuration and presents either an HTML page or a JSON response based on the Accept header.
//
// Parameters:
// - app: A pointer to the application context containing configuration and services.
//
// Returns:
// - http.HandlerFunc: An HTTP handler function.
func Home(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Load environment variables
		err := godotenv.Load()
		if err != nil {
			log.Fatal(fmt.Printf("Cannot locate the env file: %v", err))
		}

		// Define a struct to hold the response data
		type Info struct {
			API struct {
				Name        string        `json:"name"`
				Version     string        `json:"version"`
				Description template.HTML `json:"description"`
			} `json:"api"`
			Application struct {
				ClientName string `json:"client_name"`
				Domain     string `json:"domain"`
			} `json:"application"`
			Styles struct {
				HeaderBackground string
				HeaderColor      string
				HeaderFont       string
				BodyFont         string
				BodyColor        string
				BodyBackground   string
				HeaderFontSize   string
			}
		}

		// Load configuration from the config file
		confFile := os.Getenv("CONFIG_FILE")
		conf, err := application.LoadConfig(confFile)
		if err != nil {
			http.Error(w, "Could not load configuration file", http.StatusInternalServerError)
			return
		}

		description := template.HTML(strings.ReplaceAll(conf.API.Description, "!", "! <br>"))

		info := &Info{
			API: struct {
				Name        string        `json:"name"`
				Version     string        `json:"version"`
				Description template.HTML `json:"description"`
			}{
				Name:        conf.API.Name,
				Version:     conf.API.Version,
				Description: description,
			},
			Application: struct {
				ClientName string `json:"client_name"`
				Domain     string `json:"domain"`
			}{
				ClientName: conf.Application.ClientName,
				Domain:     conf.Application.Domain,
			},
			Styles: struct {
				HeaderBackground string
				HeaderColor      string
				HeaderFont       string
				BodyFont         string
				BodyColor        string
				BodyBackground   string
				HeaderFontSize   string
			}{
				HeaderBackground: conf.Styles.HeaderBackground,
				HeaderColor:      conf.Styles.HeaderColor,
				HeaderFont:       conf.Styles.HeaderFont,
				BodyFont:         conf.Styles.BodyFont,
				BodyColor:        conf.Styles.BodyColor,
				BodyBackground:   conf.Styles.BodyBackground,
				HeaderFontSize:   conf.Styles.HeaderFontSize,
			},
		}

		// Check the Accept header and return either JSON or HTML
		acceptHeader := r.Header.Get("Accept")

		if strings.Contains(acceptHeader, "application/json") {
			payload := struct {
				Status  string `json:"status"`
				Message string `json:"message"`
				Info    *Info  `json:"info"`
			}{
				Status:  "active",
				Message: "Authentication service is up and running!",
				Info:    info,
			}

			err = utils.WriteJSON(w, http.StatusOK, payload)
			if err != nil {
				http.Error(w, "Failed to write JSON response", http.StatusInternalServerError)
			}
		} else {
			tmpl, err := template.ParseFiles("./template/index.html")
			if err != nil {
				http.Error(w, fmt.Sprintf("Error parsing template: %v", err), http.StatusInternalServerError)
				return
			}

			err = tmpl.Execute(w, info)
			if err != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		}
	}
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

		// Create a mode record for the user
		userMode := models.Mode{
			Name:   "1212",
			UserID: newUser.ID,
		}
		err = app.Repository.CreateMode(&userMode)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Send confirmation email via Mailgun
		apiKey := os.Getenv("MAILGUN_API_KEY")
		domain := os.Getenv("MAILGUN_DOMAIN")

		_, err = controllers.SendEmail(domain, apiKey, newUser.Email, newUser.UserName, userID, "confirmation", 1)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
	}
}

// GetUserByID retrieves a user by their ID and sends the user information in the response.
// It concurrently fetches the user from the database and returns the result.
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

		errChan := make(chan error)
		userChan := make(chan *models.User)

		// Concurrently fetch the user from the database
		go func() {
			user, err := app.Repository.GetUserByID(idParam)
			if err != nil {
				errChan <- err
			} else {
				userChan <- user
			}
		}()

		// Wait for either the user or an error
		select {
		case user := <-userChan:
			_ = utils.WriteJSON(w, http.StatusOK, user)
		case err := <-errChan:
			utils.ErrorJSON(w, err)
		}
	}
}
