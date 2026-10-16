package sqlconsole

import "errors"

var (
	ErrSQLQueryEmpty            = errors.New("[SQL CONSOLE] query is required")
	ErrSQLQueryTooLong          = errors.New("[SQL CONSOLE] query is too long")
	ErrSQLMultipleStatements    = errors.New("[SQL CONSOLE] only one SQL statement is allowed")
	ErrSQLNotSelect             = errors.New("[SQL CONSOLE] only SELECT or WITH queries are allowed")
	ErrSQLForbiddenKeyword      = errors.New("[SQL CONSOLE] query contains a forbidden keyword")
	ErrSQLExecution             = errors.New("[SQL CONSOLE] query execution failed")
	ErrSQLSavedNameRequired     = errors.New("[SQL CONSOLE] saved query name is required")
	ErrSQLSavedNameTooLong      = errors.New("[SQL CONSOLE] saved query name is too long")
	ErrSQLSavedQueryNotFound    = errors.New("[SQL CONSOLE] saved query not found")
	ErrSQLSavedInvalidAction    = errors.New("[SQL CONSOLE] invalid saved query action")
	ErrSQLSavedNameDuplicate    = errors.New("[SQL CONSOLE] a saved query with this name already exists")
)
