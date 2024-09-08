package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Logger)
	mux.Use(app.enableCORS)

	fs := http.FileServer(http.Dir("./template/docs/assets"))
	mux.Handle("/assets/*", http.StripPrefix("/assets/", fs))

	// Auth
	mux.Get("/auth/api/", app.Home)
	mux.Post("/auth/api/login", app.authenticate)
	mux.Post("/auth/api/refresh", app.refreshToken)
	mux.Post("/auth/api/register", app.RegisterNewUser)

	mux.Post("/auth/api/confirmation/{user_id}", app.ConfirmUser)
	mux.Get("/auth/api/confirmation/user/{user_id}", app.GetLastConfirmation)
	mux.Get("/auth/api/user/{user_email}", app.GetUserByEmail)

	mux.Post("/auth/api/send_password_email", app.SendForgottenPasswordEmail)
	mux.Post("/auth/api/user/password_reset/{user_id}", app.ChangePasswordByUserID)
	mux.Get("/auth/api/user/password_reset/token/{user_id}", app.FetchPasswordTokenByUserID)
	mux.Post("/auth/api/user/password_reset/token/use/{user_id}", app.SetTokenToUsed)

	// Protected routes
	mux.Route("/auth/api/logged_in", func(mux chi.Router) {
		mux.Use(app.authRequired)

		mux.Post("/logout", app.logout)

		// User data related routes

		mux.Get("/user/{user_email}", app.GetUserByEmail)
		mux.Get("/user/{user_id}", app.GetUserByID)
		mux.Get("/user/profile/{filename}", app.ServeStaticProfileImage)

		mux.Patch("/user/{user_id}", app.Updateuser)

		mux.Post("/user/password_reset/{user_id}", app.ChangePasswordByUserID)
		mux.Post("/user/send_password_email/{user_id}", app.SendPasswordResetEmail)
		mux.Post("/upload/profile", app.UploadProfileImage)
	})

	return mux
}
