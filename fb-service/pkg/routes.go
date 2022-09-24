package fb_service

import (
	"github.com/gorilla/mux"
	"log"
)

func AddRoutes(route *mux.Router) {
	log.Println("Loading routes...")

	route.HandleFunc("/", RenderHome)
	route.HandleFunc("/profile", RenderProfile)
	route.HandleFunc("/facebook", InitFacebookLogin)
	route.HandleFunc("/facebook/callback", HandleFacebookLoginCallback)
	route.HandleFunc("/userDetails", GetUserDetails).Methods("GET")

	log.Println("Loaded routes!")
}
