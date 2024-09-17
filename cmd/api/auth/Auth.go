// Package auth provides functionalities for handling authentication and token management.
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Auth handles the configuration needed for authentication.
type Auth struct {
	Issuer        string        // The issuer of the token, typically your application name.
	Audience      string        // The audience of the token, typically your application or client name.
	Secret        string        // The secret key used to sign JWTs.
	TokenExpiry   time.Duration // Duration for which the access token is valid.
	RefreshExpiry time.Duration // Duration for which the refresh token is valid.
	CookieDomain  string        // Domain for setting the refresh token cookie.
	CookieName    string        // Name of the refresh token cookie.
	CookiePath    string        // Path for setting the refresh token cookie.
}

// JwtUser represents a user and their associated JWT claims.
type JwtUser struct {
	ID        string `json:"id"`         // User ID.
	FirstName string `json:"first_name"` // User's first name.
	UserName  string `json:"username"`   // User's username.
	LastName  string `json:"last_name"`  // User's last name.
}

// TokenPairs represents the access and refresh tokens.
type TokenPairs struct {
	Token        string `json:"access_token"`  // JWT access token.
	RefreshToken string `json:"refresh_token"` // JWT refresh token.
}

// Claims represents the JWT claims for the user.
type Claims struct {
	jwt.RegisteredClaims // Standard JWT registered claims (e.g., iat, exp, etc.).
}

// generateTokenID generates a new unique token ID.
//
// Returns:
// - string: A UUID string representing the token ID.
func generateTokenID() string {
	return uuid.NewString()
}

// GenerateTokenPair generates an access and refresh token pair for the given user.
// The tokens are signed using the secret key from the Auth configuration.
//
// Parameters:
// - user: A pointer to the JwtUser containing user information for the token claims.
//
// Returns:
// - TokenPairs: A struct containing the signed access and refresh tokens.
// - error: An error if the tokens fail to be generated.
func (j *Auth) GenerateTokenPair(user *JwtUser) (TokenPairs, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenID := generateTokenID()

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["jti"] = tokenID
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"
	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	// Create a signed access token
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// Create refresh token and set claims
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["jti"] = tokenID
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()
	refreshTokenClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()

	// Create signed refresh token
	signedRefreshToken, err := refreshToken.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// Return the token pairs
	return TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}, nil
}

// GetRefreshCookie returns an HTTP cookie for the refresh token with the specified configurations.
//
// Parameters:
// - refreshToken: The refresh token to set in the cookie.
//
// Returns:
// - *http.Cookie: A pointer to the HTTP cookie containing the refresh token.
func (j *Auth) GetRefreshCookie(refreshToken string) *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    refreshToken,
		Expires:  time.Now().Add(j.RefreshExpiry),
		MaxAge:   int(j.RefreshExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

// GetExpiredRefreshCookie returns an HTTP cookie that is immediately expired, to effectively log out a user.
//
// Returns:
// - *http.Cookie: A pointer to the expired HTTP cookie.
func (j *Auth) GetExpiredRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:     j.CookieName,
		Path:     j.CookiePath,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
		Domain:   j.CookieDomain,
		HttpOnly: true,
		Secure:   true,
	}
}

// GetTokenFromHeaderAndVerify retrieves a token from the Authorization header and verifies its validity.
// It checks for proper token structure, signature, and claims, such as issuer and expiration.
//
// Parameters:
// - w: The HTTP response writer to modify headers.
// - r: The HTTP request containing the Authorization header.
//
// Returns:
// - string: The token if valid.
// - *Claims: A pointer to the Claims struct containing the token claims.
// - error: An error if the token is invalid or expired.
func (j *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	w.Header().Add("Vary", "Authorization")

	// get the auth header
	authHeader := r.Header.Get("Authorization")

	// sanity check
	if authHeader == "" {
		return "", nil, errors.New("no auth header")
	}

	// split the header on spaces
	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 {
		return "", nil, errors.New("invalid auth header")
	}

	if headerParts[0] != "Bearer" {
		return "", nil, errors.New("invalid auth header")
	}

	token := headerParts[1]

	// declare an empty claims
	claims := &Claims{}

	// parse the token
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.Secret), nil
	})

	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}
		return "", nil, err
	}

	if claims.Issuer != j.Issuer {
		return "", nil, errors.New("invalid issuer")
	}

	return token, claims, nil
}
