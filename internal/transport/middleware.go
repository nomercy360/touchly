package transport

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"touchly/internal/terrors"
)

func parseToken(authorizationHeader string) (string, error) {
	fields := strings.Fields(authorizationHeader)
	if len(fields) == 2 && strings.EqualFold(fields[0], "bearer") {
		return fields[1], nil
	}

	return "", fmt.Errorf("not a bearer authorization")
}

func WithAuth(jwtSecret string, required bool) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				if required {
					WriteError(r, w, terrors.Unauthorized(nil, "authorization header required"))
					return
				}
				handler.ServeHTTP(w, r)
				return
			}

			auth, err := parseToken(authHeader)
			if err != nil {
				WriteError(r, w, terrors.Unauthorized(err, "invalid authorization header"))
				return
			}

			token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil {
				WriteError(r, w, terrors.Unauthorized(err, "invalid token"))
				return
			}

			userID := token.Claims.(jwt.MapClaims)["userID"].(float64)

			ctx := context.WithValue(r.Context(), "userID", int64(userID))

			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func WithAdminAuth(secret string) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("X-Api-Key")

			if auth != secret {
				WriteError(r, w, terrors.Unauthorized(nil, "invalid token"))
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}
