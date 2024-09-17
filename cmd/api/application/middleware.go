package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type contextKey string

const userContextKey contextKey = "userID"

// EnableCORS is a middleware function that enables Cross-Origin Resource Sharing (CORS).
// It loads allowed origins from the environment file and applies the appropriate headers
// for incoming requests. The function also handles preflight (OPTIONS) requests.
//
// Parameters:
// - h: The next HTTP handler to call after processing the CORS headers.
//
// Returns:
// - http.Handler: The middleware handler that processes CORS and calls the next handler.
func (app *Application) EnableCORS(h http.Handler) http.Handler {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Printf("Cannot locate the env file: %v", err))
	}

	allowedOrigins := os.Getenv("CORS")

	if allowedOrigins == "" {
		log.Fatal("No allowed origins found in the environment")
	}

	allowOriginList := strings.Split(allowedOrigins, ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isOriginAllowed(origin, allowOriginList) {
			for _, url := range allowOriginList {
				fmt.Println(url)
				w.Header().Set("Access-Control-Allow-Origin", url)
			}
		}

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-Auth-Email, X-Auth-Key, X-CSRF-Token, Origin, X-Requested-With, Authorization")
			return
		} else {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			h.ServeHTTP(w, r)
		}
	})
}

// AuthRequired is a middleware function that checks if a request is authenticated.
// It verifies the JWT token from the Authorization header, extracts the user ID,
// and stores it in the request context for further processing.
//
// Parameters:
// - next: The next HTTP handler to call after authentication succeeds.
//
// Returns:
// - http.Handler: The middleware handler that checks authentication and calls the next handler.
func (app *Application) AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _, err := app.Auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Store the user ID in the request context
		ctx := context.WithValue(r.Context(), userContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminRequired is a middleware function that ensures the user has admin privileges.
// It verifies the JWT token and checks if the user has admin permissions by querying
// the user's role in the database. If the user is not an admin, it returns a 403 Forbidden status.
//
// Parameters:
// - next: The next HTTP handler to call after admin authorization succeeds.
//
// Returns:
// - http.Handler: The middleware handler that checks for admin privileges and calls the next handler.
func (app *Application) AdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := app.Auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		isAdmin, err := app.IsUserAdmin(claims.Subject)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !isAdmin {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// IsUserAdmin checks if a user has admin privileges based on their user ID.
// It retrieves the user's information from the database and checks their role.
//
// Parameters:
// - userID: The ID of the user whose role is being checked.
//
// Returns:
// - bool: True if the user is an admin, false otherwise.
// - error: An error if the user could not be retrieved from the database.
func (app *Application) IsUserAdmin(userID string) (bool, error) {

	user, err := app.Repository.GetUserByID(userID)
	if err != nil {
		return false, err
	}
	return user.Mode.Name == "admin", nil
}
