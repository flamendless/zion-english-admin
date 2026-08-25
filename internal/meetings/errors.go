package meetings

import "errors"

var (
	ErrZoomNotConfigured          = errors.New("zoom integration is not configured")
	ErrZoomAuthorizeURLNotSet     = errors.New("zoom authorization url is not configured")
	ErrZoomNotConnected           = errors.New("connect zoom on your profile first")
	ErrProviderNotFound           = errors.New("meeting provider not found")
)
