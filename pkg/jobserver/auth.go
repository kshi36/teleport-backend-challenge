package jobserver

import (
	"net/http"
	"strings"
	"teleport-jobworker/pkg/job"
)

// For the prototype, the server will contain a static store of valid Bearer tokens.
// In the future, tokens will be auto-generated (eg. JWT), and stored securely.
var validTokens = map[string]tokenClaims{
	"user1_token":  {"user1", job.User},
	"user2_token":  {"user2", job.User},
	"admin1_token": {"admin1", job.Admin},
}

// tokenClaims contains user information to be used in job library calls.
type tokenClaims struct {
	userId string
	role   string
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
		claims, ok := validTokens[token]
		if !ok {
			responseJSON(w, ErrorResponse{"Unauthorized action"}, http.StatusUnauthorized)
			return
		}

		// store token (as userID) and role in context for use in Manager library calls
		ctx := job.WithUserInfo(r.Context(), claims.userId, claims.role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
