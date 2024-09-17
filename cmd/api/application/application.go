package application

import (
	"TriceraPass/cmd/api/auth"
	"TriceraPass/internal/repositories"
)

// Application holds the configuration and dependencies required by the API.
// It includes database connection details, authentication configurations, and repository access.
type Application struct {
	DSN          string                 // Data Source Name for database connection.
	Domain       string                 // Domain of the application.
	Repository   *repositories.GORMRepo // Pointer to the GORMRepo, which handles database interactions.
	Auth         auth.Auth              // Authentication handler for managing JWTs and user auth logic.
	JWTSecret    string                 // Secret key used for signing JWT tokens.
	JWTAudience  string                 // Audience claim for JWT tokens.
	JWTIssuer    string                 // Issuer claim for JWT tokens.
	CookieDomain string                 // Domain used for setting authentication cookies.
	// APIKey     string                // (Optional) API key for external services or further authentication.
}
