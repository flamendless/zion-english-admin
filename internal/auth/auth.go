package auth

import (
	"encoding/base64"
	"net/http"
	"strings"
)

type Config struct {
	Username string `env:"ADMIN_USERNAME" required:"true"`
	Password string `env:"ADMIN_PASSWORD" required:"true"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) IsValid() bool {
	return c.Username != "" && c.Password != ""
}

func Middleware(cfg *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, hasAuth := r.BasicAuth()

		if !hasAuth || !validateCredentials(username, password, cfg) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("401 Unauthorized"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateCredentials(username, password string, cfg *Config) bool {
	return username == cfg.Username && password == cfg.Password
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
