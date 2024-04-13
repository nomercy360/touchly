package transport

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

func parseToken(authorizationHeader string) (string, error) {
	fields := strings.Fields(authorizationHeader)
	if len(fields) == 2 && strings.EqualFold(fields[0], "bearer") {
		return fields[1], nil
	}

	return "", fmt.Errorf("not a bearer authorization")
}

func WithAuth(jwtSecret string) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/admin/auth" {
				handler.ServeHTTP(w, r)
				return
			}

			auth, err := parseToken(r.Header.Get("Authorization"))
			if err != nil {
				_ = WriteError(w, http.StatusUnauthorized, err.Error())
				return
			}

			token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil {
				_ = WriteError(w, http.StatusUnauthorized, err.Error())
				return
			}

			userID := token.Claims.(jwt.MapClaims)["userID"].(float64)

			ctx := context.WithValue(r.Context(), "userID", int64(userID))

			handler.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
