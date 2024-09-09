// internal/handlers/handlers.go
package handlers

import (
	
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"strings"
	"github.com/gorilla/mux"
	"github.com/puneet105/ott-app/internal/auth"
	"github.com/puneet105/ott-app/internal/manifest"
	"github.com/google/uuid"
	"sync"
)

var outputDir string = "./manifests-final"
var (
	manifestStatus = make(map[string]string)
	statusMutex    = sync.Mutex{}
)

func GenerateManifestHandler(w http.ResponseWriter, r *http.Request) {
	
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		fmt.Println("Authorization header missing")
		return
	}
	
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		fmt.Println("Invalid Authorization header format")
		return
	}

	
	_, err := auth.ValidateJWT(tokenString)
	if err != nil {
		fmt.Println("Unauthorized: "+err.Error())
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	id := uuid.New()
	taskID:= id.String() // Replace with actual logic to generate unique ID
	fmt.Println("Generated UUID:",taskID)

	statusMutex.Lock()
	manifestStatus[taskID] = "InProgress...!!!"
	statusMutex.Unlock()

	inputFile := r.URL.Query().Get("input")
	
	go func(taskID string) {
		defer func() {
			statusMutex.Lock()
			manifestStatus[taskID] = "completed"
			statusMutex.Unlock()
		}()

		err = manifest.GenerateManifest(inputFile, outputDir)
		if err != nil {
			// Handle error and update status
			statusMutex.Lock()
			manifestStatus[taskID] = fmt.Sprintf("failed: %v", err)
			statusMutex.Unlock()
		}
	}(taskID)
	response := map[string]string{
		"status": "Manifest generation started",
		"taskID": taskID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	fmt.Println(response)
	w.Write([]byte(fmt.Sprintf("status: Manifest generation started\n taskID: %s", taskID)))
}

func StreamManifestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
        w.WriteHeader(http.StatusOK)
        return
    }

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header missing", http.StatusUnauthorized)
		fmt.Println("Authorization header missing")
		return
	}
	
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		fmt.Println("Invalid Authorization header format")
		return
	}

	
	_, err := auth.ValidateJWT(tokenString)
	if err != nil {
		fmt.Println("Unauthorized: "+err.Error())
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	protocol := vars["protocol"]
	resolution := vars["resolution"]

	var manifestPath string
	if protocol == "hls" {
		manifestPath = filepath.Join("/Users/puneetsharma/go/src/github.com/puneet105/ott-app/",outputDir, resolution, "hls", "manifest.m3u8")
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	} else if protocol == "dash" {
		manifestPath = filepath.Join("/Users/puneetsharma/go/src/github.com/puneet105/ott-app/",outputDir, resolution, "dash", "manifest.mpd")
		w.Header().Set("Content-Type", "application/dash+xml")
	} else {
		http.Error(w, "Invalid protocol", http.StatusBadRequest)
		return
	}

	fmt.Println("Manifest served successfully")
	http.ServeFile(w, r, manifestPath)
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		http.Error(w, "Manifest not found", http.StatusNotFound)
		fmt.Println("Manifest not found")
		return
	}
	w.Write(manifestData)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	taskID := mux.Vars(r)["taskID"]

	statusMutex.Lock()
	status, exists := manifestStatus[taskID]
	statusMutex.Unlock()

	if !exists {
		http.Error(w, "Task ID not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"taskID":"%s","status":"%s"}`, taskID, status)
}
