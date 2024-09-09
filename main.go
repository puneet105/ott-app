package main

import (
	"log"
	"net/http"

	gohandler "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/puneet105/ott-app/internal/auth"
	"github.com/puneet105/ott-app/internal/handlers"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/login", auth.LoginHandler).Methods("POST")
	
	generateManifestHandler := http.HandlerFunc(handlers.GenerateManifestHandler)
	authGenMiddlewareHandler := auth.AuthMiddleware(generateManifestHandler)
	r.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
    	authGenMiddlewareHandler.ServeHTTP(w, r)
	}).Methods("GET")
	
	r.HandleFunc("/status/{taskID:[a-zA-Z0-9_-]+}", handlers.StatusHandler).Methods("GET")

	streamManifestHandler := http.HandlerFunc(handlers.StreamManifestHandler)
	authStreamMiddlewareHandler := auth.AuthMiddleware(streamManifestHandler)
	r.HandleFunc("/stream/{protocol}/{resolution}/manifest", func(w http.ResponseWriter, r *http.Request) {
		authStreamMiddlewareHandler.ServeHTTP(w, r)
	}).Methods("GET")

	corsHandler := gohandler.CORS(
        gohandler.AllowedOrigins([]string{"*"}), // Adjust this to specific origins if needed
        gohandler.AllowedHeaders([]string{"Authorization", "Content-Type"}),
        gohandler.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
    )(r)

	log.Println("Server is running on port 8080")
	if err := http.ListenAndServe(":8080", corsHandler); err != nil {
		log.Fatal(err)
	}
}
