package auth

import (
	"context"
	"errors"
	"net/http"
	"time"
	"zion-english/internal/conf"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

func Middleware(cfg *conf.Config, dbRO *queries.Queries, next http.Handler) http.Handler {
	const logtag = "[Auth Middleware]"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			unauthorized(w, r)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			logs.Log().Error(logtag, zap.Error(err))
			unauthorized(w, r)
			return
		}

		user := User{
			ID:    claims.UserID,
			Name:  claims.Name,
			Email: claims.Email,
			Role:  claims.Role,
		}

		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Login(w http.ResponseWriter, r *http.Request, cfg *conf.Config, dbRO *queries.Queries, email, password string) error {
	var user User

	switch {
	case email == cfg.SuperuserUsername && password == cfg.SuperuserPassword:
		user = User{
			Name:  "superuser",
			Email: "superuser",
			Role:  RoleSuperuser,
		}

	default:
		teacher, err := dbRO.GetTeacherByEmail(r.Context(), email)
		if err != nil {
			return errors.New("invalid credentials")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(teacher.Password), []byte(password)); err != nil {
			return errors.New("invalid credentials")
		}

		user = User{
			ID:    teacher.ID,
			Name:  teacher.Name,
			Email: teacher.Email,
			Role:  RoleTeacher,
		}
	}

	claims := Claims{
		UserID: user.ID,
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    tokenString,
		Path:     conf.Conf().BasePath,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	return nil
}

func UserFromRequest(r *http.Request, cfg *conf.Config) (User, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return User{}, false
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return User{}, false
	}

	return User{
		ID:    claims.UserID,
		Name:  claims.Name,
		Email: claims.Email,
		Role:  claims.Role,
	}, true
}

func Logout(w http.ResponseWriter) {
	logs.Log().Info("Triggered log out")
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     conf.Conf().BasePath,
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	})
}

func unauthorized(w http.ResponseWriter, r *http.Request) {
	url := utils.URL("/auth/login")
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
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

			unauthorized(w, r)
		}
	}
}
