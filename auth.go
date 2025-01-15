package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// login hander
func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			http.Error(w, "Invalid login request", http.StatusBadRequest)
			return
		}

		// Query database for the user
		row := db.QueryRow("SELECT password FROM users WHERE username = $1", user.Username)

		var hashedPassword string
		err = row.Scan(&hashedPassword)

		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusBadRequest)
			return
		} else if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// compare provided password with hashed password
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// generate JWT token
		expirationTime := time.Now().Add(24 * time.Hour) // token expires in 24HR

		claims := &Claims{
			Username: user.Username,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		tokenString, err := token.SignedString(jwtKey)

		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		// set jwt token as cookie, add secure flag
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tokenString,
			Expires:  expirationTime,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode, // or http.SameSiteLaxMode
			Path:     "/",
		})

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}
