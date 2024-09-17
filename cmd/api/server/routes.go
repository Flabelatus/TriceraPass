package server

import (
	"TriceraPass/cmd/api/server/handlers"
	"TriceraPass/cmd/api/application"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Routes(app *application.Application) http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Logger)
	mux.Use(app.EnableCORS)

	fs := http.FileServer(http.Dir("./template/docs/assets"))
	mux.Handle("/assets/*", http.StripPrefix("/assets/", fs))

	// Auth
	mux.Get("/auth/api/", handlers.Home(app))
	mux.Post("/auth/api/login", handlers.Authenticate(app))
	mux.Post("/auth/api/refresh", handlers.RefreshToken(app))
	mux.Post("/auth/api/register", handlers.RegisterNewUser(app))

	mux.Post("/auth/api/confirmation/{user_id}", handlers.ConfirmUser(app))
	mux.Get("/auth/api/confirmation/user/{user_id}", handlers.GetLastConfirmation(app))
	mux.Get("/auth/api/user/{user_email}", handlers.GetUserByEmail(app))

	mux.Post("/auth/api/send_password_email", handlers.SendForgottenPasswordEmail(app))
	mux.Post("/auth/api/user/password_reset/{user_id}", handlers.ChangePasswordByUserID(app))
	mux.Get("/auth/api/user/password_reset/token/{user_id}", handlers.FetchPasswordTokenByUserID(app))
	mux.Post("/auth/api/user/password_reset/token/use/{user_id}", handlers.SetTokenToUsed(app))

	// Protected routes
	mux.Route("/auth/api/logged_in", func(mux chi.Router) {
		mux.Use(app.AuthRequired)

		mux.Post("/logout", handlers.Logout(app))

		// User data related routes

		mux.Get("/user/{user_email}", handlers.GetUserByEmail(app))
		mux.Get("/user/{user_id}", handlers.GetUserByID(app))
		mux.Get("/user/profile/{filename}", handlers.ServeStaticProfileImage(app))

		mux.Patch("/user/{user_id}", handlers.Updateuser(app))

		mux.Post("/user/password_reset/{user_id}", handlers.ChangePasswordByUserID(app))
		mux.Post("/user/send_password_email/{user_id}", handlers.SendPasswordResetEmail(app))
		mux.Post("/upload/profile", handlers.UploadProfileImage(app))
	})

	return mux
}
