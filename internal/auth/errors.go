package auth

import "errors"

var (
	ErrTeacherPendingApproval = errors.New("[AUTH] your account is pending approval. please wait for an administrator to approve your registration")
	ErrInvalidToken           = errors.New("[AUTH] invalid token")
	ErrInvalidRole            = errors.New("[AUTH] invalid role")
	ErrInvalidCredentials     = errors.New("[AUTH] invalid credentials")
)
