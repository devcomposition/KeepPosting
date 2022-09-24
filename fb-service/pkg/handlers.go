package fb_service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"log"
	"net/http"
	"os"
	"strings"
)

var facebookOauth2Config *oauth2.Config

// GetFacebookOAuthConfig returns the config required to call facebook login
func init() {
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Fatal(err)
		return
	}
	facebookOauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURL:  os.Getenv("FACEBOOK_REDIRECT_URL"),
		Endpoint:     facebook.Endpoint,
		Scopes:       []string{"email"},
	}
}

func RenderHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/home.html")
}

func RenderProfile(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/profile.html")
}

func InitFacebookLogin(w http.ResponseWriter, r *http.Request) {
	url := facebookOauth2Config.AuthCodeURL(GetRandomOAuthStateString(r))
	log.Println(url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func HandleFacebookLoginCallback(w http.ResponseWriter, r *http.Request) {
	var state = r.FormValue("state")
	var code = r.FormValue("code")

	if state != GetRandomOAuthStateString(r) {
		log.Printf("invalid oauth state")
		http.Redirect(w, r, "/?invalidlogin=true", http.StatusTemporaryRedirect)
	}

	// changed oauth2.NoContext to context.Background() as its deprecated. Check this if we get any error while running the code
	token, err := facebookOauth2Config.Exchange(context.Background(), code)
	if err != nil || token == nil {
		log.Printf("code exchange failed: %s", err.Error())
		http.Redirect(w, r, "/?invalidlogin=true", http.StatusTemporaryRedirect)
	}

	fbUserDetails, fbUserDetailsError := GetUserInfoFromFacebook(token.AccessToken)
	if fbUserDetailsError != nil {
		http.Redirect(w, r, "/?invalidlogin=true", http.StatusTemporaryRedirect)
	}

	authToken, authTokenError := SignInUser(fbUserDetails)
	if authTokenError != nil {
		log.Printf("authtokenerror encountered: %s\n", authTokenError)
		http.Redirect(w, r, "/?invalidlogin=true", http.StatusTemporaryRedirect)
	}

	cookie := http.Cookie{Name: "Authorization", Value: "Bearer " + authToken, Path: "/"}
	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/profile", http.StatusTemporaryRedirect)
}

func SignInUser(facebookUserDetails FacebookUserDetails) (string, error) {
	var result UserDetails

	if facebookUserDetails == (FacebookUserDetails{}) {
		return "", errors.New("user details cannot be empty")
	}

	if facebookUserDetails.Email == "" {
		return "", errors.New("email cannot be empty")
	}

	if facebookUserDetails.Name == "" {
		return "", errors.New("name cannot be empty")
	}

	mysqlPwd := os.Getenv("MYSQL_PWD")
	// get a handle to the database
	DB, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(localhost:3306)/users", mysqlPwd))
	if err != nil {
		log.Fatalf("Unable to connect to database: %s", err.Error())
	}
	// defer DB.Close()
	defer func(DB *sql.DB) {
		err := DB.Close()
		if err != nil {
			return
		}
	}(DB)
	rows, err := DB.Query("select name, email from users where email = ?", facebookUserDetails.Email)
	if err != nil {
		log.Fatalf("Problem fetching results from db: %s", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	for rows.Next() {
		err := rows.Scan(&result.Name, &result.Email)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	if result == (UserDetails{}) {
		log.Println("inserting data into users table")
		res, err := DB.Exec("insert into users (Name, Email, Password) values(?, ?, NULL)", facebookUserDetails.Name, facebookUserDetails.Email)
		if err != nil {
			return "", err
		}
		_, err = res.LastInsertId()
		if err != nil {
			return "", errors.New("error occurred while fetching lastid")
		}
		resra, err := res.RowsAffected()
		if err != nil {
			return "", errors.New("error occurred while fetching rows affected count")
		}
		log.Printf("successfully inserted data into users table: %s \n", resra)
	}

	tokenString, err := CreateJWT(facebookUserDetails.Email)
	if tokenString == "" {
		return "", err
	}

	return tokenString, nil
}

func GetUserDetails(w http.ResponseWriter, r *http.Request) {
	var result UserDetails
	var errorResponse = ErrorResponse{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
	}

	bearerToken := r.Header.Get("Authorization")
	var authorizationToken = strings.Split(bearerToken, " ")[1]

	email, _ := VerifyToken(authorizationToken)
	if email == "" {
		returnErrorResponse(w, r, errorResponse)
	} else {
		mysqlPwd := os.Getenv("MYSQL_PWD")
		// get a handle to the database
		DB, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(localhost:3306)/users", mysqlPwd))
		if err != nil {
			log.Fatalf("Unable to connect to database: %s", err.Error())
		}
		// defer DB.Close()
		defer func(DB *sql.DB) {
			err := DB.Close()
			if err != nil {
				return
			}
		}(DB)

		rows, err := DB.Query("select name, email from users where email = ?", email)
		if err != nil {
			log.Fatal(err)
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				return
			}
		}(rows)

		for rows.Next() {
			err := rows.Scan(&result.Name, &result.Email)
			if err != nil {
				log.Fatal(err)
			}
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		} else {
			var successResponse = SuccessResponse{
				Code:     http.StatusOK,
				Message:  "You are logged in successfully",
				Response: result.Name,
			}

			successJsonResponse, jsonError := json.Marshal(successResponse)

			if jsonError != nil {
				returnErrorResponse(w, r, errorResponse)
			}
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write(successJsonResponse)
			if err != nil {
				return
			}
		}
	}
}

func returnErrorResponse(w http.ResponseWriter, r *http.Request, errorMessage ErrorResponse) {
	httpResponse := &ErrorResponse{
		Code:    errorMessage.Code,
		Message: errorMessage.Message,
	}
	jsonResponse, err := json.Marshal(httpResponse)
	if err != nil {
		log.Fatalf("error while fetching the jsonResponse: %s", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorMessage.Code)
	_, err = w.Write(jsonResponse)
	if err != nil {
		return
	}
}
