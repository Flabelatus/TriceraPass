package main

import (
	"TriceraPass/internal/controllers"
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
func (a *application) Home(w http.ResponseWriter, r *http.Request) {

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
	conf, err := loadConfig(confFile)
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

		err = a.writeJSON(w, http.StatusOK, payload)
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

// Authentication Handlers
func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.Repository.GetUserByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid email or password"), http.StatusUnauthorized)
		return
	}

	// create a jwt user
	u := jwtUser{
		ID:        user.ID,
		UserName:  user.UserName,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	// generate tokens
	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)

	app.writeJSON(w, http.StatusAccepted, tokens)
}

func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			// parse the token to get the claims
			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})
			if err != nil {
				app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			// get the User ID from the token claims
			userID := claims.Subject
			// if err != nil {
			// 	app.errorJSON(w, fmt.Errorf("unknown user - %v", err), http.StatusUnauthorized)
			// }
			user, err := app.Repository.GetUserByID(userID)
			if err != nil {
				app.errorJSON(w, fmt.Errorf("unknown user - %v", err), http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				ID:        user.ID,
				UserName:  user.UserName,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}
			tokenPairs, err := app.auth.GenerateTokenPair(&u)
			if err != nil {
				app.errorJSON(w, errors.New("error generating tokens"), http.StatusUnauthorized)
				return
			}

			http.SetCookie(w, app.auth.GetRefreshCookie(tokenPairs.RefreshToken))

			app.writeJSON(w, http.StatusOK, tokenPairs)
		}
	}
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
	response := JSONResponse{
		Message: "successfully logged out",
	}
	_ = app.writeJSON(w, http.StatusOK, response)
}

// User Handlers
func (app *application) RegisterNewUser(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	newUser := models.User{ID: uuid.NewString(), CreatedAt: time.Now()}
	err = app.readJSON(w, r, &newUser)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	hashedPassword, err := controllers.HashAPassword(newUser.Password)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	newUser.Password = hashedPassword
	newUser.CreatedAt = time.Now()
	userID, err := app.Repository.CreateUser(&newUser)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
	response := JSONResponse{
		Error:   false,
		Message: "user successfully created",
		Data:    userID,
	}
	_ = app.writeJSON(w, http.StatusCreated, response)

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
		app.errorJSON(w, err)
		return
	}

	userMode := models.Mode{
		Name:   "1212",
		UserID: newUser.ID,
	}
	err = app.Repository.CreateMode(&userMode)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	// Send Mailgun here
	apiKey := os.Getenv("MAILGUN_API_KEY")
	domain := os.Getenv("MAILGUN_DOMAIN")

	_, err = controllers.SendEmail(domain, apiKey, newUser.Email, newUser.UserName, userID, "confirmation", 1)
	if err != nil {
		app.errorJSON(w, err)
		return
	}
}

func (app *application) GetUserByID(w http.ResponseWriter, r *http.Request) {
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
		_ = app.writeJSON(w, http.StatusOK, user)
	case err := <-errChan:
		app.errorJSON(w, err)
	}
}
