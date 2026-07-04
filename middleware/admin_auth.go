package middleware

import (
	"context"
	"net/http"
	
	"github.com/PauloHInocencio/testers-admin-dashboard/services/session"
	"github.com/PauloHInocencio/testers-admin-dashboard/utils"
)

type contextKey string

const AdminEmailKey contextKey = "admin_email"

func RequireAdmin(store session.SessionStore) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
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
			next(w, r.WithContext(ctx))
		}
	}
}
