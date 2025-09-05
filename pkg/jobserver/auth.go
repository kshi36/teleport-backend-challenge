package jobserver

import (
	"net/http"
	"strings"
	"teleport-jobworker/pkg/job"
)

const (
	UserRole  = "user"
	AdminRole = "admin"
)

// Pre-generated Bearer tokens are equivalent to userID, and provide a role: user, admin.
// In the future, tokens will be auto-generated (eg. JWT), and stored securely.
var validTokens = map[string]string{
	"user1":  UserRole,
	"user2":  UserRole,
	"admin1": AdminRole,
}

// bearerAuth inspects the Authorization: Bearer header and manages authentication.
func bearerAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			resJSON(w, ErrRes{"Unauthorized action"}, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		// simple lookup to static (token -> role) table
		role, ok := validTokens[token]
		if !ok {
			resJSON(w, ErrRes{"Unauthorized action"}, http.StatusUnauthorized)
			return
		}

		// store token (as userID) and role in context for handlers
		ctx := job.WithUserInfo(r.Context(), token, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
