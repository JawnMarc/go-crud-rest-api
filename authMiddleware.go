package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		var tokenString string

		// prioritize the Authorization header and only check the cookie if the header is absent
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)

			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			} else {
				http.Error(w, "Unauthorized: Malformed Authorization header", http.StatusBadRequest)
				return
			}
		} else if cookie, err := r.Cookie("token"); err == nil {
			tokenString = cookie.Value
		} else {
			http.Error(w, "Unathorized: No token present", http.StatusProxyAuthRequired)
		}

		// parse jwt token
		claims := &Claims{}
		tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil // Will replace jwtKey with proper key management in production.
		})

		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorMalformed != 0 {
					http.Error(w, "Forbidden: Token expired", http.StatusForbidden) // 403
					return
				} else {
					http.Error(w, fmt.Sprintf("Forbidden: Invalid token: %s", err), http.StatusForbidden)
					return
				}
			}

			http.Error(w, fmt.Sprintf("Forbidden: Invalid token: %s", err), http.StatusForbidden)
			return
		}

		if !tkn.Valid {
			http.Error(w, "Forbidden: Invalid token", http.StatusForbidden) // 403
			return
		}

		if claims.Username == "" {
			http.Error(w, "Forbidden: Token missing username claim", http.StatusForbidden)
			return
		}

		// token is invalid. Store username in the context
		ctx := context.WithValue(r.Context(), "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
