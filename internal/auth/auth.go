package auth

import (
	"errors"
	"fmt"
	"time"
	"net/http"
	"github.com/golang-jwt/jwt/v5"
	"github.com/puneet105/ott-app/internal/models"
	"encoding/json"
)

var jwtKey = []byte("YOU_LIVE_ONLY_ONCE_PUNEET")

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}
var users = map[string]string{"puneet": "puneet123", "devops": "devops123"}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		fmt.Println("Invalid request payload")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	for uname,pass := range users{
		if user.Username == "" || user.Password == "" {
			fmt.Println("Invalid credentials")
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}else if user.Username == uname && user.Password == pass{
			fmt.Printf(" User exists, Generating Token ...!!!\n")
			w.Write([]byte(" User exists, Generating Token ...!!!\n"))
			break
		}else{
			fmt.Println("User Does Not Exists")
			http.Error(w, "User Does Not Exists", http.StatusNotFound)
			return
		}
	}

	token, err := GenerateJWT(user.Username)
	if err != nil {
		fmt.Println("Error generating token")
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
    w.Write([]byte(token))
}

func GenerateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
