package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"
	"zion-english/internal/conf"
)

func Middleware(cfg *conf.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			unauthorized(w)
			return
		}

		var role Role

		switch {
		case username == cfg.AdminTeacherUsername && password == cfg.AdminTeacherPassword:
			role = RoleTeacher

		case username == cfg.SuperuserUsername && password == cfg.SuperuserPassword:
			role = RoleSuperuser

		default:
			unauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), roleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func ExtractBasicAuth(r *http.Request) (username, password string, ok bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", "", false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}

	credentials := string(decoded)
	colonIndex := strings.Index(credentials, ":")
	if colonIndex == -1 {
		return "", "", false
	}

	return credentials[:colonIndex], credentials[colonIndex+1:], true
}

func RequireRole(allowed ...Role) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			role := GetRole(r.Context())

			for _, a := range allowed {
				if role == a {
					next(w, r)
					return
				}
			}

			unauthorized(w)
		}
	}
}
