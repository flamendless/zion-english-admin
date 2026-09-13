package meetings

import "errors"

var (
	ErrZoomNotConfigured      = errors.New("[ZOOM] zoom integration is not configured")
	ErrZoomAuthorizeURLNotSet = errors.New("[ZOOM] zoom authorization url is not configured")
	ErrZoomNotConnected       = errors.New("[ZOOM] connect zoom on your profile first")
	ErrProviderNotFound       = errors.New("[MEETINGS] meeting provider not found")
	ErrInvalidEncryptedToken  = errors.New("[MEETINGS] invalid encrypted token")
)
