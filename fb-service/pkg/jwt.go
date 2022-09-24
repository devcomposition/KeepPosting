package fb_service

import (
	"github.com/golang-jwt/jwt"
	"time"
)

var jwtSecretKey = []byte("jwt_secret_key")

// CreateJWT function creates JWT while signing in and signing out
func CreateJWT(email string) (response string, err error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecretKey)
	if err == nil {
		return tokenString, nil
	}
	return "", err
}

// VerifyToken function verifies the JWT token while using APIs
func VerifyToken(tokenString string) (email string, err error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})
	if token != nil {
		return claims.Email, nil
	}
	return "", err
}
