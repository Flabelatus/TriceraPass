package application

import (
	"TriceraPass/cmd/api/auth"
	"TriceraPass/internal/repositories"
)

type Application struct {
	DSN          string
	Domain       string
	Repository   *repositories.GORMRepo
	Auth         auth.Auth
	JWTSecret    string
	JWTAudience  string
	JWTIssuer    string
	CookieDomain string
	// APIKey       string
}
