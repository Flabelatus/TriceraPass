package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type Auth struct {
	Issuer        string
	Audience      string
	Secret        string
	TokenExpiry   time.Duration
	RefreshExpiry time.Duration
	CookieDomain  string
	CookieName    string
	CookiePath    string
}

// var BlackList map[string]struct{}
var BlackList []string

type jwtUser struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	UserName  string `json:"username"`
	LastName  string `json:"last_name"`
}

type TokenPairs struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
}

func generateTokenID() string {
	return uuid.NewString()
}

// func (j *Auth) extractTokenID(token *jwt.Token) string {
// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok {
// 		fmt.Println("invalid claims while extracting tokenID claims")
// 		return ""
// 	}

// 	tokenID, ok := claims["jti"].(string)
// 	if !ok {
// 		fmt.Println("invalid token while extracting tokenID")
// 		return ""
// 	}

// 	return tokenID
// }

func (j *Auth) GenerateTokenPair(user *jwtUser) (TokenPairs, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenID := generateTokenID()

	// set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["jti"] = tokenID
	claims["name"] = fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	claims["sub"] = fmt.Sprint(user.ID)
	claims["aud"] = j.Audience
	claims["iss"] = j.Issuer
	claims["iat"] = time.Now().UTC().Unix()
	claims["typ"] = "JWT"

	claims["exp"] = time.Now().UTC().Add(j.TokenExpiry).Unix()

	// create a signed token
	signedAccessToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}

	// create refresh token and set claims
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshTokenClaims := refreshToken.Claims.(jwt.MapClaims)
	refreshTokenClaims["jti"] = tokenID
	refreshTokenClaims["sub"] = fmt.Sprint(user.ID)
	refreshTokenClaims["iat"] = time.Now().UTC().Unix()

	// set the expiry for the refresh token
	refreshTokenClaims["exp"] = time.Now().UTC().Add(j.RefreshExpiry).Unix()

	// create signed refresh token
	signedRefreshToken, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return TokenPairs{}, err
	}
	var tokenPairs = TokenPairs{
		Token:        signedAccessToken,
		RefreshToken: signedRefreshToken,
	}
	return tokenPairs, nil
}

// func (j *Auth) validateToken(tokenString string) (*jwt.Token, error) {
// 	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(j.Secret), nil
// 	})
// 	if err != nil || !token.Valid {
// 		return nil, fmt.Errorf("invalid token")
// 	}

// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if !ok {
// 		return nil, fmt.Errorf("invalid token claims")
// 	}

// 	tokenID := claims["jti"].(string)

// 	if j.isTokenBlackListed(tokenID) {
// 		return nil, fmt.Errorf("revoked token--")
// 	}

// 	return token, nil
// }

// func (j *Auth) isTokenBlackListed(tokenID string) bool {
// 	// if _, ok := BlackList[tokenID]; ok {
// 	// 	fmt.Println(BlackList[tokenID])
// 	// 	return true
// 	// } else {
// 	// 	return false
// 	// }
// 	for _, token := range BlackList {
// 		if token == tokenID {
// 			return true
// 		}
// 	}
// 	return false
// }

// func addToBlackList(tokenID string) {
// 	if BlackList == nil {
// 		BlackList = make([]string, 0)
// 	}
// 	// add the tokenID to the blacklist set
// 	BlackList = append(BlackList, tokenID)
// }

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
		return "", nil, errors.New("invalud auth header")
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
