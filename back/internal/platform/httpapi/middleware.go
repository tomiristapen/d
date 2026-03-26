package httpapi

import (
	"net/http"

	"back/internal/platform/authctx"
	"back/internal/platform/jwtutil"
	"back/internal/platform/httpx"
)

func authMiddleware(tokens *jwtutil.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := httpx.BearerToken(r.Header.Get("Authorization"))
			if token == "" {
				httpx.WriteAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
				return
			}

			claims, err := tokens.ParseAccessToken(token)
			if err != nil {
				httpx.WriteAPIError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
				return
			}

			next.ServeHTTP(w, r.WithContext(authctx.WithUserID(r.Context(), claims.UserID)))
		})
	}
}
