package server

import (
	"TriceraPass/cmd/api/application"
	"TriceraPass/cmd/api/server/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Routes sets up all the routes for the application using the chi router.
// It handles both public and protected routes, as well as middleware such as CORS and logging.
//
// Parameters:
// - app: A pointer to the Application struct, which holds shared configurations and services.
//
// Returns:
// - http.Handler: A configured router that handles all the registered routes.
func Routes(app *application.Application) http.Handler {
	mux := chi.NewRouter()

	// Middleware
	mux.Use(middleware.Recoverer) // Recovers from panics in the middleware chain
	mux.Use(middleware.Logger)    // Logs requests
	mux.Use(app.EnableCORS)       // Custom CORS middleware

	// Serve static assets for documentation
	fs := http.FileServer(http.Dir("./template/docs/assets"))
	mux.Handle("/assets/*", http.StripPrefix("/assets/", fs))

	// Authentication routes
	mux.Get("/auth/api/", handlers.Home(app))                     // Home page for the auth API
	mux.Post("/auth/api/login", handlers.Authenticate(app))       // Login route
	mux.Post("/auth/api/refresh", handlers.RefreshToken(app))     // Token refresh route
	mux.Post("/auth/api/register", handlers.RegisterNewUser(app)) // User registration route

	// Email confirmation routes
	mux.Post("/auth/api/confirmation/{user_id}", handlers.ConfirmUser(app))             // Confirm user by user ID
	mux.Get("/auth/api/confirmation/user/{user_id}", handlers.GetLastConfirmation(app)) // Get last confirmation for a user by user ID
	mux.Get("/auth/api/user/{user_email}", handlers.GetUserByEmail(app))                // Get user by email

	// Password reset routes
	mux.Post("/auth/api/send_password_email", handlers.SendForgottenPasswordEmail(app))                // Send password reset email
	mux.Post("/auth/api/user/password_reset/{user_id}", handlers.ChangePasswordByUserID(app))          // Reset password by user ID
	mux.Get("/auth/api/user/password_reset/token/{user_id}", handlers.FetchPasswordTokenByUserID(app)) // Fetch password reset token by user ID
	mux.Post("/auth/api/user/password_reset/token/use/{user_id}", handlers.SetTokenToUsed(app))        // Set the token to used after password reset

	// Protected routes (require authentication)
	mux.Route("/auth/api/logged_in", func(mux chi.Router) {
		mux.Use(app.AuthRequired) // Middleware to require authentication

		mux.Post("/logout", handlers.Logout(app)) // Logout route

		// User-related routes
		mux.Get("/user/{user_email}", handlers.GetUserByEmail(app))                // Get user by email
		mux.Get("/user/{user_id}", handlers.GetUserByID(app))                      // Get user by user ID
		mux.Get("/user/profile/{filename}", handlers.ServeStaticProfileImage(app)) // Serve static profile image

		mux.Patch("/user/{user_id}", handlers.Updateuser(app))                                // Update user by user ID
		mux.Post("/user/password_reset/{user_id}", handlers.ChangePasswordByUserID(app))      // Reset password by user ID
		mux.Post("/user/send_password_email/{user_id}", handlers.SendPasswordResetEmail(app)) // Send password reset email to user by user ID

		// Profile image upload
		mux.Post("/upload/profile", handlers.UploadProfileImage(app)) // Upload user profile image
	})

	return mux
}
