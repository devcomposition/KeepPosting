package fb_service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func GetRandomOAuthStateString(r *http.Request) string {
	var s = r.FormValue("state")
	if len(s) > 0 {
		return s
	}
	n := 5
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "randpackagefailedsoheresarandomstring"
	}
	s = fmt.Sprintf("%X", b)
	return s
}

// GetUserInfoFromFacebook returns information of user fetched from Facebook
func GetUserInfoFromFacebook(token string) (FacebookUserDetails, error) {
	var fbUserDetails FacebookUserDetails
	res, err := http.Get("https://graph.facebook.com/me?fields=id,name,email&access_token=" + token)

	if err != nil {
		return FacebookUserDetails{}, errors.New("error occurred while getting info from Facebook")
	}

	decoder := json.NewDecoder(res.Body)
	decoderErr := decoder.Decode(&fbUserDetails)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	if decoderErr != nil {
		return FacebookUserDetails{}, errors.New("error occurred while decoding information received from Facebook")
	}

	return fbUserDetails, nil
}
