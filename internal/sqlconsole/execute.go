package sqlconsole

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"zion-english/internal/constants"
)

type Options struct {
	MaxRows int
	Timeout time.Duration
}

func DefaultOptions() Options {
	return Options{
		MaxRows: constants.SQLConsoleDefaultRowLimit,
		Timeout: time.Duration(constants.SQLConsoleQueryTimeoutSec) * time.Second,
	}
}

type Stats struct {
	DurationMs   int64
	ColumnCount  int
	RowsReturned int
	RowLimit     int
	Truncated    bool
	QueryLength  int
}

type Result struct {
	Columns []string
	Rows    [][]string
	Stats   Stats
}

func Execute(ctx context.Context, db *sql.DB, query string, opts Options) (Result, error) {
	if opts.MaxRows <= 0 {
		opts.MaxRows = constants.SQLConsoleDefaultRowLimit
	}
	if opts.Timeout <= 0 {
		opts.Timeout = time.Duration(constants.SQLConsoleQueryTimeoutSec) * time.Second
	}

	start := time.Now()
	stats := Stats{
		RowLimit:    opts.MaxRows,
		QueryLength: len(query),
	}

	runCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	conn, err := db.Conn(runCtx)
	if err != nil {
		return Result{Stats: stats}, fmt.Errorf("%w: %v", ErrSQLExecution, err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(runCtx, "PRAGMA query_only=ON"); err != nil {
		return Result{Stats: stats}, fmt.Errorf("%w: %v", ErrSQLExecution, err)
	}

	rows, err := conn.QueryContext(runCtx, query)
	if err != nil {
		stats.DurationMs = time.Since(start).Milliseconds()
		return Result{Stats: stats}, fmt.Errorf("%w: %v", ErrSQLExecution, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		stats.DurationMs = time.Since(start).Milliseconds()
		return Result{Stats: stats}, fmt.Errorf("%w: %v", ErrSQLExecution, err)
	}
	stats.ColumnCount = len(columns)

	var data [][]string
	for rows.Next() {
		if len(data) >= opts.MaxRows {
			stats.Truncated = true
			continue
		}
		raw := make([]any, len(columns))
		ptrs := make([]any, len(columns))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			stats.DurationMs = time.Since(start).Milliseconds()
			return Result{Stats: stats}, fmt.Errorf("%w: %v", ErrSQLExecution, err)
		}
		row := make([]string, len(columns))
		for i, v := range raw {
			row[i] = formatCell(v)
		}
		data = append(data, row)
	}
	if err := rows.Err(); err != nil {
		stats.DurationMs = time.Since(start).Milliseconds()
		return Result{Stats: stats}, fmt.Errorf("%w: %v", ErrSQLExecution, err)
	}

	stats.RowsReturned = len(data)
	stats.DurationMs = time.Since(start).Milliseconds()

	return Result{
		Columns: columns,
		Rows:    data,
		Stats:   stats,
	}, nil
}

func formatCell(v interface{}) string {
	if v == nil {
		return constants.DisplayNotSet
	}
	switch val := v.(type) {
	case []byte:
		return string(val)
	case string:
		return val
	default:
		return fmt.Sprint(val)
	}
}
