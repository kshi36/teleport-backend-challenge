package jobserver

import (
	"net/http"
	"strings"
	"teleport-jobworker/pkg/job"
)

// Pre-generated Bearer tokens are equivalent to userID, and provide a role: user, admin.
// In the future, tokens will be auto-generated (eg. JWT), and stored securely.
var validTokens = map[string]string{
	"user1":  job.User,
	"user2":  job.User,
	"admin1": job.Admin,
}

// bearerAuth inspects the Authorization: Bearer header and manages authentication.
func bearerAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		authHeaderFields := strings.Fields(authHeader)
		if len(authHeaderFields) != 2 || !strings.EqualFold(authHeaderFields[0], "Bearer") {
			responseJSON(w, ErrorResponse{"Unauthorized action"}, http.StatusUnauthorized)
			return
		}

		token := authHeaderFields[1]
		role, ok := validTokens[token]
		if !ok {
			responseJSON(w, ErrorResponse{"Unauthorized action"}, http.StatusUnauthorized)
			return
		}

		// store token (as userID) and role in context for use in Manager library calls
		ctx := job.WithUserInfo(r.Context(), token, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
