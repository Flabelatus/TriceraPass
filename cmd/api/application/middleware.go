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

func (app *Application) EnableCORS(h http.Handler) http.Handler {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Printf("Can not locate the env file: %v", err))
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
			// w.Header().Set("Access-Control-Allow-Origin", origin)
			// w.Header().Set("Access-Control-Allow-Credentials", "true")
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

// func (app *application) adminRequired(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		_, claims, err := app.auth.GetTokenFromHeaderAndVerify(w, r)
// 		if err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}

// 		isAdmin, err := app.isUserAdmin(claims.Subject)
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		if !isAdmin {
// 			w.WriteHeader(http.StatusForbidden)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// // Helper function to check if the user is an admin
// func (app *application) isUserAdmin(userID string) (bool, error) {
// 	IntegerUserID, err := strconv.Atoi(userID)
// 	if err != nil {
// 		return false, err
// 	}
// 	user, err := app.Repository.GetUserByID(uint(IntegerUserID))
// 	if err != nil {
// 		return false, err
// 	}
// 	return user.Mode.Name == "admin", nil
// }
