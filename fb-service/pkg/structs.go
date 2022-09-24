package fb_service

import "github.com/golang-jwt/jwt"

// ErrorResponse is a struct for sending error messages with code
type ErrorResponse struct {
	Code    int
	Message string
}

// SuccessResponse is a struct for sending success messages with code
type SuccessResponse struct {
	Code     int
	Message  string
	Response interface{}
}

// Claims is a struct that will be encoded to a JWT.
// jwt.StandardClaims is an embedded type to provide expiry time
type Claims struct {
	Email string
	jwt.StandardClaims
}

// UserDetails is a struct used to store user details
type UserDetails struct {
	Name     string
	Email    string
	Password string
}

// FacebookUserDetails is a struct to store user details with the facebook ID associated with the particular user
type FacebookUserDetails struct {
	ID    string
	Name  string
	Email string
}
