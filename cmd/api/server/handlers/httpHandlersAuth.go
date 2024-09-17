package handlers

import (
	// "TriceraPass/cmd/api"
	// "TriceraPass/cmd/api"
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/auth"
	"TriceraPass/cmd/api/utils"
	"TriceraPass/cmd/api/controllers"
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

// Request/Response Handlers
func Home(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// State that app is running
		err := godotenv.Load()
		if err != nil {
			log.Fatal(fmt.Printf("Can not locate the env file: %v", err))
		}

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
		// Check the request header to see if the client expects JSON or HTML
		acceptHeader := r.Header.Get("Accept")

		if strings.Contains(acceptHeader, "application/json") {
			// If the client expects JSON, return the JSON response (API)
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
			// Otherwise, serve the HTML template (for browsers)
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

// Authentication Handlers
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

		user, err := app.Repository.GetUserByEmail(requestPayload.Email)
		if err != nil {
			utils.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
			return
		}

		valid, err := user.PasswordMatches(requestPayload.Password)
		if err != nil || !valid {
			utils.ErrorJSON(w, errors.New("invalid email or password"), http.StatusUnauthorized)
			return
		}

		// create a jwt user
		u := auth.JwtUser{
			ID:        user.ID,
			UserName:  user.UserName,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}

		// generate tokens
		tokens, err := app.Auth.GenerateTokenPair(&u)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		refreshCookie := app.Auth.GetRefreshCookie(tokens.RefreshToken)
		http.SetCookie(w, refreshCookie)

		utils.WriteJSON(w, http.StatusAccepted, tokens)
	}
}

func RefreshToken(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, cookie := range r.Cookies() {
			if cookie.Name == app.Auth.CookieName {
				claims := &auth.Claims{}
				refreshToken := cookie.Value

				// parse the token to get the claims
				_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
					return []byte(app.JWTSecret), nil
				})
				if err != nil {
					utils.ErrorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
					return
				}

				// get the User ID from the token claims
				userID := claims.Subject
				// if err != nil {
				// 	utils.ErrorJSON(w, fmt.Errorf("unknown user - %v", err), http.StatusUnauthorized)
				// }
				user, err := app.Repository.GetUserByID(userID)
				if err != nil {
					utils.ErrorJSON(w, fmt.Errorf("unknown user - %v", err), http.StatusUnauthorized)
					return
				}

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

				http.SetCookie(w, app.Auth.GetRefreshCookie(tokenPairs.RefreshToken))

				utils.WriteJSON(w, http.StatusOK, tokenPairs)
			}
		}
	}
}

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

// User Handlers
func RegisterNewUser(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := godotenv.Load()
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		newUser := models.User{ID: uuid.NewString(), CreatedAt: time.Now()}
		err = utils.ReadJSON(w, r, &newUser)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		hashedPassword, err := controllers.HashAPassword(newUser.Password)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		newUser.Password = hashedPassword
		newUser.CreatedAt = time.Now()
		userID, err := app.Repository.CreateUser(&newUser)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
		response := utils.JSONResponse{
			Error:   false,
			Message: "user successfully created",
			Data:    userID,
		}
		_ = utils.WriteJSON(w, http.StatusCreated, response)

		newConfirmation := models.UserConfirmation{
			ID:        uuid.NewString(),
			CreatedAt: time.Now(),
			Confirmed: false,
			ExpiredAt: time.Now().UTC().Add(30 * time.Minute).Unix(),
			UserID:    userID,
		}

		fmt.Println("Is Expired: ", newConfirmation.IsExpired())

		_, err = app.Repository.InsertConfirmation(&newConfirmation)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		userMode := models.Mode{
			Name:   "1212",
			UserID: newUser.ID,
		}
		err = app.Repository.CreateMode(&userMode)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}

		// Send Mailgun here
		apiKey := os.Getenv("MAILGUN_API_KEY")
		domain := os.Getenv("MAILGUN_DOMAIN")

		_, err = controllers.SendEmail(domain, apiKey, newUser.Email, newUser.UserName, userID, "confirmation", 1)
		if err != nil {
			utils.ErrorJSON(w, err)
			return
		}
	}
}

func GetUserByID(app *application.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		idParam := chi.URLParam(r, "user_id")
		fmt.Println(idParam)

		errChan := make(chan error)
		userChan := make(chan *models.User)

		go func() {
			user, err := app.Repository.GetUserByID(idParam)
			if err != nil {
				errChan <- err
			} else {
				userChan <- user
			}
		}()

		select {
		case user := <-userChan:
			_ = utils.WriteJSON(w, http.StatusOK, user)
		case err := <-errChan:
			utils.ErrorJSON(w, err)
		}
	}
}
