package auth

import (
	"context"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

type contextKey string
const FirebaseUIDKey contextKey = "firebase_uid"

type AuthMiddleware struct {
	App *firebase.App
}

func NewAuthMiddleware(credPath string) *AuthMiddleware {
	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic(err)
	}
	return &AuthMiddleware{App: app}
}

func (m *AuthMiddleware) Secure(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		client, err := m.App.Auth(r.Context())
		if err != nil {
			http.Error(w, "Internal auth configuration error", http.StatusInternalServerError)
			return
		}

		token, err := client.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Inject the authenticated UID into context
		ctx := context.WithValue(r.Context(), FirebaseUIDKey, token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}