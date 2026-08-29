package constants

type IntegrationOAuthError string

const (
	IntegrationOAuthErrorInvalidState   IntegrationOAuthError = "invalid_state"
	IntegrationOAuthErrorMissingCode    IntegrationOAuthError = "missing_code"
	IntegrationOAuthErrorExchangeFailed IntegrationOAuthError = "exchange_failed"
	IntegrationOAuthErrorSaveFailed     IntegrationOAuthError = "save_failed"
)
