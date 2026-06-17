package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"time"
	"zion-english/internal/conf"
	"zion-english/internal/constants"
	"zion-english/internal/database/queries"
	"zion-english/internal/logs"
	"zion-english/internal/utils"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var loginLimiter = NewLoginLimiter(5, 15*time.Minute)

type Claims struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   Role   `json:"role"`
	jwt.RegisteredClaims
}

func LoginAllowed(ip string) bool {
	return loginLimiter.Allow(ip)
}

func RecordLoginFailure(ip string) {
	loginLimiter.RecordFailure(ip)
}

func ResetLoginFailures(ip string) {
	loginLimiter.Reset(ip)
}

var ErrTeacherPendingApproval = errors.New("your account is pending approval. please wait for an administrator to approve your registration")

func parseClaims(tokenString string, cfg *conf.Config) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Role != RoleTeacher && claims.Role != RoleSuperuser {
		return nil, errors.New("invalid role")
	}
	return claims, nil
}

func setSessionCookie(w http.ResponseWriter, cfg *conf.Config, value string, expires time.Time, maxAge int) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    value,
		Path:     cfg.BasePath,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expires,
	}
	if maxAge != 0 {
		cookie.MaxAge = maxAge
	}
	if cfg.IsProd() {
		cookie.Secure = true
	}
	http.SetCookie(w, cookie)
}

func Middleware(cfg *conf.Config, dbRO *queries.Queries, next http.Handler) http.Handler {
	const logtag = "[Auth Middleware]"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			invalidateSession(w, r)
			return
		}

		claims, err := parseClaims(cookie.Value, cfg)
		if err != nil {
			logs.Log().Error(logtag, zap.Error(err))
			invalidateSession(w, r)
			return
		}

		if claims.Role == RoleTeacher {
			if claims.UserID == 0 {
				logs.Log().Error(logtag, zap.String("reason", "teacher token missing user id"))
				invalidateSession(w, r)
				return
			}
			if _, err := dbRO.GetTeacherByID(r.Context(), claims.UserID); err != nil {
				logs.Log().Error(logtag, zap.Error(err), zap.Int64("teacher_id", claims.UserID))
				invalidateSession(w, r)
				return
			}
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

func Login(w http.ResponseWriter, r *http.Request, cfg *conf.Config, dbRO *queries.Queries, email, password string) (User, error) {
	var user User

	isSuperuser := subtle.ConstantTimeCompare([]byte(email), []byte(cfg.SuperuserUsername)) == 1 &&
		subtle.ConstantTimeCompare([]byte(password), []byte(cfg.SuperuserPassword)) == 1

	switch {
	case isSuperuser:
		user = User{
			Name:  "superuser",
			Email: cfg.SuperuserUsername,
			Role:  RoleSuperuser,
		}

	default:
		teacher, err := dbRO.GetTeacherByEmail(r.Context(), email)
		if err != nil {
			return User{}, errors.New("invalid credentials")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(teacher.Password), []byte(password)); err != nil {
			return User{}, errors.New("invalid credentials")
		}

		if teacher.Status != string(constants.TeacherStatusApproved) {
			return User{}, ErrTeacherPendingApproval
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
		return User{}, err
	}

	expires := time.Now().Add(24 * time.Hour)
	setSessionCookie(w, cfg, tokenString, expires, 0)

	return user, nil
}

func UserFromRequest(r *http.Request, cfg *conf.Config) (User, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return User{}, false
	}

	claims, err := parseClaims(cookie.Value, cfg)
	if err != nil {
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
	cfg := conf.Conf()
	setSessionCookie(w, cfg, "", time.Unix(0, 0), -1)
}

func ClearSession(w http.ResponseWriter) {
	cfg := conf.Conf()
	setSessionCookie(w, cfg, "", time.Unix(0, 0), -1)
}

func SessionUserValid(ctx context.Context, dbRO *queries.Queries, user User) bool {
	switch user.Role {
	case RoleSuperuser:
		return true
	case RoleTeacher:
		if user.ID == 0 {
			return false
		}
		_, err := dbRO.GetTeacherByID(ctx, user.ID)
		return err == nil
	default:
		return false
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	url := utils.URL("/auth/login")
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func invalidateSession(w http.ResponseWriter, r *http.Request) {
	ClearSession(w)
	redirectToLogin(w, r)
}

func accessDenied(w http.ResponseWriter, r *http.Request) {
	redirectToLogin(w, r)
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

			accessDenied(w, r)
		}
	}
}
