package sqlconsole

import (
	"errors"
	"testing"
)

func TestValidateReadOnlyQuery(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantErr error
	}{
		{name: "simple select", sql: "SELECT 1", wantErr: nil},
		{name: "with cte", sql: "WITH t AS (SELECT 1 AS x) SELECT * FROM t", wantErr: nil},
		{name: "trailing semicolon", sql: "SELECT 1;", wantErr: nil},
		{name: "empty", sql: "  ", wantErr: ErrSQLQueryEmpty},
		{name: "multiple statements", sql: "SELECT 1; DELETE FROM tbl_students", wantErr: ErrSQLMultipleStatements},
		{name: "update", sql: "UPDATE tbl_students SET name = 'x'", wantErr: ErrSQLNotSelect},
		{name: "literal string ok", sql: "SELECT 'hello' AS label", wantErr: nil},
		{name: "explain", sql: "EXPLAIN SELECT 1", wantErr: ErrSQLNotSelect},
		{name: "insert keyword", sql: "SELECT 1 WHERE INSERT = 1", wantErr: ErrSQLForbiddenKeyword},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateReadOnlyQuery(tt.sql)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v want %v", err, tt.wantErr)
			}
		})
	}
}
