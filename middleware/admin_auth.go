package middleware

import (
	"context"
	"net/http"

	"github.com/PauloHInocencio/testers-admin-dashboard/services/admin"
	"github.com/PauloHInocencio/testers-admin-dashboard/utils"
)

type contextKey string

const AdminEmailKey contextKey = "admin_email"

func RequireAdmin(store admin.AdminStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("admin_session")
			if err != nil || cookie.Value == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenHash := utils.HashToken(cookie.Value)
			email, err := store.FindValidSession(r.Context(), tokenHash)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AdminEmailKey, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
